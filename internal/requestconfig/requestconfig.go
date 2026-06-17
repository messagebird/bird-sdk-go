// Package requestconfig holds the resolved per-request configuration. It lives
// in internal/ so the option package can mutate it and the bird package can
// read it without an import cycle between them.
package requestconfig

import (
	"net/http"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
)

// Config is the settings for a single API call. Client options seed it at
// construction; per-call options Clone it and override individual fields.
//
// BaseURL, APIKey, Region, and HTTPClient take effect at client construction
// only — the underlying HTTP client is built once. The remaining fields are
// honored per call.
type Config struct {
	BaseURL        string
	APIKey         string
	Region         string
	APIVersion     string
	HTTPClient     *http.Client
	MaxRetries     int
	Timeout        time.Duration
	IdempotencyKey string
	WebhookSecret  string
	Header         http.Header
	ResponseInto   *Response
	EmailDefaults  EmailDefaults
	// Sealed is set on the per-call config so construction-only options
	// (WithBaseURL/WithAPIKey/WithHTTPClient) can reject being applied per call.
	Sealed bool
}

// EmailDefaults are values applied to an email send when the per-send params
// leave the corresponding field unset. Re-exported as bird.EmailDefaults.
type EmailDefaults struct {
	From        string
	ReplyTo     []string
	Category    oapi.EmailMessageCategory
	TrackOpens  *bool
	TrackClicks *bool
	Headers     map[string]string
	Tags        []oapi.EmailTag
	Metadata    map[string]interface{}
}

// Response is the transport metadata for one call, populated when the caller
// passes option.WithResponseInto. Re-exported as bird.Response.
type Response struct {
	Status    int
	Header    http.Header
	RequestID string
}

// Clone returns a copy safe to mutate with per-call options. The Header map is
// deep-copied; ResponseInto stays a shared pointer so the call can write into
// the caller's Response.
func (c Config) Clone() Config {
	cp := c
	if c.Header != nil {
		cp.Header = c.Header.Clone()
	}
	return cp
}
