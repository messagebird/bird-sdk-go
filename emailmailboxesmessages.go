package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
	"github.com/messagebird/bird-sdk-go/option"
)

// EmailMailboxesMessagesService sends messages from a mailbox's own address.
// Reach it via Client.Email.Mailboxes.Messages.
type EmailMailboxesMessagesService struct{ resource }

// EmailMailboxesMessagesCreateParams sends a new message from the mailbox.
type EmailMailboxesMessagesCreateParams struct {
	To       []string // required; plain address or "Name <addr>"
	Subject  string   // required
	HTML     string
	Text     string
	CC       []string
	BCC      []string
	ReplyTo  []string
	Category string // marketing | transactional
	Tags     []EmailTag
	Metadata map[string]any
}

func (p EmailMailboxesMessagesCreateParams) toWire() oapi.EmailMailboxComposeRequest {
	body := oapi.EmailMailboxComposeRequest{
		Subject: p.Subject,
		To:      addressInputs(p.To),
	}
	if len(p.CC) > 0 {
		cc := addressInputs(p.CC)
		body.Cc = &cc
	}
	if len(p.BCC) > 0 {
		bcc := addressInputs(p.BCC)
		body.Bcc = &bcc
	}
	if len(p.ReplyTo) > 0 {
		rt := addressInputs(p.ReplyTo)
		body.ReplyTo = &rt
	}
	if p.HTML != "" {
		html := p.HTML
		body.Html = &html
	}
	if p.Text != "" {
		text := p.Text
		body.Text = &text
	}
	if p.Category != "" {
		c := oapi.EmailMessageCategory(p.Category)
		body.Category = &c
	}
	if len(p.Tags) > 0 {
		t := p.Tags
		body.Tags = &t
	}
	if len(p.Metadata) > 0 {
		m := p.Metadata
		body.Metadata = &m
	}
	return body
}

// applyComposeDefaults fills any field the compose params left unset from the
// configured email defaults. Only four of them: a compose has no From (the
// mailbox is the sender) and no sending-infrastructure fields, and the request
// is rejected for one it does not accept.
func applyComposeDefaults(wire *oapi.EmailMailboxComposeRequest, d requestconfig.EmailDefaults) {
	if wire.ReplyTo == nil && len(d.ReplyTo) > 0 {
		replyTo := addressInputs(d.ReplyTo)
		wire.ReplyTo = &replyTo
	}
	if wire.Category == nil && d.Category != "" {
		category := oapi.EmailMessageCategory(d.Category)
		wire.Category = &category
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

// Compose sends a new email from the mailbox's own address, starting a new
// conversation. Retried safely with a reused idempotency key.
func (s *EmailMailboxesMessagesService) Create(ctx context.Context, mailboxID string, params EmailMailboxesMessagesCreateParams, opts ...option.RequestOption) (*EmailThreadMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	applyComposeDefaults(&wire, cfg.EmailDefaults)
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		op := &oapi.CreateMailboxMessageParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateMailboxMessage(ctx, oapi.MailboxID(mailboxID), op, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailThreadMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
