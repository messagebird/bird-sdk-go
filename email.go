package bird

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
	"github.com/messagebird/bird-sdk-go/option"
)

// EmailService sends and reads email messages. Reach it via Client.Email.
type EmailService struct {
	resource

	// Stats reads aggregated delivery and engagement statistics.
	Stats *EmailStatsService

	// Mailboxes manages durable agent mailboxes that receive, store, and send email.
	Mailboxes *EmailMailboxesService

	// Threads reads and manages email conversations across every mailbox.
	Threads *EmailThreadsService
}

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
	// with Parameters. The value is the template's ID (`emt_…`) or its slug handle.
	Template string
	// Language selects which of the template's languages to send, as a BCP-47 tag
	// (e.g. "en", "pt-BR"). Template sends only. Omit it to send the template's
	// default language, unless the template's language_source_required is true, in
	// which case a send naming none is rejected. A language the template doesn't
	// carry is resolved by the template's own on_missing_language setting
	// (fallback to the closest match, or fail the send).
	Language string
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
		category := oapi.EmailMessageCategory(p.Category)
		body.Category = &category
	}
	if p.IpPoolId != "" {
		ipPool := p.IpPoolId
		body.IpPoolId = &ipPool
	}
	// A template send nests its reference (id or slug), language, and variables
	// under the template object; an inline send uses the top-level parameters.
	// The two content modes are exclusive. The `emt_` prefix marks an id;
	// anything else is a slug handle. Language alone (no Template) still opens
	// the template object, so a caller who sets it without a reference gets the
	// server's own validation error rather than a silently dropped field.
	if p.Template != "" || p.Language != "" {
		var tmpl oapi.EmailTemplateSend
		if p.Template != "" {
			if strings.HasPrefix(p.Template, "emt_") {
				id := oapi.EmailTemplateID(p.Template)
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
		category := oapi.EmailMessageCategory(d.Category)
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
