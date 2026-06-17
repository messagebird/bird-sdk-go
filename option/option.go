// Package option carries the functional options that configure a bird.Client at
// construction and override settings for a single call.
//
//	client, _ := bird.NewClient(option.WithAPIKey(key))
//	msg, _ := client.Email.Send(ctx, params, option.WithMaxRetries(0))
package option

import (
	"fmt"
	"net/http"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
)

// RequestOption configures a request. Options are applied in order, so a later
// option wins over an earlier one. Most options work at construction and per
// call; the client-only options (WithAPIKey, WithBaseURL, WithRegion,
// WithHTTPClient) return an error if applied to a single call.
type RequestOption func(*requestconfig.Config) error

// clientOnly rejects an option that only makes sense at client construction.
func clientOnly(name string, apply func(*requestconfig.Config)) RequestOption {
	return func(c *requestconfig.Config) error {
		if c.Sealed {
			return fmt.Errorf("option.%s is a client option; pass it to bird.NewClient, not per call", name)
		}
		apply(c)
		return nil
	}
}

// WithAPIKey sets the API key used for Bearer authentication. The key's region
// prefix (bk_{region}_…) also selects the base URL unless WithBaseURL is set.
func WithAPIKey(key string) RequestOption {
	return clientOnly("WithAPIKey", func(c *requestconfig.Config) { c.APIKey = key })
}

// WithBaseURL overrides the base URL, bypassing region inference. Use it for a
// local or self-hosted server. Takes effect at client construction.
func WithBaseURL(url string) RequestOption {
	return clientOnly("WithBaseURL", func(c *requestconfig.Config) { c.BaseURL = url })
}

// WithHTTPClient supplies the underlying *http.Client. Per-attempt timeouts are
// applied through the request context, so leave the client's own Timeout unset.
func WithHTTPClient(client *http.Client) RequestOption {
	return clientOnly("WithHTTPClient", func(c *requestconfig.Config) { c.HTTPClient = client })
}

// WithRegion selects the region (e.g. "eu1") for base-URL resolution, overriding
// the API key's prefix. WithBaseURL takes precedence over both.
func WithRegion(region string) RequestOption {
	return clientOnly("WithRegion", func(c *requestconfig.Config) { c.Region = region })
}

// WithMaxRetries caps automatic retries of transient failures (timeouts, 429,
// 5xx). Zero disables retries.
func WithMaxRetries(n int) RequestOption {
	return func(c *requestconfig.Config) error { c.MaxRetries = n; return nil }
}

// WithTimeout sets the per-attempt timeout. Each retry gets a fresh budget.
func WithTimeout(d time.Duration) RequestOption {
	return func(c *requestconfig.Config) error { c.Timeout = d; return nil }
}

// WithAPIVersion pins the API version, sent as the X-Bird-API-Version header.
// Off by default until the server honors it.
func WithAPIVersion(version string) RequestOption {
	return func(c *requestconfig.Config) error { c.APIVersion = version; return nil }
}

// WithWebhookSecret sets the signing secret used by Client.Webhooks.Unwrap.
func WithWebhookSecret(secret string) RequestOption {
	return func(c *requestconfig.Config) error { c.WebhookSecret = secret; return nil }
}

// WithEmailDefaults sets values applied to every email send when the per-send
// params leave the corresponding field unset (the per-send value always wins).
func WithEmailDefaults(d requestconfig.EmailDefaults) RequestOption {
	return func(c *requestconfig.Config) error { c.EmailDefaults = d; return nil }
}

// WithIdempotencyKey sets the idempotency key for a mutating call. When unset,
// one is generated and reused across that call's retries.
func WithIdempotencyKey(key string) RequestOption {
	return func(c *requestconfig.Config) error { c.IdempotencyKey = key; return nil }
}

// WithHeader sets an extra request header, replacing any prior value for the key.
func WithHeader(key, value string) RequestOption {
	return func(c *requestconfig.Config) error {
		if c.Header == nil {
			c.Header = http.Header{}
		}
		c.Header.Set(key, value)
		return nil
	}
}

// WithResponseInto captures the transport metadata (status, headers, request
// ID) of a single call into r, without changing the method's return signature.
func WithResponseInto(r *requestconfig.Response) RequestOption {
	return func(c *requestconfig.Config) error { c.ResponseInto = r; return nil }
}
