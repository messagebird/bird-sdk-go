package bird

import (
	"context"
	"net/http"
	"strings"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// WhatsappService sends WhatsApp messages and reads them back. Reach it
// via Client.Whatsapp.
type WhatsappService struct{ resource }

// WhatsappSendParams is a single WhatsApp message send. Carry exactly one kind
// of content — a template, or one free-form arm. Which arms a send may use, and
// whether From is required for it, are the server's to decide, so this type
// enforces neither. Zero-value fields are omitted from the request.
type WhatsappSendParams struct {
	To   string // required; recipient phone number in E.164 format, or a business-scoped user ID
	From string // the business number to send from; omit only for a Bird-managed template

	Template   string                             // the template's id (wat_…) or its slug (e.g. bird_otp)
	Language   string                             // template language as a BCP-47 tag; omit when the template has a single language
	Components []WhatsAppMessageTemplateComponent // values that fill the template's placeholders

	Text     *WhatsAppTextSend
	Image    *WhatsAppImageSend
	Video    *WhatsAppVideoSend
	Audio    *WhatsAppAudioSend
	Sticker  *WhatsAppStickerSend
	Document *WhatsAppDocumentSend
	Location *WhatsAppLocationSend

	Tags     []WhatsAppTag  // structured {name, value} labels for filtering and analytics
	Metadata map[string]any // arbitrary JSON stored on the message and echoed in webhooks
}

func (p WhatsappSendParams) toWire() oapi.WhatsAppMessageSendRequest {
	body := oapi.WhatsAppMessageSendRequest{To: p.To}
	if p.From != "" {
		from := p.From
		body.From = &from
	}
	if p.Template != "" || p.Language != "" || len(p.Components) > 0 {
		var tmpl oapi.WhatsAppTemplateSend
		if p.Template != "" {
			// A wat_-prefixed value is the id; anything else is the slug handle.
			if strings.HasPrefix(p.Template, "wat_") {
				id := oapi.WhatsAppTemplateID(p.Template)
				tmpl.Id = &id
			} else {
				slug := oapi.TemplateSlug(p.Template)
				tmpl.Slug = &slug
			}
		}
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
	body.Text = p.Text
	body.Image = p.Image
	body.Video = p.Video
	body.Audio = p.Audio
	body.Sticker = p.Sticker
	body.Document = p.Document
	body.Location = p.Location
	if len(p.Tags) > 0 {
		tags := p.Tags
		body.Tags = &tags
	}
	if len(p.Metadata) > 0 {
		metadata := p.Metadata
		body.Metadata = &metadata
	}
	return body
}

// Send sends one WhatsApp message. Retried safely: a single idempotency key is
// reused across attempts. Provide your own key with option.WithIdempotencyKey.
func (s *WhatsappService) Send(ctx context.Context, params WhatsappSendParams, opts ...option.RequestOption) (*WhatsAppMessage, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.CreateWhatsAppMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateWhatsAppMessage(ctx, p, wire, cfg...)
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
