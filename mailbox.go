package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// MailboxService manages agent mailboxes — durable inboxes on inbox.ai or
// your own domain that receive, store, and send email. Reach it via
// Client.Mailbox.
type MailboxService struct{ resource }

// MailboxComposeParams sends a new message from the mailbox.
type MailboxComposeParams struct {
	To       []string // required; plain address or "Name <addr>"
	Subject  string   // required
	HTML     string
	Text     string
	CC       []string
	BCC      []string
	ReplyTo  []string
	Category string // marketing | transactional
	Metadata map[string]any
}

func parseAddresses(addrs []string) []oapi.EmailAddressInput {
	out := make([]oapi.EmailAddressInput, len(addrs))
	for i, a := range addrs {
		_ = out[i].FromEmailAddressInput0(a)
	}
	return out
}

func (p MailboxComposeParams) toWire() oapi.EmailMailboxComposeRequest {
	body := oapi.EmailMailboxComposeRequest{
		Subject: p.Subject,
		To:      parseAddresses(p.To),
	}
	if len(p.CC) > 0 {
		cc := parseAddresses(p.CC)
		body.Cc = &cc
	}
	if len(p.BCC) > 0 {
		bcc := parseAddresses(p.BCC)
		body.Bcc = &bcc
	}
	if len(p.ReplyTo) > 0 {
		rt := parseAddresses(p.ReplyTo)
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
	if len(p.Metadata) > 0 {
		m := p.Metadata
		body.Metadata = &m
	}
	return body
}

// Compose sends a new email from the mailbox's own address, starting a new
// conversation. Retried safely with a reused idempotency key.
func (s *MailboxService) Compose(ctx context.Context, mailboxID string, params MailboxComposeParams, opts ...option.RequestOption) (*EmailThreadMessage, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreateMailboxMessageParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateMailboxMessage(ctx, oapi.MailboxID(mailboxID), op, wire, cfg...)
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
