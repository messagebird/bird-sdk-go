package bird

import (
	"context"
	"iter"
	"net/http"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// WhatsAppService sends WhatsApp template messages and reads them back. Reach
// it via Client.Whatsapp.
type WhatsAppService struct{ client *Client }

// WhatsappSendParams is a single WhatsApp message send. Templates are
// currently the only supported content type, so Template is required; free-form
// content will be added in a future release. Zero-value fields are omitted from
// the request.
type WhatsappSendParams struct {
	To         string                            // required; recipient phone number in E.164 format
	Template   string                            // required; the template's stable handle (e.g. bird_otp)
	Language   string                            // template language variant; omit when the template has a single language
	Components []WhatsAppMessageTemplateComponent // values that fill the template's placeholders
}

func (p WhatsappSendParams) toWire() oapi.SendWhatsAppMessageRequest {
	body := oapi.SendWhatsAppMessageRequest{To: p.To}
	if p.Template != "" || p.Language != "" || len(p.Components) > 0 {
		tmpl := oapi.SendWhatsAppMessageTemplate{Name: p.Template}
		if p.Language != "" {
			language := p.Language
			tmpl.Language = &language
		}
		if len(p.Components) > 0 {
			components := p.Components
			tmpl.Components = &components
		}
		body.Template = &tmpl
	}
	return body
}

// Send sends one WhatsApp template message. Retried safely: a single
// idempotency key is reused across attempts. Provide your own key with
// option.WithIdempotencyKey.
func (s *WhatsAppService) Send(ctx context.Context, params WhatsappSendParams, opts ...option.RequestOption) (*WhatsAppMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.SendWhatsAppMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.SendWhatsAppMessage(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out WhatsAppMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single WhatsApp message with its current delivery status and
// failure detail if it failed.
func (s *WhatsAppService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*WhatsAppMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetWhatsAppMessage(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out WhatsAppMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WhatsappListParams filters the message list. Zero-value fields are omitted.
type WhatsappListParams struct {
	Statuses      []WhatsAppMessageStatus // filter by any of several delivery statuses
	PhoneNumber   string                  // filter by contact phone number (E.164 exact match)
	Bsuid         string                  // filter by business-scoped user ID (Meta identifier)
	CreatedAfter  time.Time
	CreatedBefore time.Time
	Limit         int
}

func (p WhatsappListParams) toWire(startingAfter string) *oapi.ListWhatsAppMessagesParams {
	w := &oapi.ListWhatsAppMessagesParams{}
	if len(p.Statuses) > 0 {
		statuses := p.Statuses
		w.Status = &statuses
	}
	if p.PhoneNumber != "" {
		phoneNumber := p.PhoneNumber
		w.PhoneNumber = &phoneNumber
	}
	if p.Bsuid != "" {
		bsuid := p.Bsuid
		w.Bsuid = &bsuid
	}
	if !p.CreatedAfter.IsZero() {
		after := oapi.CreatedAfter(p.CreatedAfter)
		w.CreatedAfter = &after
	}
	if !p.CreatedBefore.IsZero() {
		before := oapi.CreatedBefore(p.CreatedBefore)
		w.CreatedBefore = &before
	}
	if p.Limit > 0 {
		limit := oapi.PaginationLimit(p.Limit)
		w.Limit = &limit
	}
	if startingAfter != "" {
		cursor := oapi.StartingAfter(startingAfter)
		w.StartingAfter = &cursor
	}
	return w
}

// ListPage fetches one page of messages. Pass the previous page's NextCursor as
// startingAfter to advance; "" starts from the most recent.
func (s *WhatsAppService) ListPage(ctx context.Context, params WhatsappListParams, startingAfter string, opts ...option.RequestOption) (*WhatsAppMessageList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListWhatsAppMessages(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out WhatsAppMessageList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every message matching params, fetching pages lazily. Range over
// it; the second value is non-nil only on the iteration where a fetch failed.
func (s *WhatsAppService) List(ctx context.Context, params WhatsappListParams, opts ...option.RequestOption) iter.Seq2[*WhatsAppMessage, error] {
	return paginate(func(cursor string) ([]WhatsAppMessage, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}

// WhatsappListEventsParams filters a message's event timeline. Zero-value
// fields are omitted.
type WhatsappListEventsParams struct {
	Type string // filter by event type (e.g. "whatsapp.delivered"); empty returns every event
}

func (p WhatsappListEventsParams) toWire() *oapi.ListWhatsAppMessageEventsParams {
	w := &oapi.ListWhatsAppMessageEventsParams{}
	if p.Type != "" {
		eventType := p.Type
		w.Type = &eventType
	}
	return w
}

// ListEvents returns the lifecycle event timeline for a WhatsApp message, in
// chronological order. The timeline is bounded and returned in full — this
// list is not paginated.
func (s *WhatsAppService) ListEvents(ctx context.Context, id string, params WhatsappListEventsParams, opts ...option.RequestOption) (*WhatsAppEventList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListWhatsAppMessageEvents(ctx, id, params.toWire(), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out WhatsAppEventList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
