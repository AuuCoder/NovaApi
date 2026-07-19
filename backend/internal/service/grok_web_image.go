package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const grokWebImageMaxBridgeBodyBytes int64 = 128 << 20

var grokImageCleanupLastUnix atomic.Int64

type grokWebBridgeRequest struct {
	SSOToken string `json:"sso_token"`
	Prompt   string `json:"prompt"`
	N        int    `json:"n"`
	ProxyURL string `json:"proxy_url,omitempty"`
}

type grokWebBridgeImage struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
}

type grokWebBridgeResponse struct {
	Images []grokWebBridgeImage `json:"images"`
}

func (s *OpenAIGatewayService) ForwardGrokWebImage(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	requestInfo GrokMediaRequestInfo,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()
	if s == nil || s.cfg == nil {
		return nil, fmt.Errorf("grok web image runtime is not configured")
	}
	if account == nil || account.GetGrokSSOToken() == "" {
		return nil, &UpstreamFailoverError{StatusCode: http.StatusUnauthorized}
	}
	if strings.TrimSpace(requestInfo.Prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if requestInfo.N < 1 || requestInfo.N > 4 {
		return nil, fmt.Errorf("n must be between 1 and 4 for %s", requestInfo.Model)
	}
	if requestInfo.ResponseFormat != "url" && requestInfo.ResponseFormat != "b64_json" {
		return nil, fmt.Errorf("response_format must be url or b64_json")
	}

	bridgeURL := strings.TrimSpace(s.cfg.Gateway.GrokWebBridgeURL)
	bridgeKey := strings.TrimSpace(s.cfg.Gateway.GrokWebBridgeKey)
	if bridgeURL == "" || bridgeKey == "" {
		return nil, fmt.Errorf("grok web image runtime is disabled")
	}
	proxyURL := strings.TrimSpace(s.cfg.Gateway.GrokWebDefaultProxyURL)
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	payload, err := json.Marshal(grokWebBridgeRequest{
		SSOToken: account.GetGrokSSOToken(),
		Prompt:   requestInfo.Prompt,
		N:        requestInfo.N,
		ProxyURL: proxyURL,
	})
	if err != nil {
		return nil, err
	}

	upstreamCtx, releaseUpstreamCtx := detachUpstreamContext(ctx)
	defer releaseUpstreamCtx()
	req, err := http.NewRequestWithContext(upstreamCtx, http.MethodPost, bridgeURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Internal-Key", bridgeKey)

	upstreamStart := time.Now()
	resp, err := (&http.Client{Timeout: 10 * time.Minute}).Do(req)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if err != nil {
		return nil, &UpstreamFailoverError{StatusCode: http.StatusBadGateway, ResponseBody: []byte(sanitizeUpstreamErrorMessage(err.Error()))}
	}
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, grokWebImageMaxBridgeBodyBytes+1))
	if readErr != nil {
		return nil, readErr
	}
	if int64(len(body)) > grokWebImageMaxBridgeBodyBytes {
		return nil, fmt.Errorf("grok web image bridge response too large")
	}
	if resp.StatusCode >= http.StatusBadRequest {
		setOpsUpstreamError(c, resp.StatusCode, "grok web image generation failed", "")
		return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: body}
	}

	var bridgeResponse grokWebBridgeResponse
	if err := json.Unmarshal(body, &bridgeResponse); err != nil {
		return nil, fmt.Errorf("decode grok web image response: %w", err)
	}
	if len(bridgeResponse.Images) == 0 {
		return nil, fmt.Errorf("grok web image runtime returned no images")
	}

	data := make([]map[string]string, 0, len(bridgeResponse.Images))
	outputSizes := make([]string, 0, len(bridgeResponse.Images))
	for _, image := range bridgeResponse.Images {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(image.Data))
		if err != nil || len(raw) == 0 {
			return nil, fmt.Errorf("decode grok web image payload")
		}
		if requestInfo.ResponseFormat == "b64_json" {
			data = append(data, map[string]string{"b64_json": base64.StdEncoding.EncodeToString(raw)})
			continue
		}
		name, err := s.saveGrokImage(raw, image.MIMEType)
		if err != nil {
			return nil, err
		}
		data = append(data, map[string]string{"url": grokImagePublicURL(c, name)})
	}

	responseBody, err := json.Marshal(map[string]any{
		"created": time.Now().Unix(),
		"data":    data,
	})
	if err != nil {
		return nil, err
	}
	c.Data(http.StatusOK, "application/json", responseBody)

	return &OpenAIForwardResult{
		RequestID:        firstNonEmpty(resp.Header.Get("x-request-id"), uuid.NewString()),
		Model:            requestInfo.Model,
		BillingModel:     requestInfo.Model,
		UpstreamModel:    requestInfo.Model,
		ResponseHeaders:  resp.Header.Clone(),
		Duration:         time.Since(startTime),
		ImageCount:       len(data),
		ImageSize:        requestInfo.SizeTier,
		ImageInputSize:   requestInfo.Size,
		ImageOutputSizes: outputSizes,
	}, nil
}

func (s *OpenAIGatewayService) saveGrokImage(raw []byte, mimeType string) (string, error) {
	cacheDir := strings.TrimSpace(s.cfg.Gateway.GrokImageCacheDir)
	if cacheDir == "" {
		return "", fmt.Errorf("grok image cache directory is not configured")
	}
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return "", fmt.Errorf("create grok image cache: %w", err)
	}
	extension := extensionForImageMIME(mimeType)
	name := uuid.NewString() + extension
	path := filepath.Join(cacheDir, name)
	if err := os.WriteFile(path, raw, 0o640); err != nil {
		return "", fmt.Errorf("save grok image: %w", err)
	}
	s.cleanupExpiredGrokImages(cacheDir)
	return name, nil
}

func extensionForImageMIME(mimeType string) string {
	mediaType, _, _ := mime.ParseMediaType(strings.TrimSpace(mimeType))
	switch strings.ToLower(mediaType) {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

func (s *OpenAIGatewayService) cleanupExpiredGrokImages(cacheDir string) {
	now := time.Now()
	last := grokImageCleanupLastUnix.Load()
	if last > 0 && now.Unix()-last < int64(time.Hour/time.Second) {
		return
	}
	if !grokImageCleanupLastUnix.CompareAndSwap(last, now.Unix()) {
		return
	}
	retention := time.Duration(s.cfg.Gateway.GrokImageRetentionHours) * time.Hour
	if retention <= 0 {
		retention = 24 * time.Hour
	}
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}
	cutoff := now.Add(-retention)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(cacheDir, entry.Name()))
		}
	}
}

func grokImagePublicURL(c *gin.Context, name string) string {
	scheme := "http"
	host := ""
	if c != nil && c.Request != nil {
		host = strings.TrimSpace(c.Request.Host)
		if forwarded := strings.TrimSpace(strings.Split(c.GetHeader("X-Forwarded-Proto"), ",")[0]); forwarded != "" {
			scheme = forwarded
		} else if c.Request.TLS != nil {
			scheme = "https"
		}
	}
	return fmt.Sprintf("%s://%s/v1/files/grok-image/%s", scheme, host, name)
}

func ResolveGrokImageFile(cfgDir, name string) (string, string, bool) {
	name = filepath.Base(strings.TrimSpace(name))
	extension := strings.ToLower(filepath.Ext(name))
	id := strings.TrimSuffix(name, extension)
	if _, err := uuid.Parse(id); err != nil {
		return "", "", false
	}
	mimeType := ""
	switch extension {
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	case ".gif":
		mimeType = "image/gif"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	default:
		return "", "", false
	}
	cacheDir := strings.TrimSpace(cfgDir)
	if cacheDir == "" {
		return "", "", false
	}
	path := filepath.Join(cacheDir, name)
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return "", "", false
	}
	return path, mimeType, true
}
