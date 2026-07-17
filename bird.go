// Package bird is the Go SDK for the Bird email platform (ADR-0044).
//
// The wire types and a low-level client are generated from the OpenAPI spec
// into internal/oapi and never hand-edited. This package is the hand-written,
// idiomatic layer on top: a curated resource surface, a typed error hierarchy,
// context cancellation, functional options, safe retries with
// reused idempotency keys, range-over-func pagination, and webhook
// verification. The request lifecycle lives in internal/requestconfig and the
// error model in internal/apierror; both are re-exported here.
//
//	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
//	if err != nil { ... }
//	msg, err := client.Email.Send(ctx, bird.EmailSendParams{
//		From: "hello@acme.com", To: []string{"customer@example.com"},
//		Subject: "Welcome", HTML: "<h1>Hi</h1>",
//	})
package bird

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
	"github.com/messagebird/bird-sdk-go/option"
)

const (
	version = "0.8.0"
	// userAgent is human-readable only; the API attributes the SDK from the
	// Bird-* headers set in callEditors (ADR-0074), not the UA.
	userAgent = "bird-sdk-go/" + version
)

const (
	defaultMaxRetries = 2
	defaultTimeout    = 60 * time.Second
)

// Response is the transport metadata for one call, captured via
// option.WithResponseInto.
type Response = requestconfig.Response

// EmailDefaults are values applied to an email send when the per-send params
// leave the field unset. Configure with option.WithEmailDefaults.
type EmailDefaults = requestconfig.EmailDefaults

// Client is the entry point to the SDK. Construct it with NewClient and reach
// the API through its resource fields.
type Client struct {
	cfg  requestconfig.Config
	oapi *oapi.Client

	Email             *EmailService
	Sms               *SMSService
	SmsTemplates      *SMSTemplatesService
	Whatsapp          *WhatsAppService
	WhatsappTemplates *WhatsAppTemplatesService
	Verify            *VerifyService
	Webhooks          *WebhookService
	Contacts          *ContactsService
	Audiences         *AudiencesService
	ContactProperties *ContactPropertiesService
	Domains           *DomainsService
}

// NewClient builds a Client. An API key is required (option.WithAPIKey); the
// base URL is derived from the key's region prefix unless option.WithBaseURL or
// option.WithRegion is given.
func NewClient(opts ...option.RequestOption) (*Client, error) {
	cfg := requestconfig.Config{MaxRetries: defaultMaxRetries, Timeout: defaultTimeout}
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}
	if cfg.APIKey == "" {
		return nil, errors.New("bird: an API key is required; pass option.WithAPIKey")
	}
	if cfg.BaseURL == "" {
		region := cfg.Region
		if region == "" {
			region = regionFromAPIKey(cfg.APIKey)
		}
		if region == "" {
			return nil, errors.New("bird: cannot determine region; pass option.WithRegion or option.WithBaseURL (or use a bk_{region}_{token} key)")
		}
		cfg.BaseURL = baseURLForRegion(region)
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{}
	}

	oc, err := oapi.NewClient(cfg.BaseURL, oapi.WithHTTPClient(cfg.HTTPClient))
	if err != nil {
		return nil, err
	}

	c := &Client{cfg: cfg, oapi: oc}
	c.Email = &EmailService{client: c}
	c.Sms = &SMSService{client: c}
	c.SmsTemplates = &SMSTemplatesService{client: c}
	c.Whatsapp = &WhatsAppService{client: c}
	c.WhatsappTemplates = &WhatsAppTemplatesService{client: c}
	c.Verify = &VerifyService{Verifications: &VerificationsService{client: c}}
	c.Webhooks = &WebhookService{client: c}
	c.Contacts = &ContactsService{client: c}
	c.Audiences = &AudiencesService{client: c}
	c.ContactProperties = &ContactPropertiesService{client: c}
	c.Domains = &DomainsService{client: c}
	return c, nil
}

var regionPattern = regexp.MustCompile(`^[a-z]{2}[0-9]+$`)

// regionFromAPIKey extracts the region from a bk_{region}_{token} key, or "".
func regionFromAPIKey(key string) string {
	parts := strings.SplitN(key, "_", 3)
	if len(parts) < 3 || parts[0] != "bk" || parts[2] == "" {
		return ""
	}
	if !regionPattern.MatchString(parts[1]) {
		return ""
	}
	return parts[1]
}

func baseURLForRegion(region string) string {
	return "https://" + region + ".platform.bird.com"
}

// resolveRawRequestURL validates a caller-supplied escape-hatch path before the
// API key is attached and the request is sent. The verb methods join the path
// onto the base URL, so a crafted path — "//host", "user@host", an absolute
// URL, or a bare-relative segment — could otherwise redirect the bearer token
// to a different host. It requires an absolute path with a single leading slash
// and asserts the resolved URL stays on the configured base-URL origin.
func resolveRawRequestURL(baseURL, path string) (string, error) {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "", fmt.Errorf("bird: request path must be an absolute path starting with a single '/': got %q", path)
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("bird: invalid base URL %q: %w", baseURL, err)
	}
	full, err := url.Parse(baseURL + path)
	if err != nil {
		return "", fmt.Errorf("bird: invalid request path %q: %w", path, err)
	}
	if full.Scheme != base.Scheme || full.Host != base.Host {
		return "", fmt.Errorf("bird: request path %q must stay on the configured Bird API origin %s://%s", path, base.Scheme, base.Host)
	}
	return full.String(), nil
}

// resolve clones the client config and applies per-call options. The clone is
// sealed first, so construction-only options are rejected per call.
func (c *Client) resolve(opts []option.RequestOption) (requestconfig.Config, error) {
	cfg := c.cfg.Clone()
	cfg.Sealed = true
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

// callEditors returns the per-request editors that stamp caller-supplied
// headers and then the SDK-owned headers. SDK-owned headers are applied last so
// a caller's option.WithHeader can never override auth, User-Agent, the API
// version, or the idempotency key.
func (c *Client) callEditors(cfg requestconfig.Config) []oapi.RequestEditorFn {
	return []oapi.RequestEditorFn{func(_ context.Context, req *http.Request) error {
		for k, vs := range cfg.Header {
			if isReservedHeader(k) {
				continue
			}
			req.Header[http.CanonicalHeaderKey(k)] = append([]string(nil), vs...)
		}
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("User-Agent", userAgent)
		// Bird-* client-identity headers (ADR-0074) — telemetry labels only.
		// Header keys are the canonical wire form http.Header.Set normalizes to.
		req.Header.Set("Bird-Surface", "sdk-go")
		req.Header.Set("Bird-Version", version)
		req.Header.Set("Bird-Lang", "go")
		req.Header.Set("Bird-Os", runtime.GOOS)
		req.Header.Set("Bird-Arch", runtime.GOARCH)
		if c := detectCaller(os.Getenv); c != "" {
			req.Header.Set("Bird-Caller", c)
		}
		if cfg.APIVersion != "" {
			req.Header.Set("X-Bird-API-Version", cfg.APIVersion)
		}
		return nil
	}}
}

// isReservedHeader reports whether a header is owned by the SDK and so must not
// be set from a caller's option.WithHeader.
func isReservedHeader(key string) bool {
	switch http.CanonicalHeaderKey(key) {
	case "Authorization", "User-Agent", "X-Bird-Api-Version", "Idempotency-Key",
		"Bird-Surface", "Bird-Version", "Bird-Lang", "Bird-Os", "Bird-Arch", "Bird-Caller":
		return true
	default:
		return false
	}
}

// Get, Post, Put, Patch, and Delete are the escape hatch for endpoints outside
// the curated surface. They run through the same auth, retry, idempotency, and
// base-URL handling as the typed methods. body (if non-nil) is sent as JSON; a
// 2xx response is decoded into out (if non-nil).
//
//	var out SuppressionList
//	err := client.Get(ctx, "/v1/email/suppressions", &out)
func (c *Client) Get(ctx context.Context, path string, out any, opts ...option.RequestOption) error {
	return c.Do(ctx, http.MethodGet, path, nil, out, opts...)
}

func (c *Client) Post(ctx context.Context, path string, body, out any, opts ...option.RequestOption) error {
	return c.Do(ctx, http.MethodPost, path, body, out, opts...)
}

func (c *Client) Put(ctx context.Context, path string, body, out any, opts ...option.RequestOption) error {
	return c.Do(ctx, http.MethodPut, path, body, out, opts...)
}

func (c *Client) Patch(ctx context.Context, path string, body, out any, opts ...option.RequestOption) error {
	return c.Do(ctx, http.MethodPatch, path, body, out, opts...)
}

func (c *Client) Delete(ctx context.Context, path string, out any, opts ...option.RequestOption) error {
	return c.Do(ctx, http.MethodDelete, path, nil, out, opts...)
}

// Do is the low-level call the verb methods build on: it marshals body as JSON
// (when non-nil), runs the request lifecycle, and decodes a 2xx body into out
// (when non-nil).
func (c *Client) Do(ctx context.Context, method, path string, body, out any, opts ...option.RequestOption) error {
	cfg, err := c.resolve(opts)
	if err != nil {
		return err
	}
	reqURL, err := resolveRawRequestURL(cfg.BaseURL, path)
	if err != nil {
		return err
	}
	var bodyBytes []byte
	if body != nil {
		if bodyBytes, err = json.Marshal(body); err != nil {
			return fmt.Errorf("bird: encoding request: %w", err)
		}
	}
	respBody, err := cfg.Execute(ctx, isMutation(method), func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		var reader io.Reader
		if bodyBytes != nil {
			reader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
		if err != nil {
			return nil, err
		}
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}
		for _, edit := range c.callEditors(cfg) {
			if err := edit(ctx, req); err != nil {
				return nil, err
			}
		}
		return cfg.HTTPClient.Do(req)
	})
	if err != nil {
		return err
	}
	return decodeBody(respBody, out)
}

func isMutation(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
