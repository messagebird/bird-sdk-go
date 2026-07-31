package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// WhatsappService sends WhatsApp template messages and reads them back. Reach it
// via Client.Whatsapp.
type WhatsappService struct{ resource }

// WhatsappSendParams is a single WhatsApp message send. Templates are
// currently the only supported content type, so Template is required; free-form
// content will be added in a future release. Zero-value fields are omitted from
// the request.
type WhatsappSendParams struct {
	To         string                             // required; recipient phone number in E.164 format
	Template   string                             // required; the template's stable handle (e.g. bird_otp)
	Language   string                             // template language variant; omit when the template has a single language
	Components []WhatsAppMessageTemplateComponent // values that fill the template's placeholders
}

func (p WhatsappSendParams) toWire() oapi.WhatsAppMessageSendRequest {
	body := oapi.WhatsAppMessageSendRequest{To: p.To}
	if p.Template != "" || p.Language != "" || len(p.Components) > 0 {
		tmpl := oapi.WhatsAppTemplateSend{Name: p.Template}
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
func (s *WhatsappService) Send(ctx context.Context, params WhatsappSendParams, opts ...option.RequestOption) (*WhatsAppMessage, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.SendWhatsAppMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.SendWhatsAppMessage(ctx, p, wire, cfg...)
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
