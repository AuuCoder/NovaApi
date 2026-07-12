package xai

import "net/http"

const (
	GrokCLIClientVersion    = "0.2.93"
	GrokCLIClientIdentifier = "grok-pager"
	GrokCLITokenAuth        = "xai-grok-cli"
	GrokCLIUserAgent        = "grok-pager/" + GrokCLIClientVersion + " grok-shell/" + GrokCLIClientVersion + " (linux; x86_64)"
)

// SetGrokCLIHeaders applies the client identity required by the Grok CLI upstream.
func SetGrokCLIHeaders(header http.Header) {
	header.Set("User-Agent", GrokCLIUserAgent)
	header.Set("X-XAI-Token-Auth", GrokCLITokenAuth)
	header.Set("x-grok-client-identifier", GrokCLIClientIdentifier)
	header.Set("x-grok-client-version", GrokCLIClientVersion)
}
