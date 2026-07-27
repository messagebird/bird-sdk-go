package bird

import (
	"context"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// MailboxThreadMessageService reads messages in a conversation thread and
// sends replies. Reach it via Client.MailboxThreadMessage.
type MailboxThreadMessageService struct{ client *Client }

// MailboxThreadMessageListParams filters the message list.
type MailboxThreadMessageListParams struct {
	Label     string // inbox | archive | spam | blocked | trash | unread | custom
	Direction string // inbound | outbound
	Include   string // extracted_text
	Limit     int
}

func (p MailboxThreadMessageListParams) toWire(startingAfter string) *oapi.ListEmailThreadMessagesParams {
	w := &oapi.ListEmailThreadMessagesParams{}
	if p.Label != "" {
		l := p.Label
		w.Label = &l
	}
	if p.Direction != "" {
		d := oapi.ListEmailThreadMessagesParamsDirection(p.Direction)
		w.Direction = &d
	}
	if p.Include != "" {
		inc := oapi.ListEmailThreadMessagesParamsInclude(p.Include)
		w.Include = &inc
	}
	if p.Limit > 0 {
		l := oapi.PaginationLimit(p.Limit)
		w.Limit = &l
	}
	if startingAfter != "" {
		c := oapi.StartingAfter(startingAfter)
		w.StartingAfter = &c
	}
	return w
}

// MailboxThreadMessageReplyParams sends a reply to a message.
type MailboxThreadMessageReplyParams struct {
	Text     string // required if HTML is empty
	HTML     string // required if Text is empty
	ReplyAll bool
	Category string // marketing | transactional
	Metadata map[string]any
}

func (p MailboxThreadMessageReplyParams) toWire() oapi.EmailThreadMessageReplyRequest {
	body := oapi.EmailThreadMessageReplyRequest{}
	if p.Text != "" {
		text := p.Text
		body.Text = &text
	}
	if p.HTML != "" {
		html := p.HTML
		body.Html = &html
	}
	if p.ReplyAll {
		v := true
		body.ReplyAll = &v
	}
	if p.Category != "" {
		c := oapi.EmailThreadMessageReplyRequestCategory(p.Category)
		body.Category = &c
	}
	if len(p.Metadata) > 0 {
		m := p.Metadata
		body.Metadata = &m
	}
	return body
}

// Get returns metadata for a single message (not the body; call Body for that).
func (s *MailboxThreadMessageService) Get(ctx context.Context, threadID, messageID string, opts ...option.RequestOption) (*EmailThreadMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetEmailThreadMessage(ctx, threadID, messageID, s.client.callEditors(cfg)...)
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

// Body returns the parsed HTML and plain-text body of a message.
func (s *MailboxThreadMessageService) Body(ctx context.Context, threadID, messageID string, opts ...option.RequestOption) (*EmailThreadMessageBody, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetEmailThreadMessageBody(ctx, threadID, messageID, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailThreadMessageBody
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Reply sends a reply to a specific message from the mailbox's own address.
// Retried safely with a reused idempotency key.
func (s *MailboxThreadMessageService) Reply(ctx context.Context, threadID, messageID string, params MailboxThreadMessageReplyParams, opts ...option.RequestOption) (*EmailThreadMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.ReplyEmailThreadMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.ReplyEmailThreadMessage(ctx, threadID, messageID, p, wire, s.client.callEditors(cfg)...)
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

// Attachments lists the attachment manifest for a message.
func (s *MailboxThreadMessageService) Attachments(ctx context.Context, threadID, messageID string, opts ...option.RequestOption) (*EmailThreadMessageAttachmentList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListEmailThreadMessageAttachments(ctx, threadID, messageID, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailThreadMessageAttachmentList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPage fetches one page of messages in a thread.
func (s *MailboxThreadMessageService) ListPage(ctx context.Context, threadID string, params MailboxThreadMessageListParams, startingAfter string, opts ...option.RequestOption) (*EmailThreadMessageList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListEmailThreadMessages(ctx, threadID, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailThreadMessageList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every message in a thread, fetching pages lazily.
func (s *MailboxThreadMessageService) List(ctx context.Context, threadID string, params MailboxThreadMessageListParams, opts ...option.RequestOption) iter.Seq2[*EmailThreadMessage, error] {
	return paginate(func(cursor string) ([]EmailThreadMessage, *string, error) {
		page, err := s.ListPage(ctx, threadID, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
