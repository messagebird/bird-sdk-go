package bird

import (
	"context"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// MailboxReceiveRuleService manages per-sender allow/block rules on a mailbox.
// Reach it via Client.MailboxReceiveRule.
type MailboxReceiveRuleService struct{ client *Client }

// MailboxReceiveRuleCreateParams adds a rule to a mailbox.
type MailboxReceiveRuleCreateParams struct {
	Action string // allow | block; required
	Entry  string // address or domain to match; required
	Note   string // optional explanation
}

func (p MailboxReceiveRuleCreateParams) toWire() oapi.ReceiveRuleCreate {
	body := oapi.ReceiveRuleCreate{
		Action: oapi.ReceiveRuleCreateAction(p.Action),
		Entry:  p.Entry,
	}
	if p.Note != "" {
		note := p.Note
		body.Note = &note
	}
	return body
}

// MailboxReceiveRuleListParams filters the rule list.
type MailboxReceiveRuleListParams struct {
	Action string // allow | block
	Limit  int
}

func (p MailboxReceiveRuleListParams) toWire(startingAfter string) *oapi.ListMailboxReceiveRulesParams {
	w := &oapi.ListMailboxReceiveRulesParams{}
	if p.Action != "" {
		a := oapi.ListMailboxReceiveRulesParamsAction(p.Action)
		w.Action = &a
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

// Create adds an allow or block rule to a mailbox. Block rules always win.
// Retried safely with a reused idempotency key.
func (s *MailboxReceiveRuleService) Create(ctx context.Context, mailboxID string, params MailboxReceiveRuleCreateParams, opts ...option.RequestOption) (*ReceiveRule, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateMailboxReceiveRuleParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateMailboxReceiveRule(ctx, mailboxID, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ReceiveRule
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a receive rule.
func (s *MailboxReceiveRuleService) Delete(ctx context.Context, mailboxID, ruleID string, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DeleteMailboxReceiveRuleParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DeleteMailboxReceiveRule(ctx, mailboxID, ruleID, p, s.client.callEditors(cfg)...)
	})
	return err
}

// ListPage fetches one page of rules for a mailbox.
func (s *MailboxReceiveRuleService) ListPage(ctx context.Context, mailboxID string, params MailboxReceiveRuleListParams, startingAfter string, opts ...option.RequestOption) (*ReceiveRuleList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListMailboxReceiveRules(ctx, mailboxID, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ReceiveRuleList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every receive rule for a mailbox, fetching pages lazily.
func (s *MailboxReceiveRuleService) List(ctx context.Context, mailboxID string, params MailboxReceiveRuleListParams, opts ...option.RequestOption) iter.Seq2[*ReceiveRule, error] {
	return paginate(func(cursor string) ([]ReceiveRule, *string, error) {
		page, err := s.ListPage(ctx, mailboxID, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
