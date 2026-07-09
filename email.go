package bird

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
	"github.com/messagebird/bird-sdk-go/option"
)

// EmailService sends and reads email messages. Reach it via Client.Email.
type EmailService struct{ client *Client }

// EmailSendParams is an email send. Optional fields are omitted from the request
// when left at their zero value.
//
// Address fields (From, To, Cc, Bcc, ReplyTo) accept either a bare email address
// or RFC 5322 mailbox syntax with a display name: "Support Team <support@example.com>".
type EmailSendParams struct {
	From        string            // sender; bare address or "Name <addr>" form; must be on a verified domain
	To          []string          // primary recipients; each may be bare or "Name <addr>" form
	Cc          []string          // optional; same syntax as To
	Bcc         []string          // optional; same syntax as To
	ReplyTo     []string          // optional Reply-To; same syntax as To
	Subject     string            // subject line
	HTML        string            // HTML body; at least one of HTML or Text is required
	Text        string            // plain-text body
	Tags        []EmailTag        // structured {name,value} labels for filtering and analytics
	Metadata    map[string]any    // arbitrary JSON, echoed on reads and in webhook payloads
	Headers     map[string]string // custom email headers
	Attachments []EmailAttachment // file attachments
	Category    Category          // transactional (default) or marketing
	IpPoolId    string            // IP pool ID (ipp_…); workspace default when empty
	// TrackOpens and TrackClicks are pointers because the server default is
	// true — a nil leaves the default, false explicitly disables tracking.
	TrackOpens  *bool
	TrackClicks *bool
	// Template, when set, sends a published template in place of inline content:
	// leave Subject/HTML/Text empty (the template supplies them) and personalize
	// with Parameters. The value is the template's ID (`emt_…`) or its name handle.
	Template string
	// Parameters holds template variables rendered into the subject and
	// body at send time; works with both inline content and a Template.
	Parameters map[string]any
}

func (p EmailSendParams) toWire() (oapi.EmailMessageSendRequest, error) {
	var from oapi.EmailAddressInput
	if p.From != "" {
		var err error
		from, err = parseAddressInput(p.From)
		if err != nil {
			return oapi.EmailMessageSendRequest{}, fmt.Errorf("bird: invalid from address %q: %w", p.From, err)
		}
	}
	to, err := parseAddressInputs(p.To)
	if err != nil {
		return oapi.EmailMessageSendRequest{}, fmt.Errorf("bird: invalid to address: %w", err)
	}
	body := oapi.EmailMessageSendRequest{
		From: from,
		To:   to,
	}
	// Subject is optional on the wire (a template supplies it); omit it when
	// empty so a send-by-template doesn't trip subject/template exclusivity.
	if p.Subject != "" {
		subject := p.Subject
		body.Subject = &subject
	}
	if len(p.Cc) > 0 {
		cc, err := parseAddressInputs(p.Cc)
		if err != nil {
			return body, fmt.Errorf("bird: invalid cc address: %w", err)
		}
		body.Cc = &cc
	}
	if len(p.Bcc) > 0 {
		bcc, err := parseAddressInputs(p.Bcc)
		if err != nil {
			return body, fmt.Errorf("bird: invalid bcc address: %w", err)
		}
		body.Bcc = &bcc
	}
	if len(p.ReplyTo) > 0 {
		replyTo, err := parseAddressInputs(p.ReplyTo)
		if err != nil {
			return body, fmt.Errorf("bird: invalid reply_to address: %w", err)
		}
		body.ReplyTo = &replyTo
	}
	if p.HTML != "" {
		html := p.HTML
		body.Html = &html
	}
	if p.Text != "" {
		text := p.Text
		body.Text = &text
	}
	if len(p.Tags) > 0 {
		tags := p.Tags
		body.Tags = &tags
	}
	if len(p.Metadata) > 0 {
		metadata := p.Metadata
		body.Metadata = &metadata
	}
	if len(p.Headers) > 0 {
		headers := p.Headers
		body.Headers = &headers
	}
	if len(p.Attachments) > 0 {
		attachments := p.Attachments
		body.Attachments = &attachments
	}
	if p.Category != "" {
		category := oapi.EmailMessageSendRequestCategory(p.Category)
		body.Category = &category
	}
	if p.IpPoolId != "" {
		ipPool := p.IpPoolId
		body.IpPoolId = &ipPool
	}
	// A template send nests its reference (id or name) and variables under the
	// template object; an inline send uses the top-level parameters. The two
	// content modes are exclusive. The `emt_` prefix marks an id; anything else
	// is a name handle.
	if p.Template != "" {
		var tmpl oapi.EmailTemplateSend
		if strings.HasPrefix(p.Template, "emt_") {
			id := oapi.EmailTemplateID(p.Template)
			tmpl.Id = &id
		} else {
			name := p.Template
			tmpl.Name = &name
		}
		if len(p.Parameters) > 0 {
			parameters := p.Parameters
			tmpl.Parameters = &parameters
		}
		body.Template = &tmpl
	} else if len(p.Parameters) > 0 {
		parameters := p.Parameters
		body.Parameters = &parameters
	}
	body.TrackOpens = p.TrackOpens
	body.TrackClicks = p.TrackClicks
	return body, nil
}

// Send delivers an email and returns the created message. Sends are retried
// safely: a single idempotency key is reused across attempts, so a retry never
// double-delivers. Provide your own key with option.WithIdempotencyKey.
func (s *EmailService) Send(ctx context.Context, params EmailSendParams, opts ...option.RequestOption) (*EmailMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire, err := params.toWire()
	if err != nil {
		return nil, err
	}
	applyEmailDefaults(&wire, cfg.EmailDefaults)
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		params := &oapi.CreateEmailMessageParams{}
		if idempotencyKey != "" {
			params.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateEmailMessage(ctx, params, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailSendBatchParams is a batch of email sends submitted in one request. Each
// Message is an individual send; the whole batch is validated before any item is
// queued. The result preserves submission order.
type EmailSendBatchParams struct {
	Messages []EmailSendParams
}

func (p EmailSendBatchParams) toWire() (oapi.EmailMessageBatchRequest, error) {
	wire := make(oapi.EmailMessageBatchRequest, len(p.Messages))
	for i, m := range p.Messages {
		item, err := m.toWire()
		if err != nil {
			return nil, fmt.Errorf("bird: batch message %d: %w", i, err)
		}
		wire[i] = item
	}
	return wire, nil
}

// SendBatch queues multiple emails in one request and returns one result item
// per submitted message, in submission order. The whole batch is validated
// before any item is queued. Like Send, the batch is retried safely: a single
// idempotency key is reused across attempts, so a retry never double-delivers.
// Provide your own key with option.WithIdempotencyKey.
func (s *EmailService) SendBatch(ctx context.Context, params EmailSendBatchParams, opts ...option.RequestOption) (*EmailBatch, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire, err := params.toWire()
	if err != nil {
		return nil, err
	}
	for i := range wire {
		applyEmailDefaults(&wire[i], cfg.EmailDefaults)
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		params := &oapi.CreateEmailMessageBatchParams{}
		if idempotencyKey != "" {
			params.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateEmailMessageBatch(ctx, params, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailBatch
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single message by ID, with aggregate delivery status rolled up
// across recipients.
func (s *EmailService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*EmailMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetEmailMessage(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailListParams filters the message list. Zero-value fields are omitted.
type EmailListParams struct {
	Limit         int
	Status        EmailStatus
	Category      Category
	Tag           string
	To            string
	From          string
	CreatedAfter  time.Time
	CreatedBefore time.Time
}

func (p EmailListParams) toWire(startingAfter string) *oapi.ListEmailMessagesParams {
	w := &oapi.ListEmailMessagesParams{}
	if p.Limit > 0 {
		limit := oapi.PaginationLimit(p.Limit)
		w.Limit = &limit
	}
	if startingAfter != "" {
		cursor := oapi.StartingAfter(startingAfter)
		w.StartingAfter = &cursor
	}
	if p.Status != "" {
		status := oapi.ListEmailMessagesParamsStatus(p.Status)
		w.Status = &status
	}
	if p.Category != "" {
		category := oapi.ListEmailMessagesParamsCategory(p.Category)
		w.Category = &category
	}
	if p.Tag != "" {
		tags := []string{p.Tag}
		w.Tag = &tags
	}
	if p.To != "" {
		to := openapi_types.Email(p.To)
		w.To = &to
	}
	if p.From != "" {
		from := openapi_types.Email(p.From)
		w.From = &from
	}
	if !p.CreatedAfter.IsZero() {
		createdAfter := oapi.CreatedAfter(p.CreatedAfter)
		w.CreatedAfter = &createdAfter
	}
	if !p.CreatedBefore.IsZero() {
		createdBefore := oapi.CreatedBefore(p.CreatedBefore)
		w.CreatedBefore = &createdBefore
	}
	return w
}

// ListPage fetches one page of messages. Pass the previous page's NextCursor as
// startingAfter to advance; "" starts from the most recent.
func (s *EmailService) ListPage(ctx context.Context, params EmailListParams, startingAfter string, opts ...option.RequestOption) (*EmailMessageList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListEmailMessages(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailMessageList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every message matching params, fetching pages lazily. Range over
// it; the second value is non-nil only on the iteration where a fetch failed,
// after which the sequence ends.
//
//	for msg, err := range client.Email.List(ctx, bird.EmailListParams{Status: bird.EmailStatusBounced}) {
//		if err != nil { return err }
//		log.Println(msg.Id)
//	}
func (s *EmailService) List(ctx context.Context, params EmailListParams, opts ...option.RequestOption) iter.Seq2[*EmailMessage, error] {
	return paginate(func(cursor string) ([]EmailMessage, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}

// applyEmailDefaults fills any field the per-send params left unset (nil on the
// wire body) from the configured defaults. The per-send value always wins.
func applyEmailDefaults(wire *oapi.EmailMessageSendRequest, d requestconfig.EmailDefaults) {
	if isZeroAddressInput(wire.From) && d.From != "" {
		// Best-effort: malformed defaults are silently dropped rather than
		// failing here — Send already validated the per-send From above.
		if from, err := parseAddressInput(d.From); err == nil {
			wire.From = from
		}
	}
	if wire.ReplyTo == nil && len(d.ReplyTo) > 0 {
		if replyTo, err := parseAddressInputs(d.ReplyTo); err == nil {
			wire.ReplyTo = &replyTo
		}
	}
	if wire.Category == nil && d.Category != "" {
		category := oapi.EmailMessageSendRequestCategory(d.Category)
		wire.Category = &category
	}
	if wire.TrackOpens == nil {
		wire.TrackOpens = d.TrackOpens
	}
	if wire.TrackClicks == nil {
		wire.TrackClicks = d.TrackClicks
	}
	if wire.Headers == nil && len(d.Headers) > 0 {
		headers := d.Headers
		wire.Headers = &headers
	}
	if wire.Tags == nil && len(d.Tags) > 0 {
		tags := d.Tags
		wire.Tags = &tags
	}
	if wire.Metadata == nil && len(d.Metadata) > 0 {
		metadata := d.Metadata
		wire.Metadata = &metadata
	}
}

// parseAddressInput wraps an address string in the wire union's string arm
// verbatim — no client-side parsing. The wire's string form accepts both a plain
// address and an RFC 5322 mailbox with a display name ("Jane <jane@x.com>"), so
// the server parses; the SDK passes the string straight through.
func parseAddressInput(s string) (oapi.EmailAddressInput, error) {
	var inp oapi.EmailAddressInput
	err := inp.FromEmailAddressInput0(s)
	return inp, err
}

func parseAddressInputs(addresses []string) ([]oapi.EmailAddressInput, error) {
	out := make([]oapi.EmailAddressInput, len(addresses))
	for i, a := range addresses {
		inp, err := parseAddressInput(a)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", a, err)
		}
		out[i] = inp
	}
	return out, nil
}

// isZeroAddressInput reports whether an EmailAddressInput is at its zero value
// (no data has been set on it). Used to check whether applyEmailDefaults should
// fill a From from the configured defaults.
func isZeroAddressInput(a oapi.EmailAddressInput) bool {
	b, err := a.MarshalJSON()
	return err != nil || string(b) == "null"
}

func decodeBody(body []byte, out any) error {
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("bird: decoding response: %w", err)
	}
	return nil
}
