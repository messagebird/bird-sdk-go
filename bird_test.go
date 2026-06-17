package bird_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

// newServer starts a test server that closes itself when the test ends.
func newServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// newClient points a client at the test server with retries off unless a test
// opts back in.
func newClient(t *testing.T, server *httptest.Server, opts ...option.RequestOption) *bird.Client {
	t.Helper()
	base := []option.RequestOption{
		option.WithAPIKey("bk_eu1_test"),
		option.WithBaseURL(server.URL),
		option.WithMaxRetries(0),
	}
	client, err := bird.NewClient(append(base, opts...)...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

const messageJSON = `{
	"id": "em_abc123",
	"from": {"email": "hello@acme.com"},
	"to": [{"email": "customer@example.com"}],
	"subject": "Welcome",
	"category": "transactional",
	"status": "accepted",
	"track_clicks": true,
	"track_opens": true
}`

func TestSendSuccess(t *testing.T) {
	var gotMethod, gotAuth, gotUA, gotIdem, gotBody string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotIdem = r.Header.Get("Idempotency-Key")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("X-Request-Id", "req_1")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, messageJSON)
	})

	client := newClient(t, server)
	var meta bird.Response
	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "hello@acme.com",
		To:      []string{"customer@example.com"},
		Subject: "Welcome",
		HTML:    "<h1>Hi</h1>",
	}, option.WithResponseInto(&meta))
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if msg.Id != "em_abc123" {
		t.Errorf("Id = %q, want em_abc123", msg.Id)
	}
	if msg.Status == nil || *msg.Status != bird.EmailStatusAccepted {
		t.Errorf("Status = %v, want accepted", msg.Status)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotAuth != "Bearer bk_eu1_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotUA != "bird-sdk-go/0.1.0" {
		t.Errorf("User-Agent = %q", gotUA)
	}
	if gotIdem == "" {
		t.Error("Idempotency-Key was not set on a mutation")
	}
	var sent struct {
		From string `json:"from"`
		HTML string `json:"html"`
	}
	if err := json.Unmarshal([]byte(gotBody), &sent); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if sent.From != "hello@acme.com" || sent.HTML != "<h1>Hi</h1>" {
		t.Errorf("request body = %+v", sent)
	}
	if meta.Status != http.StatusCreated || meta.RequestID != "req_1" {
		t.Errorf("response metadata = %+v", meta)
	}
}

func TestIdempotencyKeyReusedAcrossRetries(t *testing.T) {
	var keys []string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		if len(keys) == 1 {
			w.Header().Set("Retry-After", "0") // retry immediately
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, messageJSON)
	})

	client := newClient(t, server, option.WithMaxRetries(2))
	if _, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From: "hello@acme.com", To: []string{"c@example.com"}, Subject: "Hi", Text: "yo",
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("attempts = %d, want 2", len(keys))
	}
	if keys[0] == "" || keys[0] != keys[1] {
		t.Errorf("idempotency key not reused across retries: %q vs %q", keys[0], keys[1])
	}
}

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name   string
		status int
		header map[string]string
		body   string
		check  func(t *testing.T, err error)
	}{
		{
			name:   "validation",
			status: http.StatusUnprocessableEntity,
			body:   `{"error":{"type":"validation_error","code":"E12001","name":"invalid_recipient","message":"bad","doc_url":"https://docs.bird.com/errors/E12001","request_id":"req_body","param":"to","vendor_code":"550","details":[{"param":"to[0]","message":"invalid address"}]}}`,
			check: func(t *testing.T, err error) {
				var ve *bird.ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("want *ValidationError, got %T", err)
				}
				if len(ve.Details) != 1 || ve.Details[0].Param != "to[0]" {
					t.Errorf("details = %+v", ve.Details)
				}
				var api *bird.APIError // base must also match via Unwrap
				if !errors.As(err, &api) {
					t.Fatalf("APIError not reachable from ValidationError")
				}
				// Every field must be pulled from inside the {"error":{...}} envelope.
				if api.Code != "E12001" {
					t.Errorf("Code = %q, want E12001", api.Code)
				}
				if api.Name != "invalid_recipient" {
					t.Errorf("Name = %q, want invalid_recipient", api.Name)
				}
				if api.Message != "bad" {
					t.Errorf("Message = %q, want bad", api.Message)
				}
				if api.DocURL != "https://docs.bird.com/errors/E12001" {
					t.Errorf("DocURL = %q", api.DocURL)
				}
				if api.RequestID != "req_body" {
					t.Errorf("RequestID = %q, want req_body", api.RequestID)
				}
				if api.Param != "to" {
					t.Errorf("Param = %q, want to", api.Param)
				}
				if api.VendorCode != "550" {
					t.Errorf("VendorCode = %q, want 550", api.VendorCode)
				}
			},
		},
		{
			name:   "rate limit",
			status: http.StatusTooManyRequests,
			header: map[string]string{"Retry-After": "30"},
			body:   `{"error":{"type":"rate_limit_error","message":"slow down"}}`,
			check: func(t *testing.T, err error) {
				var rle *bird.RateLimitError
				if !errors.As(err, &rle) {
					t.Fatalf("want *RateLimitError, got %T", err)
				}
				if rle.RetryAfter != 30*time.Second {
					t.Errorf("RetryAfter = %v, want 30s", rle.RetryAfter)
				}
				if rle.Message != "slow down" {
					t.Errorf("Message = %q, want slow down", rle.Message)
				}
			},
		},
		{
			name:   "request id falls back to header when body omits it",
			status: http.StatusNotFound,
			header: map[string]string{"X-Request-Id": "req_header"},
			body:   `{"error":{"type":"not_found_error","code":"E40400","message":"no such email"}}`,
			check: func(t *testing.T, err error) {
				var api *bird.APIError
				if !errors.As(err, &api) {
					t.Fatalf("want *APIError, got %T", err)
				}
				if api.Type != bird.ErrorTypeNotFound {
					t.Errorf("Type = %q, want not_found_error", api.Type)
				}
				if api.Code != "E40400" || api.Message != "no such email" {
					t.Errorf("code/message not read from envelope: %+v", api)
				}
				if api.RequestID != "req_header" {
					t.Errorf("RequestID = %q, want req_header from X-Request-Id", api.RequestID)
				}
			},
		},
		{
			name:   "auth from status when body is not json",
			status: http.StatusUnauthorized,
			body:   `not json`,
			check: func(t *testing.T, err error) {
				var api *bird.APIError
				if !errors.As(err, &api) {
					t.Fatalf("want *APIError, got %T", err)
				}
				if api.Type != bird.ErrorTypeAuth {
					t.Errorf("Type = %q, want auth_error", api.Type)
				}
				if api.Message == "" {
					t.Error("Message should fall back to a status-derived string")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				for k, v := range tc.header {
					w.Header().Set(k, v)
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			})

			client := newClient(t, server)
			_, err := client.Email.Get(context.Background(), "em_x")
			if err == nil {
				t.Fatal("want error, got nil")
			}
			tc.check(t, err)
		})
	}
}

// emailPage renders a one-message list page; nextCursor is the literal JSON for
// "next_cursor" (a quoted cursor or null).
func emailPage(id, subject, nextCursor string) string {
	return `{"data":[{"id":"` + id + `","from":{"email":"a@b.com"},"to":[{"email":"c@d.com"}],"subject":"` + subject + `","category":"transactional","track_clicks":false,"track_opens":false}],"next_cursor":` + nextCursor + `}`
}

func TestListPaginatesAcrossPages(t *testing.T) {
	var cursors []string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		cursors = append(cursors, r.URL.Query().Get("starting_after"))
		if r.URL.Query().Get("starting_after") == "" {
			_, _ = io.WriteString(w, emailPage("em_1", "1", `"cur2"`))
			return
		}
		_, _ = io.WriteString(w, emailPage("em_2", "2", `null`))
	})

	client := newClient(t, server)
	var ids []string
	for msg, err := range client.Email.List(context.Background(), bird.EmailListParams{Status: bird.EmailStatusBounced}) {
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		ids = append(ids, msg.Id)
	}
	if strings.Join(ids, ",") != "em_1,em_2" {
		t.Errorf("ids = %v, want [em_1 em_2]", ids)
	}
	if len(cursors) != 2 || cursors[1] != "cur2" {
		t.Errorf("cursors = %v, want second page to use cur2", cursors)
	}
}

func TestListStopsEarlyOnBreak(t *testing.T) {
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, emailPage("em_1", "1", `"never"`))
	})

	client := newClient(t, server)
	count := 0
	for _, err := range client.Email.List(context.Background(), bird.EmailListParams{}) {
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		count++
		break // must not fetch the (infinite) next page
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestContextCancellationSurfacesVerbatim(t *testing.T) {
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, messageJSON)
	})

	client := newClient(t, server)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Email.Get(ctx, "em_x")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

func TestNewClientRequiresResolvableKey(t *testing.T) {
	if _, err := bird.NewClient(); err == nil {
		t.Error("want error without an API key")
	}
	if _, err := bird.NewClient(option.WithAPIKey("not-a-bird-key")); err == nil {
		t.Error("want error for an unresolvable region")
	}
	if _, err := bird.NewClient(option.WithAPIKey("bk_eu1_token")); err != nil {
		t.Errorf("region inference from key should succeed: %v", err)
	}
}

func TestClientOnlyOptionsRejectedPerCall(t *testing.T) {
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, messageJSON)
	})

	client := newClient(t, server)
	if _, err := client.Email.Get(context.Background(), "em_x", option.WithBaseURL("http://evil")); err == nil {
		t.Error("WithBaseURL applied per call should error")
	}
}

const webhookSecret = "whsec_MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw"

func sign(t *testing.T, id, timestamp string, body []byte) string {
	t.Helper()
	key, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(webhookSecret, "whsec_"))
	if err != nil {
		t.Fatalf("decode secret: %v", err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id + "." + timestamp + "."))
	mac.Write(body)
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestWebhookUnwrap(t *testing.T) {
	client, err := bird.NewClient(option.WithAPIKey("bk_eu1_x"), option.WithWebhookSecret(webhookSecret))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	payload := []byte(`{"type":"email.delivered","email_id":"em_1","recipient":"c@d.com"}`)
	id := "msg_2KWPBgLlAfxdpx2AI54pPJ85f4W"
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	headers := http.Header{}
	headers.Set("webhook-id", id)
	headers.Set("webhook-timestamp", ts)
	validSig := sign(t, id, ts, payload)
	headers.Set("webhook-signature", validSig)

	event, err := client.Webhooks.Unwrap(payload, headers)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if event.Type() != bird.EventTypeEmailDelivered {
		t.Errorf("Type = %q, want email.delivered", event.Type())
	}
	payloadAny, err := event.AsAny()
	if err != nil {
		t.Fatalf("AsAny: %v", err)
	}
	if _, ok := payloadAny.(bird.EmailDeliveredEvent); !ok {
		t.Errorf("AsAny returned %T, want EmailDeliveredEvent", payloadAny)
	}

	// A tampered signature must fail with a typed verification error.
	headers.Set("webhook-signature", "v1,deadbeef")
	_, err = client.Webhooks.Unwrap(payload, headers)
	var verr *bird.WebhookVerificationError
	if !errors.As(err, &verr) {
		t.Errorf("want *WebhookVerificationError, got %v", err)
	}

	// Only a v1-tagged entry counts: the same bytes under a v2 tag are rejected.
	b64 := strings.TrimPrefix(validSig, "v1,")
	headers.Set("webhook-signature", "v2,"+b64)
	if _, err := client.Webhooks.Unwrap(payload, headers); !errors.As(err, &verr) {
		t.Errorf("a v2-tagged signature should be rejected, got %v", err)
	}

	// A mixed header with a junk v2 entry and the valid v1 entry still verifies.
	headers.Set("webhook-signature", "v2,junk "+validSig)
	if _, err := client.Webhooks.Unwrap(payload, headers); err != nil {
		t.Errorf("mixed v2+v1 header should verify: %v", err)
	}
}

func TestSDKHeadersWinOverCallerHeaders(t *testing.T) {
	var gotAuth, gotIdem, gotUA, gotCustom string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotIdem = r.Header.Get("Idempotency-Key")
		gotUA = r.Header.Get("User-Agent")
		gotCustom = r.Header.Get("X-Trace-Id")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, messageJSON)
	})

	client := newClient(t, server)
	_, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From: "a@b.com", To: []string{"c@d.com"}, Subject: "x", Text: "y",
	},
		option.WithHeader("Authorization", "Bearer attacker"),
		option.WithHeader("Idempotency-Key", "caller-key"),
		option.WithHeader("User-Agent", "evil/1.0"),
		option.WithHeader("X-Trace-Id", "trace-123"),
	)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotAuth != "Bearer bk_eu1_test" {
		t.Errorf("caller overrode Authorization: %q", gotAuth)
	}
	if gotIdem == "caller-key" || gotIdem == "" {
		t.Errorf("caller overrode Idempotency-Key: %q", gotIdem)
	}
	if gotUA != "bird-sdk-go/0.1.0" {
		t.Errorf("caller overrode User-Agent: %q", gotUA)
	}
	if gotCustom != "trace-123" {
		t.Errorf("non-reserved custom header should be forwarded: %q", gotCustom)
	}
}

func TestVerbMethodsEscapeHatch(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotBody, gotIdem string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotIdem = r.Header.Get("Idempotency-Key")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(w, `{"ok":true}`)
	})

	client := newClient(t, server)
	ctx := context.Background()

	var out struct {
		OK bool `json:"ok"`
	}
	if err := client.Get(ctx, "/v1/anything", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !out.OK {
		t.Error("Get did not decode the body")
	}
	if gotMethod != http.MethodGet || gotPath != "/v1/anything" {
		t.Errorf("method/path = %q %q", gotMethod, gotPath)
	}
	if gotAuth != "Bearer bk_eu1_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}

	// Post marshals the body and, being a mutation, sends an idempotency key.
	if err := client.Post(ctx, "/v1/anything", map[string]string{"hello": "world"}, &out); err != nil {
		t.Fatalf("Post: %v", err)
	}
	if !strings.Contains(gotBody, `"hello":"world"`) {
		t.Errorf("post body = %q", gotBody)
	}
	if gotIdem == "" {
		t.Error("Post should send a generated idempotency key")
	}
}

func TestVerbMethodRejectsOffOriginPaths(t *testing.T) {
	// A crafted path must be rejected before any request is sent, so the API
	// key can never reach a host other than the configured origin.
	for _, path := range []string{
		"@attacker.example/collect",  // userinfo: host becomes attacker.example
		"//attacker.example/collect", // protocol-relative authority
		"https://attacker.example/x", // absolute URL
		"v1/email/domains",           // bare-relative, no leading slash
	} {
		t.Run(path, func(t *testing.T) {
			var called bool
			server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				called = true
				_, _ = io.WriteString(w, `{}`)
			})
			client := newClient(t, server)
			err := client.Get(context.Background(), path, nil)
			if err == nil {
				t.Fatalf("Get(%q) = nil error, want rejection", path)
			}
			if called {
				t.Errorf("Get(%q) sent a request; the API key may have leaked off-origin", path)
			}
		})
	}
}

func TestVerbMethodAllowsAbsolutePathOnOrigin(t *testing.T) {
	var gotPath, gotAuth string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{}`)
	})
	client := newClient(t, server)
	if err := client.Get(context.Background(), "/v1/email/domains", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/v1/email/domains" {
		t.Errorf("path = %q, want /v1/email/domains", gotPath)
	}
	if gotAuth != "Bearer bk_eu1_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
}

func TestPointerHelpers(t *testing.T) {
	if got := bird.Bool(false); got == nil || *got {
		t.Errorf("Bool(false) = %v", got)
	}
	if got := bird.String("x"); got == nil || *got != "x" {
		t.Errorf("String(x) = %v", got)
	}
	if got := bird.Ptr(42); got == nil || *got != 42 {
		t.Errorf("Ptr(42) = %v", got)
	}
}

func TestEmailChannelDefaults(t *testing.T) {
	var gotFrom, gotCategory string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			From     string `json:"from"`
			Category string `json:"category"`
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		gotFrom, gotCategory = body.From, body.Category
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, messageJSON)
	})

	client := newClient(t, server, option.WithEmailDefaults(bird.EmailDefaults{
		From:     "default@acme.com",
		Category: bird.CategoryTransactional,
	}))
	ctx := context.Background()

	// From + Category omitted → defaults apply.
	if _, err := client.Email.Send(ctx, bird.EmailSendParams{
		To: []string{"c@example.com"}, Subject: "Hi", Text: "y",
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotFrom != "default@acme.com" || gotCategory != "transactional" {
		t.Errorf("defaults not applied: from=%q category=%q", gotFrom, gotCategory)
	}

	// Per-send value wins over the default.
	if _, err := client.Email.Send(ctx, bird.EmailSendParams{
		From: "override@acme.com", To: []string{"c@example.com"}, Subject: "Hi", Text: "y",
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotFrom != "override@acme.com" {
		t.Errorf("per-send From should win: %q", gotFrom)
	}
}

func TestWithRegionResolvesBaseURL(t *testing.T) {
	// A non-bk_ key can't resolve a region on its own; WithRegion supplies it.
	if _, err := bird.NewClient(option.WithAPIKey("rawkey")); err == nil {
		t.Error("want error: region unresolvable from a raw key")
	}
	if _, err := bird.NewClient(option.WithAPIKey("rawkey"), option.WithRegion("eu1")); err != nil {
		t.Errorf("WithRegion should resolve the base URL: %v", err)
	}
}
