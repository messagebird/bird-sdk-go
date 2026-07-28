package bird

import (
	"context"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// MailboxThreadService reads and manages email conversation threads stored in
// mailboxes. Reach it via Client.MailboxThread.
type MailboxThreadService struct{ client *Client }

// MailboxThreadListParams filters the thread list.
type MailboxThreadListParams struct {
	MailboxID   string
	ContactID   string
	Label       []string // inbox | archive | spam | blocked | custom
	HasUnread   bool
	Participant string // address filter (contains-match)
	Subject     string // subject contains filter
	Limit       int
}

func (p MailboxThreadListParams) toWire(startingAfter string) *oapi.ListEmailThreadsParams {
	w := &oapi.ListEmailThreadsParams{}
	if p.MailboxID != "" {
		id := oapi.MailboxID(p.MailboxID)
		w.MailboxId = &id
	}
	if p.ContactID != "" {
		id := oapi.ContactID(p.ContactID)
		w.ContactId = &id
	}
	if len(p.Label) > 0 {
		labels := p.Label
		w.Label = &labels
	}
	if p.HasUnread {
		v := true
		w.HasUnread = &v
	}
	if p.Participant != "" {
		v := p.Participant
		w.Participant = &v
	}
	if p.Subject != "" {
		v := p.Subject
		w.Subject = &v
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

// MailboxThreadUpdateParams applies label changes and contact links to a thread.
type MailboxThreadUpdateParams struct {
	AddLabels    []string
	RemoveLabels []string
	ContactID    *string // nil = unchanged; non-empty ptr = link contact. To unlink (send null) use client.Patch directly.
}

func (p MailboxThreadUpdateParams) toWire() oapi.EmailThreadUpdateRequest {
	body := oapi.EmailThreadUpdateRequest{}
	if len(p.AddLabels) > 0 || len(p.RemoveLabels) > 0 {
		lu := &oapi.EmailLabelsUpdate{}
		if len(p.AddLabels) > 0 {
			add := p.AddLabels
			lu.Add = &add
		}
		if len(p.RemoveLabels) > 0 {
			remove := p.RemoveLabels
			lu.Remove = &remove
		}
		body.Labels = lu
	}
	if p.ContactID != nil && *p.ContactID != "" {
		body.ContactId = Value(oapi.ContactID(*p.ContactID))
	}
	return body
}

// Get returns a single thread by id.
func (s *MailboxThreadService) Get(ctx context.Context, threadID string, opts ...option.RequestOption) (*EmailThread, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetEmailThread(ctx, threadID, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailThread
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies label changes or contact link changes to a thread.
func (s *MailboxThreadService) Update(ctx context.Context, threadID string, params MailboxThreadUpdateParams, opts ...option.RequestOption) (*EmailThread, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UpdateEmailThreadParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateEmailThread(ctx, threadID, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailThread
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete moves a thread and all its messages to trash. Pass permanent=true to
// delete immediately without a restore window.
func (s *MailboxThreadService) Delete(ctx context.Context, threadID string, permanent bool, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DeleteEmailThreadParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		if permanent {
			v := true
			p.Permanent = &v
		}
		return s.client.oapi.DeleteEmailThread(ctx, threadID, p, s.client.callEditors(cfg)...)
	})
	return err
}

// ListPage fetches one page of threads matching params.
func (s *MailboxThreadService) ListPage(ctx context.Context, params MailboxThreadListParams, startingAfter string, opts ...option.RequestOption) (*EmailThreadList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListEmailThreads(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailThreadList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every thread matching params, fetching pages lazily.
func (s *MailboxThreadService) List(ctx context.Context, params MailboxThreadListParams, opts ...option.RequestOption) iter.Seq2[*EmailThread, error] {
	return paginate(func(cursor string) ([]EmailThread, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
