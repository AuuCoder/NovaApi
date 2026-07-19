//go:build unit

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetGrokSSOTokenNormalizesCookieValue(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			"sso_token": " sso=token-value; Path=/; Secure ",
		},
	}

	require.Equal(t, "token-value", account.GetGrokSSOToken())
}

func TestGrokImageLiteModelRequiresSSOToken(t *testing.T) {
	withoutSSO := &Account{Platform: PlatformGrok, Credentials: map[string]any{}}
	withSSO := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			"sso_token": "token-value",
		},
	}

	require.False(t, withoutSSO.IsModelSupported(GrokImagineImageLiteModel))
	require.True(t, withSSO.IsModelSupported(GrokImagineImageLiteModel))
	require.True(t, withoutSSO.IsModelSupported(GrokImagineImageModel))
	require.True(t, withSSO.IsModelSupported(GrokImagineImageModel))
	require.True(t, withoutSSO.IsModelSupported("grok-4.5"))
}

func TestGrokImageLiteModelWithSSOIgnoresAccountModelMapping(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			"sso_token": "token-value",
			"model_mapping": map[string]any{
				"grok-4.5": "grok-4.5",
			},
		},
	}

	require.True(t, account.IsModelSupported(GrokImagineImageLiteModel))
}

func TestGrokImageLiteModelSupportsSlimSchedulerMetadata(t *testing.T) {
	withSSOCapability := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			GrokSSOAvailableCredentialKey: true,
		},
	}
	withoutSSOCapability := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			GrokSSOAvailableCredentialKey: false,
		},
	}

	require.True(t, withSSOCapability.IsModelSupported(GrokImagineImageLiteModel))
	require.False(t, withoutSSOCapability.IsModelSupported(GrokImagineImageLiteModel))
}

func TestShouldUseGrokWebImagePreservesOfficialFallback(t *testing.T) {
	withSSO := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			"sso_token": "token-value",
		},
	}
	withoutSSO := &Account{Platform: PlatformGrok, Credentials: map[string]any{}}

	require.True(t, shouldUseGrokWebImage(withSSO, GrokImagineImageLiteModel))
	require.True(t, shouldUseGrokWebImage(withoutSSO, GrokImagineImageLiteModel))
	require.True(t, shouldUseGrokWebImage(withSSO, GrokImagineImageModel))
	require.False(t, shouldUseGrokWebImage(withoutSSO, GrokImagineImageModel))
	require.False(t, shouldUseGrokWebImage(withSSO, "grok-imagine-image-quality"))
}

func TestForwardGrokMediaRoutesImagineImageSSOAccountToWebBridge(t *testing.T) {
	imageBytes := []byte("generated-image")
	bridge := newGrokWebImageBridgeTestServer(t, imageBytes)
	defer bridge.Close()

	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{
		GrokWebBridgeURL:  bridge.URL,
		GrokWebBridgeKey:  "bridge-secret",
		GrokImageCacheDir: t.TempDir(),
	}}}
	c, recorder := newGrokWebImageTestContext()
	account := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			"sso_token": "sso-token-value",
		},
	}
	body := []byte(`{"model":"grok-imagine-image","prompt":"draw a cat","n":1}`)

	result, err := svc.ForwardGrokMedia(
		context.Background(),
		c,
		account,
		GrokMediaEndpointImagesGenerations,
		"",
		body,
		"application/json",
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, GrokImagineImageModel, result.Model)
}

func TestForwardGrokWebImageReturnsPublicURLAndCachesImage(t *testing.T) {
	imageBytes := []byte("generated-image")
	bridge := newGrokWebImageBridgeTestServer(t, imageBytes)
	defer bridge.Close()

	cacheDir := t.TempDir()
	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{
		GrokWebBridgeURL:        bridge.URL,
		GrokWebBridgeKey:        "bridge-secret",
		GrokWebDefaultProxyURL:  "http://grok-privoxy:8118",
		GrokImageCacheDir:       cacheDir,
		GrokImageRetentionHours: 24,
	}}}
	c, recorder := newGrokWebImageTestContext()
	account := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			"sso_token": "sso-token-value",
		},
	}

	result, err := svc.ForwardGrokWebImage(context.Background(), c, account, GrokMediaRequestInfo{
		Model:          GrokImagineImageLiteModel,
		Prompt:         "draw a cat",
		N:              1,
		ResponseFormat: "url",
		Size:           "1024x1024",
		SizeTier:       "1K",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, GrokImagineImageLiteModel, result.Model)

	var response struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	require.True(t, strings.HasPrefix(response.Data[0].URL, "https://api.example.test/v1/files/grok-image/"))

	name := filepath.Base(response.Data[0].URL)
	path, mimeType, ok := ResolveGrokImageFile(cacheDir, name)
	require.True(t, ok)
	require.Equal(t, "image/png", mimeType)
	cached, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, imageBytes, cached)
}

func TestForwardGrokWebImageReturnsBase64WithoutCaching(t *testing.T) {
	imageBytes := []byte("generated-image")
	bridge := newGrokWebImageBridgeTestServer(t, imageBytes)
	defer bridge.Close()

	cacheDir := t.TempDir()
	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{
		GrokWebBridgeURL:  bridge.URL,
		GrokWebBridgeKey:  "bridge-secret",
		GrokImageCacheDir: cacheDir,
	}}}
	c, recorder := newGrokWebImageTestContext()
	account := &Account{
		Platform: PlatformGrok,
		Credentials: map[string]any{
			"sso_token": "sso-token-value",
		},
	}

	_, err := svc.ForwardGrokWebImage(context.Background(), c, account, GrokMediaRequestInfo{
		Model:          GrokImagineImageLiteModel,
		Prompt:         "draw a cat",
		N:              1,
		ResponseFormat: "b64_json",
	})
	require.NoError(t, err)

	var response struct {
		Data []struct {
			Base64 string `json:"b64_json"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	require.Equal(t, base64.StdEncoding.EncodeToString(imageBytes), response.Data[0].Base64)
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestResolveGrokImageFileValidatesNameAndExtension(t *testing.T) {
	cacheDir := t.TempDir()
	name := "c9f832f6-b219-4f85-b7ad-185af4937961.webp"
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, name), []byte("webp"), 0o600))

	path, mimeType, ok := ResolveGrokImageFile(cacheDir, name)
	require.True(t, ok)
	require.Equal(t, filepath.Join(cacheDir, name), path)
	require.Equal(t, "image/webp", mimeType)

	_, _, ok = ResolveGrokImageFile(cacheDir, "not-a-uuid.webp")
	require.False(t, ok)
	_, _, ok = ResolveGrokImageFile(cacheDir, "c9f832f6-b219-4f85-b7ad-185af4937961.txt")
	require.False(t, ok)
}

func newGrokWebImageBridgeTestServer(t *testing.T, imageBytes []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "bridge-secret", r.Header.Get("X-Internal-Key"))
		var body grokWebBridgeRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "sso-token-value", body.SSOToken)
		require.Equal(t, "draw a cat", body.Prompt)
		require.Equal(t, 1, body.N)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(grokWebBridgeResponse{Images: []grokWebBridgeImage{
			{MIMEType: "image/png", Data: base64.StdEncoding.EncodeToString(imageBytes)},
		}}))
	}))
}

func newGrokWebImageTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "http://api.example.test/v1/images/generations", nil)
	c.Request.Host = "api.example.test"
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	return c, recorder
}
