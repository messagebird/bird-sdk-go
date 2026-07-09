package bird

import (
	"context"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// AudiencesService manages static audiences and their contact membership:
// create, read, update, delete, list, and add/remove contacts. Reach it via
// Client.Audiences.
type AudiencesService struct{ client *Client }

// AudienceCreateParams creates an audience. Name is required. Type is the
// audience's recipient model; leave it "" to default to "static" (the only
// currently usable value — "dynamic" and "external" are preview values the
// server rejects).
type AudienceCreateParams struct {
	Name        string // required
	Description string // optional
	Type        string // optional; "" defaults to "static" server-side
}

func (p AudienceCreateParams) toWire() oapi.AudienceCreateRequest {
	body := oapi.AudienceCreateRequest{Name: p.Name}
	if p.Description != "" {
		description := p.Description
		body.Description = &description
	}
	if p.Type != "" {
		audienceType := oapi.AudienceCreateRequestType(p.Type)
		body.Type = &audienceType
	}
	return body
}

// AudienceUpdateParams is a partial update of an audience. Every field is a
// pointer: nil leaves it unchanged; point Description at "" to clear it.
type AudienceUpdateParams struct {
	Name        *string
	Description *string
}

func (p AudienceUpdateParams) toWire() oapi.AudienceUpdateRequest {
	return oapi.AudienceUpdateRequest{
		Name:        p.Name,
		Description: p.Description,
	}
}

// AudienceListParams filters the audience list. Zero-value fields are
// omitted.
type AudienceListParams struct {
	Limit int
}

func (p AudienceListParams) toWire(startingAfter string) *oapi.ListAudiencesParams {
	w := &oapi.ListAudiencesParams{}
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

// AudienceListContactsParams bounds a page of an audience's contacts.
// Zero-value fields are omitted.
type AudienceListContactsParams struct {
	Limit int
}

func (p AudienceListContactsParams) toWire(startingAfter string) *oapi.ListAudienceContactsParams {
	w := &oapi.ListAudienceContactsParams{}
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

// AudienceAddContactsParams adds up to 1,000 existing contacts to a static
// audience. Adding a contact that is already a member has no effect; if any
// ID does not exist, the whole request fails and no contacts are added.
type AudienceAddContactsParams struct {
	ContactIDs []string
}

func (p AudienceAddContactsParams) toWire() oapi.AudienceContactsAddRequest {
	ids := make([]oapi.ContactID, len(p.ContactIDs))
	for i, id := range p.ContactIDs {
		ids[i] = oapi.ContactID(id)
	}
	return oapi.AudienceContactsAddRequest{ContactIds: ids}
}

// AudienceRemoveContactsParams removes up to 1,000 contacts from a static
// audience. Removing a contact that is not a member has no effect; if any ID
// does not exist, the whole request fails and no contacts are removed.
type AudienceRemoveContactsParams struct {
	ContactIDs []string
}

func (p AudienceRemoveContactsParams) toWire() oapi.AudienceContactsRemoveRequest {
	ids := make([]oapi.ContactID, len(p.ContactIDs))
	for i, id := range p.ContactIDs {
		ids[i] = oapi.ContactID(id)
	}
	return oapi.AudienceContactsRemoveRequest{ContactIds: ids}
}

// Create creates a static audience. Retried safely: a single idempotency key
// is reused across attempts. Provide your own key with
// option.WithIdempotencyKey.
func (s *AudiencesService) Create(ctx context.Context, params AudienceCreateParams, opts ...option.RequestOption) (*Audience, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateAudienceParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateAudience(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Audience
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single audience by id.
func (s *AudiencesService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*Audience, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetAudience(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Audience
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update edits an audience. Only the fields set in params change. Retried
// safely with a reused idempotency key.
func (s *AudiencesService) Update(ctx context.Context, id string, params AudienceUpdateParams, opts ...option.RequestOption) (*Audience, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UpdateAudienceParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateAudience(ctx, id, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Audience
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes an audience. The contacts themselves are not deleted.
func (s *AudiencesService) Delete(ctx context.Context, id string, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DeleteAudienceParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DeleteAudience(ctx, id, p, s.client.callEditors(cfg)...)
	})
	return err
}

// ListPage fetches one page of audiences. Pass the previous page's
// NextCursor as startingAfter to advance; "" starts from the most recent.
func (s *AudiencesService) ListPage(ctx context.Context, params AudienceListParams, startingAfter string, opts ...option.RequestOption) (*AudienceList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListAudiences(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out AudienceList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every audience matching params, fetching pages lazily. Range
// over it; the second value is non-nil only on the iteration where a fetch
// failed.
func (s *AudiencesService) List(ctx context.Context, params AudienceListParams, opts ...option.RequestOption) iter.Seq2[*Audience, error] {
	return paginate(func(cursor string) ([]Audience, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}

// ListContactsPage fetches one page of an audience's contacts, ordered by
// join time, most recent first. Pass the previous page's NextCursor as
// startingAfter to advance; "" starts from the most recent.
func (s *AudiencesService) ListContactsPage(ctx context.Context, audienceID string, params AudienceListContactsParams, startingAfter string, opts ...option.RequestOption) (*AudienceMemberList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListAudienceContacts(ctx, audienceID, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out AudienceMemberList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListContacts walks every contact in an audience, fetching pages lazily.
// Range over it; the second value is non-nil only on the iteration where a
// fetch failed.
func (s *AudiencesService) ListContacts(ctx context.Context, audienceID string, params AudienceListContactsParams, opts ...option.RequestOption) iter.Seq2[*AudienceMember, error] {
	return paginate(func(cursor string) ([]AudienceMember, *string, error) {
		page, err := s.ListContactsPage(ctx, audienceID, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}

// AddContacts adds up to 1,000 existing contacts to a static audience.
// Retried safely with a reused idempotency key.
func (s *AudiencesService) AddContacts(ctx context.Context, audienceID string, params AudienceAddContactsParams, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	wire := params.toWire()
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.AssignAudienceContactsParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.AssignAudienceContacts(ctx, audienceID, p, wire, s.client.callEditors(cfg)...)
	})
	return err
}

// RemoveContacts removes up to 1,000 contacts from a static audience. The
// contacts themselves are not deleted. Retried safely with a reused
// idempotency key.
func (s *AudiencesService) RemoveContacts(ctx context.Context, audienceID string, params AudienceRemoveContactsParams, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	wire := params.toWire()
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UnassignAudienceContactsParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UnassignAudienceContacts(ctx, audienceID, p, wire, s.client.callEditors(cfg)...)
	})
	return err
}

// RemoveContact removes one contact's membership in an audience. The
// contact itself is not deleted and remains a member of any other
// audiences. Retried safely with a reused idempotency key.
func (s *AudiencesService) RemoveContact(ctx context.Context, audienceID, contactID string, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UnassignAudienceContactParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UnassignAudienceContact(ctx, audienceID, contactID, p, s.client.callEditors(cfg)...)
	})
	return err
}
