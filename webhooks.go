package bird

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// webhookTolerance is how far a webhook timestamp may drift from now before the
// payload is rejected as stale or replayed.
const webhookTolerance = 5 * time.Minute

// WebhooksService verifies inbound webhook deliveries and creates endpoints.
// Reach it via Client.Webhooks. Unwrap is pure crypto and needs only the signing
// secret, set with option.WithWebhookSecret on the client (or per call); Create
// calls the API, so it needs the client's API key like any other request.
type WebhooksService struct{ resource }

// Event is a verified webhook event. Switch on Type, or call AsAny and
// type-switch on the concrete payload (e.g. EmailDeliveredEvent).
type Event struct{ raw oapi.WebhookEvent }

// Type returns the event's discriminant, e.g. EventTypeEmailDelivered.
func (e Event) Type() WebhookEventType {
	t, _ := e.raw.Discriminator()
	return WebhookEventType(t)
}

// AsAny decodes the event into its concrete payload type. An unknown future
// event type returns an error rather than a panic, so an older SDK keeps
// working against a newer server.
func (e Event) AsAny() (any, error) {
	return e.raw.ValueByDiscriminator()
}

// Unwrap verifies the Standard Webhooks signature over the raw request body and
// returns the decoded event. Hand it the exact bytes received — parsing and
// re-serializing before verifying breaks the signature.
// WebhookEventType is an open string: a type added by a newer server arrives
// here as a plain string rather than one of the EventType* constants.
func (s *WebhooksService) Unwrap(payload []byte, headers http.Header, opts ...option.RequestOption) (Event, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return Event{}, err
	}
	if cfg.WebhookSecret == "" {
		return Event{}, &WebhookVerificationError{Reason: "no signing secret configured; pass option.WithWebhookSecret"}
	}
	if err := verifySignature(cfg.WebhookSecret, payload, headers); err != nil {
		return Event{}, err
	}
	var event oapi.WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return Event{}, &WebhookVerificationError{Reason: "payload is not valid JSON"}
	}
	return Event{raw: event}, nil
}

// verifySignature checks a Standard Webhooks signature: HMAC-SHA256 over
// "{id}.{timestamp}.{body}" keyed by the base64 secret, compared in constant
// time against any of the space-delimited "v1,<sig>" entries, with the
// timestamp held within the tolerance window.
func verifySignature(secret string, payload []byte, headers http.Header) error {
	id := headers.Get("webhook-id")
	timestamp := headers.Get("webhook-timestamp")
	signatures := headers.Get("webhook-signature")
	if id == "" || timestamp == "" || signatures == "" {
		return &WebhookVerificationError{Reason: "missing webhook-id, webhook-timestamp, or webhook-signature header"}
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return &WebhookVerificationError{Reason: "invalid webhook-timestamp"}
	}
	if drift := time.Since(time.Unix(ts, 0)); drift > webhookTolerance || drift < -webhookTolerance {
		return &WebhookVerificationError{Reason: "timestamp outside the tolerance window"}
	}

	key := strings.TrimPrefix(secret, "whsec_")
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return &WebhookVerificationError{Reason: "malformed signing secret"}
	}
	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(id + "." + timestamp + "."))
	mac.Write(payload)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	for _, entry := range strings.Split(signatures, " ") {
		version, sig, ok := strings.Cut(entry, ",")
		if !ok || version != "v1" {
			continue
		}
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return nil
		}
	}
	return &WebhookVerificationError{Reason: "no matching signature"}
}

// WebhooksCreateParams is the request body for Create. It is hand-written
// because the generated params cannot carry Destination, a union of the raw
// event and a connector.
type WebhooksCreateParams struct {
	// HTTPS URL to deliver events to, at most 2048 characters, on a publicly
	// reachable host. Leave it empty with a connector Destination: Bird builds
	// the URL from the connector and its config, and a URL given anyway must
	// equal that.
	URL string
	// Event types to subscribe to; the endpoint receives only matching events.
	// Types outside the event catalog return a 422, and an endpoint holds at
	// most 100 entries.
	Events []WebhookEventType
	// Human-readable label for this endpoint, up to 256 characters.
	Description *string
	// Filter limits delivery to one mailbox; Events must contain only email_mailbox events.
	Filter *WebhookFilter
	// Destination delivers through a connector instead of posting the raw
	// event: build it with FromWebhookConnectorDestinationCreate. Its
	// credentials are write-only.
	Destination *WebhookDestinationCreate
}

func (p WebhooksCreateParams) toWire() oapi.WebhookEndpointCreate {
	body := oapi.WebhookEndpointCreate{
		Events:      make([]oapi.WebhookEventType, len(p.Events)),
		Description: p.Description,
		Filter:      p.Filter,
		Destination: p.Destination,
	}
	for i, v := range p.Events {
		body.Events[i] = oapi.WebhookEventType(v)
	}
	if p.URL != "" {
		body.Url = &p.URL
	}
	return body
}

// Create registers an endpoint to receive this workspace's events, subscribed
// to the event types in Events and active immediately. The response is the
// only place the signing secret appears; it can never be read back, only
// rotated.
func (s *WebhooksService) Create(ctx context.Context, params WebhooksCreateParams, opts ...option.RequestOption) (*WebhookEndpointCreated, error) {
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreateWebhookParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateWebhook(ctx, op, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out WebhookEndpointCreated
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
