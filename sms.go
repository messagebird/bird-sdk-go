package bird

import (
	"context"
	"net/http"
	"strings"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// SmsService sends free-text or stored-template SMS messages and reads them
// back. Reach it through Client.Sms.
type SmsService struct {
	resource

	// Stats reads aggregate statistics over the workspace's SMS traffic.
	Stats *SmsStatsService
}

// SMSCategory classifies a send: carriers see it, and a country sender
// registration is approved for a category, so a send outside what it covers is
// refused. Quiet hours do not key on it; nothing implements them.
type SMSCategory = oapi.SMSMessageCategory

const (
	SMSCategoryTransactional  SMSCategory = "transactional"
	SMSCategoryMarketing      SMSCategory = "marketing"
	SMSCategoryAuthentication SMSCategory = "authentication"
	SMSCategoryService        SMSCategory = "service"
)

// SmsSendParams is a single SMS send. Provide one of Text (with Category and
// From), a Template (by id or slug, with Parameters), or Binary (with Category
// transactional). A workspace template requires an owned From; a built-in
// template selects its sender and rejects From. The content forms are mutually
// exclusive.
// Zero-value fields are omitted from the request.
type SmsSendParams struct {
	To         string         // required; recipient phone number in E.164 format
	From       string         // required with Text or a workspace Template; omit with a built-in Template
	Text       string         // free-text body (mutually exclusive with Template)
	Category   SMSCategory    // required with Text; omit on a template send
	Template   string         // stored template id (smt_…) or slug (mutually exclusive with Text)
	Language   string         // template language as a BCP-47 tag; template sends only
	Parameters map[string]any // template variable values; template sends only
	Tags       []SMSTag       // structured {name, value} labels for filtering and analytics
	Metadata   map[string]any // arbitrary JSON stored on the message and echoed in webhooks
	// SmartEncoding replaces characters outside the GSM-7 alphabet with their closest
	// equivalent, which often lowers the segment count and the cost. A pointer because
	// false is a real value the send carries: nil omits the option and takes the
	// default (off), &false records the choice explicitly.
	SmartEncoding *bool
	// IgnoreSuppression sends even if the recipient is suppressed for the sender.
	// Only authentication and transactional messages can use it, and it must be
	// enabled for your workspace first. A pointer for the same reason as SmartEncoding.
	IgnoreSuppression *bool
	// CallbackURL is an HTTPS URL that receives this message's events instead of
	// your webhook endpoints. Leave it unspecified to use the messaging profile's
	// URL, build it with Value to set one, and with Null to send this message's
	// events to your webhook subscriptions.
	CallbackURL Nullable[string]
	// MessagingProfileID supplies the sender and defaults when From, Category or an
	// option is left out.
	MessagingProfileID string
	// Reference is your own identifier for the message, such as an order number; the
	// message list filters on it.
	Reference string
	// ToCountryCode is the ISO 3166-1 alpha-2 country whose national numbering plan
	// To is written in, such as "NL" for 0612345678.
	ToCountryCode string
	// ValidityPeriod is how many seconds (60-172800) the message may wait for
	// delivery; zero uses the messaging profile's or carrier's default.
	ValidityPeriod int
	// CostCenter is your accounting identifier for the message.
	CostCenter string
	// Binary sends an 8-bit payload instead of Text or Template; it needs Category
	// transactional.
	Binary *SMSBinary
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
	// A template send folds the template id/slug, language, and parameters into the
	// nested template object.
	if p.Template != "" || p.Language != "" || len(p.Parameters) > 0 {
		var tmpl oapi.SMSTemplateSend
		if p.Template != "" {
			// An smt_-prefixed value is the id; anything else is the slug handle.
			if strings.HasPrefix(p.Template, "smt_") {
				id := oapi.SMSTemplateID(p.Template)
				tmpl.Id = &id
			} else {
				slug := p.Template
				tmpl.Slug = &slug
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
	body.MessagingProfileId = optStr(p.MessagingProfileID)
	body.Reference = optStr(p.Reference)
	body.ToCountryCode = optStr(p.ToCountryCode)
	body.CostCenter = optStr(p.CostCenter)
	body.ValidityPeriod = optInt(p.ValidityPeriod)
	body.Binary = p.Binary
	if p.SmartEncoding != nil || p.IgnoreSuppression != nil || p.CallbackURL.IsSpecified() {
		body.Options = &oapi.SMSSendOptions{
			SmartEncoding:     p.SmartEncoding,
			IgnoreSuppression: p.IgnoreSuppression,
			CallbackUrl:       p.CallbackURL,
		}
	}
	return body
}

// Send sends one SMS message. A workspace template requires an owned From; a
// built-in template selects its sender and rejects From. Retried safely: a single
// idempotency key is reused across attempts. Provide your own key with
// option.WithIdempotencyKey.
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
	batch := oapi.SMSMessageBatchRequest{Messages: make([]oapi.SMSMessageSendRequest, len(p.Messages))}
	for i, m := range p.Messages {
		batch.Messages[i] = m.toWire()
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
