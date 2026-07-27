package bird

import (
	"context"
	"iter"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// MailboxService manages agent mailboxes — durable inboxes on inbox.ai or
// your own domain that receive, store, and send email. Reach it via
// Client.Mailbox.
type MailboxService struct{ client *Client }

// MailboxCreateParams creates a mailbox. All fields are optional; omit
// LocalPart to have Bird generate a random handle on inbox.ai.
type MailboxCreateParams struct {
	LocalPart      string // address before @; omit to generate
	Domain         string // defaults to inbox.ai
	DisplayName    string
	DefaultReplyTo string
	ReceivePolicy  string // open | replies_only | allowlist | drop
	RetentionTier  string // 30d (only value today)
	Metadata       map[string]any
}

func (p MailboxCreateParams) toWire() oapi.MailboxCreate {
	body := oapi.MailboxCreate{}
	if p.LocalPart != "" {
		lp := p.LocalPart
		body.LocalPart = &lp
	}
	if p.Domain != "" {
		d := p.Domain
		body.Domain = &d
	}
	if p.DisplayName != "" {
		dn := p.DisplayName
		body.DisplayName = &dn
	}
	if p.DefaultReplyTo != "" {
		e := openapi_types.Email(p.DefaultReplyTo)
		body.DefaultReplyTo = &e
	}
	if p.ReceivePolicy != "" {
		rp := oapi.MailboxCreateReceivePolicy(p.ReceivePolicy)
		body.ReceivePolicy = &rp
	}
	if p.RetentionTier != "" {
		rt := oapi.MailboxCreateRetentionTier(p.RetentionTier)
		body.RetentionTier = &rt
	}
	if len(p.Metadata) > 0 {
		m := p.Metadata
		body.Metadata = &m
	}
	return body
}

// MailboxUpdateParams is a partial update. Nil fields leave values unchanged.
type MailboxUpdateParams struct {
	DisplayName    *string
	DefaultReplyTo *string
	ReceivePolicy  *string
	RetentionTier  *string
	Metadata       map[string]any
	Confirm        bool // required when lowering retention_tier
}

func (p MailboxUpdateParams) toWire() oapi.MailboxUpdate {
	body := oapi.MailboxUpdate{}
	body.DisplayName = p.DisplayName
	if p.DefaultReplyTo != nil {
		e := openapi_types.Email(*p.DefaultReplyTo)
		body.DefaultReplyTo = &e
	}
	if p.ReceivePolicy != nil {
		rp := oapi.MailboxUpdateReceivePolicy(*p.ReceivePolicy)
		body.ReceivePolicy = &rp
	}
	if p.RetentionTier != nil {
		rt := oapi.MailboxUpdateRetentionTier(*p.RetentionTier)
		body.RetentionTier = &rt
	}
	if p.Metadata != nil {
		m := p.Metadata
		body.Metadata = &m
	}
	return body
}

// MailboxListParams filters the mailbox list.
type MailboxListParams struct {
	Q              string
	Address        string
	Domain         string
	State          string // active | suspended | deleted
	IncludeDeleted bool
	Limit          int
}

func (p MailboxListParams) toWire(startingAfter string) *oapi.ListMailboxesParams {
	w := &oapi.ListMailboxesParams{}
	if p.Q != "" {
		q := p.Q
		w.Q = &q
	}
	if p.Address != "" {
		e := openapi_types.Email(p.Address)
		w.Address = &e
	}
	if p.Domain != "" {
		d := p.Domain
		w.Domain = &d
	}
	if p.State != "" {
		state := oapi.ListMailboxesParamsState(p.State)
		w.State = &state
	}
	if p.IncludeDeleted {
		v := true
		w.IncludeDeleted = &v
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

// MailboxStatsParams bounds the stats window.
type MailboxStatsParams struct {
	From        string // YYYY-MM-DD or RFC3339 hour
	To          string
	Timezone    string
	Granularity string // day | hour
}

func (p MailboxStatsParams) toWire() *oapi.GetMailboxStatsParams {
	w := &oapi.GetMailboxStatsParams{}
	if p.From != "" {
		v := p.From
		w.From = &v
	}
	if p.To != "" {
		v := p.To
		w.To = &v
	}
	if p.Timezone != "" {
		v := p.Timezone
		w.Timezone = &v
	}
	if p.Granularity != "" {
		g := oapi.GetMailboxStatsParamsGranularity(p.Granularity)
		w.Granularity = &g
	}
	return w
}

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
		c := oapi.EmailMailboxComposeRequestCategory(p.Category)
		body.Category = &c
	}
	if len(p.Metadata) > 0 {
		m := p.Metadata
		body.Metadata = &m
	}
	return body
}

// Create claims a new mailbox address and returns the mailbox. Retried safely
// with a reused idempotency key.
func (s *MailboxService) Create(ctx context.Context, params MailboxCreateParams, opts ...option.RequestOption) (*Mailbox, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateMailboxParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateMailbox(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Mailbox
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single mailbox by id.
func (s *MailboxService) Get(ctx context.Context, mailboxID string, opts ...option.RequestOption) (*Mailbox, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetMailbox(ctx, mailboxID, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Mailbox
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update edits a mailbox. Only the fields set in params change. Set
// params.Confirm = true when lowering the retention tier.
func (s *MailboxService) Update(ctx context.Context, mailboxID string, params MailboxUpdateParams, opts ...option.RequestOption) (*Mailbox, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UpdateMailboxParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		if params.Confirm {
			v := true
			p.Confirm = &v
		}
		return s.client.oapi.UpdateMailbox(ctx, mailboxID, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Mailbox
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete soft-deletes the mailbox. The address is quarantined for 30 days
// and can be restored with Restore.
func (s *MailboxService) Delete(ctx context.Context, mailboxID string, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DeleteMailboxParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DeleteMailbox(ctx, mailboxID, p, s.client.callEditors(cfg)...)
	})
	return err
}

// Restore recovers a deleted mailbox within its 30-day restore window.
func (s *MailboxService) Restore(ctx context.Context, mailboxID string, opts ...option.RequestOption) (*Mailbox, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.RestoreMailboxParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.RestoreMailbox(ctx, mailboxID, p, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Mailbox
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Resume reactivates a suspended mailbox.
func (s *MailboxService) Resume(ctx context.Context, mailboxID string, opts ...option.RequestOption) (*Mailbox, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.ResumeMailboxParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.ResumeMailbox(ctx, mailboxID, p, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Mailbox
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stats returns per-mailbox email activity time series.
func (s *MailboxService) Stats(ctx context.Context, mailboxID string, params MailboxStatsParams, opts ...option.RequestOption) (*MailboxStatsResponse, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetMailboxStats(ctx, mailboxID, params.toWire(), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out MailboxStatsResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Compose sends a new email from the mailbox's own address, starting a new
// conversation. Retried safely with a reused idempotency key.
func (s *MailboxService) Compose(ctx context.Context, mailboxID string, params MailboxComposeParams, opts ...option.RequestOption) (*EmailThreadMessage, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateMailboxMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateMailboxMessage(ctx, mailboxID, p, wire, s.client.callEditors(cfg)...)
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

// Labels lists the labels available in this mailbox.
func (s *MailboxService) Labels(ctx context.Context, mailboxID string, opts ...option.RequestOption) (*EmailMailboxLabelList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListMailboxLabels(ctx, mailboxID, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailMailboxLabelList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPage fetches one page of mailboxes. Pass the previous page's NextCursor
// as startingAfter to advance; "" starts from the most recent.
func (s *MailboxService) ListPage(ctx context.Context, params MailboxListParams, startingAfter string, opts ...option.RequestOption) (*MailboxList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListMailboxes(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out MailboxList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every mailbox matching params, fetching pages lazily.
func (s *MailboxService) List(ctx context.Context, params MailboxListParams, opts ...option.RequestOption) iter.Seq2[*Mailbox, error] {
	return paginate(func(cursor string) ([]Mailbox, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
