package bird

import (
	"context"
	"net/http"
	"strings"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// SmsService sends SMS messages — free text or by stored template — and reads
// them back. Reach it via Client.Sms.
type SmsService struct{ resource }

// SMSCategory classifies a send for opt-out (STOP) policy, quiet hours, and
// per-country compliance.
type SMSCategory = oapi.SMSMessageCategory

const (
	SMSCategoryTransactional  SMSCategory = "transactional"
	SMSCategoryMarketing      SMSCategory = "marketing"
	SMSCategoryAuthentication SMSCategory = "authentication"
	SMSCategoryService        SMSCategory = "service"
)

// SmsSendParams is a single SMS send. Provide either Text (with Category) or a
// Template (by id or name, with Parameters) — the two are mutually exclusive.
// Zero-value fields are omitted from the request.
type SmsSendParams struct {
	To         string         // required; recipient phone number in E.164 format
	From       string         // optional sender; Bird selects one when empty
	Text       string         // free-text body (mutually exclusive with Template)
	Category   SMSCategory    // required with Text; omit on a template send
	Template   string         // stored template id (smt_…) or name (mutually exclusive with Text)
	Language   string         // template language as a BCP-47 tag; template sends only
	Parameters map[string]any // template variable values; template sends only
	Tags       []SMSTag       // structured {name, value} labels for filtering and analytics
	Metadata   map[string]any // arbitrary JSON stored on the message and echoed in webhooks
}

func (p SmsSendParams) toWire() oapi.SMSMessageSendRequest {
	body := oapi.SMSMessageSendRequest{To: p.To}
	if p.From != "" {
		from := p.From
		body.From = &from
	}
	if p.Text != "" {
		text := p.Text
		body.Text = &text
	}
	if p.Category != "" {
		category := p.Category
		body.Category = &category
	}
	// A template send folds the template id/name, language, and parameters into the
	// nested template object.
	if p.Template != "" || p.Language != "" || len(p.Parameters) > 0 {
		var tmpl oapi.SMSTemplateSend
		if p.Template != "" {
			// An smt_-prefixed value is the id; anything else is the name handle.
			if strings.HasPrefix(p.Template, "smt_") {
				id := oapi.SMSTemplateID(p.Template)
				tmpl.Id = &id
			} else {
				name := p.Template
				tmpl.Name = &name
			}
		}
		if p.Language != "" {
			language := p.Language
			tmpl.Language = &language
		}
		if len(p.Parameters) > 0 {
			params := p.Parameters
			tmpl.Parameters = &params
		}
		body.Template = &tmpl
	}
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

// Send sends one SMS message. Retried safely: a single idempotency key is reused
// across attempts. Provide your own key with option.WithIdempotencyKey.
func (s *SmsService) Send(ctx context.Context, params SmsSendParams, opts ...option.RequestOption) (*SMSMessage, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.CreateSMSMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateSMSMessage(ctx, p, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out SMSMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SmsSendBatchParams is a batch of up to 100 independent SMS sends.
type SmsSendBatchParams struct {
	Messages []SmsSendParams
}

func (p SmsSendBatchParams) toWire() oapi.SMSMessageBatchRequest {
	batch := make(oapi.SMSMessageBatchRequest, len(p.Messages))
	for i, m := range p.Messages {
		batch[i] = m.toWire()
	}
	return batch
}

// SendBatch sends up to 100 independent SMS messages in one call. Each item is a
// full send with its own id, status, and cost; all items are validated before any
// are queued. Retried safely with a reused idempotency key.
func (s *SmsService) SendBatch(ctx context.Context, params SmsSendBatchParams, opts ...option.RequestOption) (*SMSBatch, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.CreateSMSMessageBatchParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateSMSMessageBatch(ctx, p, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out SMSBatch
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
