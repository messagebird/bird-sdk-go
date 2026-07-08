package bird

import (
	"context"
	"iter"
	"net/http"
	"strings"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// SMSService sends SMS messages — free text or by stored template — and reads
// them back. Reach it via Client.Sms.
type SMSService struct{ client *Client }

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
// Template (by id or alias, with Parameters) — the two are mutually exclusive.
// Zero-value fields are omitted from the request.
type SmsSendParams struct {
	To         string         // required; recipient phone number in E.164 format
	From       string         // optional sender; Bird selects one when empty
	Text       string         // free-text body (mutually exclusive with Template)
	Category   SMSCategory    // required with Text; omit on a template send
	Template   string         // stored template id (smt_…) or alias (mutually exclusive with Text)
	Locale     string         // template language as a BCP-47 tag; template sends only
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
	// A template send folds the template id/alias, locale, and parameters into the
	// nested template object.
	if p.Template != "" || p.Locale != "" || len(p.Parameters) > 0 {
		var tmpl oapi.SMSTemplateSend
		if p.Template != "" {
			// An smt_-prefixed value is the id; anything else is the alias handle.
			if strings.HasPrefix(p.Template, "smt_") {
				id := oapi.SMSTemplateID(p.Template)
				tmpl.Id = &id
			} else {
				alias := p.Template
				tmpl.Alias = &alias
			}
		}
		if p.Locale != "" {
			locale := p.Locale
			tmpl.Locale = &locale
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
func (s *SMSService) Send(ctx context.Context, params SmsSendParams, opts ...option.RequestOption) (*SMSMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateSMSMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateSMSMessage(ctx, p, wire, s.client.callEditors(cfg)...)
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
func (s *SMSService) SendBatch(ctx context.Context, params SmsSendBatchParams, opts ...option.RequestOption) (*SMSBatch, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateSMSMessageBatchParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateSMSMessageBatch(ctx, p, wire, s.client.callEditors(cfg)...)
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

// Get returns a single SMS message with its current delivery status, segment
// breakdown, cost, and failure detail if it failed.
func (s *SMSService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*SMSMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetSMSMessage(ctx, id, s.client.callEditors(cfg)...)
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

// SmsListParams filters the message list. Zero-value fields are omitted.
type SmsListParams struct {
	Direction     string      // "outbound" or "inbound"; empty for both
	Statuses      []string    // filter by any of several delivery statuses
	ErrorCodes    []string    // filter by any of several failure reasons
	Category      SMSCategory // filter by content classification
	To            string      // filter by recipient (E.164 exact match)
	From          string      // filter by sender (exact match)
	Tags          []string    // filter by tag name or name:value; AND-combined
	CreatedAfter  time.Time
	CreatedBefore time.Time
	Limit         int
}

func (p SmsListParams) toWire(startingAfter string) *oapi.ListSMSMessagesParams {
	w := &oapi.ListSMSMessagesParams{}
	if p.Direction != "" {
		direction := oapi.ListSMSMessagesParamsDirection(p.Direction)
		w.Direction = &direction
	}
	if len(p.Statuses) > 0 {
		statuses := p.Statuses
		w.Status = &statuses
	}
	if len(p.ErrorCodes) > 0 {
		errorCodes := p.ErrorCodes
		w.ErrorCode = &errorCodes
	}
	if p.Category != "" {
		category := oapi.ListSMSMessagesParamsCategory(p.Category)
		w.Category = &category
	}
	if p.To != "" {
		to := p.To
		w.To = &to
	}
	if p.From != "" {
		from := p.From
		w.From = &from
	}
	if len(p.Tags) > 0 {
		tags := p.Tags
		w.Tag = &tags
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
func (s *SMSService) ListPage(ctx context.Context, params SmsListParams, startingAfter string, opts ...option.RequestOption) (*SMSMessageList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListSMSMessages(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out SMSMessageList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every message matching params, fetching pages lazily. Range over it;
// the second value is non-nil only on the iteration where a fetch failed.
func (s *SMSService) List(ctx context.Context, params SmsListParams, opts ...option.RequestOption) iter.Seq2[*SMSMessage, error] {
	return paginate(func(cursor string) ([]SMSMessage, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
