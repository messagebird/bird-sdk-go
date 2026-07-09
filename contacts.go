package bird

import (
	"context"
	"iter"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// ContactsService manages workspace contacts: create, read, update, delete, bulk
// upsert, and list. Reach it via Client.Contact.
type ContactsService struct{ client *Client }

// ContactCreateParams creates a contact. Email is required; the rest are
// optional and omitted from the request when left at their zero value.
type ContactCreateParams struct {
	Email      string         // required; unique within the workspace
	ExternalID string         // optional caller-supplied identifier, unique within the workspace when set
	FirstName  string         // optional
	LastName   string         // optional
	Data       map[string]any // optional custom property values, keyed by contact property name
}

func (p ContactCreateParams) toWire() oapi.ContactCreateRequest {
	body := oapi.ContactCreateRequest{Email: openapi_types.Email(p.Email)}
	if p.ExternalID != "" {
		externalID := p.ExternalID
		body.ExternalId = &externalID
	}
	if p.FirstName != "" {
		firstName := p.FirstName
		body.FirstName = &firstName
	}
	if p.LastName != "" {
		lastName := p.LastName
		body.LastName = &lastName
	}
	if len(p.Data) > 0 {
		data := map[string]interface{}(p.Data)
		body.Data = &data
	}
	return body
}

// ContactUpdateParams is a partial update of a contact. Every field is a
// pointer: nil leaves it unchanged; point at "" to clear a name or the
// external id. A key in Data set to nil removes that key from the contact's
// stored custom values; keys omitted from Data are left unchanged.
type ContactUpdateParams struct {
	Email      *string
	ExternalID *string
	FirstName  *string
	LastName   *string
	Data       map[string]any
}

func (p ContactUpdateParams) toWire() oapi.ContactUpdateRequest {
	body := oapi.ContactUpdateRequest{
		ExternalId: p.ExternalID,
		FirstName:  p.FirstName,
		LastName:   p.LastName,
	}
	if p.Email != nil {
		email := openapi_types.Email(*p.Email)
		body.Email = &email
	}
	if len(p.Data) > 0 {
		data := map[string]interface{}(p.Data)
		body.Data = &data
	}
	return body
}

// ContactListParams filters the contact list. Zero-value fields are omitted.
type ContactListParams struct {
	Email      string // exact match; matches at most one contact
	ExternalID string // exact match; matches at most one contact
	Search     string // case-insensitive substring match against email
	Limit      int
}

func (p ContactListParams) toWire(startingAfter string) *oapi.ListContactsParams {
	w := &oapi.ListContactsParams{}
	if p.Email != "" {
		email := p.Email
		w.Email = &email
	}
	if p.ExternalID != "" {
		externalID := p.ExternalID
		w.ExternalId = &externalID
	}
	if p.Search != "" {
		search := p.Search
		w.Search = &search
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

// ContactBatchParams bulk-upserts contacts matched by email address: existing
// contacts are updated with the supplied fields, new ones are created.
// AudienceIDs and DataMode are omitted from the request when left at their
// zero value.
type ContactBatchParams struct {
	Contacts    []ContactCreateParams
	AudienceIDs []string // audiences every contact in the request is added to
	DataMode    string   // "merge" (default) or "replace"; how each contact's Data is applied to its existing stored values
}

func (p ContactBatchParams) toWire() oapi.ContactUpsertRequest {
	contacts := make([]oapi.ContactCreateRequest, len(p.Contacts))
	for i, c := range p.Contacts {
		contacts[i] = c.toWire()
	}
	body := oapi.ContactUpsertRequest{Contacts: contacts}
	if len(p.AudienceIDs) > 0 {
		ids := make([]oapi.AudienceID, len(p.AudienceIDs))
		for i, id := range p.AudienceIDs {
			ids[i] = oapi.AudienceID(id)
		}
		body.AudienceIds = &ids
	}
	if p.DataMode != "" {
		mode := oapi.ContactUpsertRequestDataMode(p.DataMode)
		body.DataMode = &mode
	}
	return body
}

// Create creates a contact. Retried safely: a single idempotency key is
// reused across attempts. Provide your own key with option.WithIdempotencyKey.
func (s *ContactsService) Create(ctx context.Context, params ContactCreateParams, opts ...option.RequestOption) (*Contact, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateContactParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateContact(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Contact
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single contact by id.
func (s *ContactsService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*Contact, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetContact(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Contact
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update edits a contact. Only the fields set in params change. Retried
// safely with a reused idempotency key.
func (s *ContactsService) Update(ctx context.Context, id string, params ContactUpdateParams, opts ...option.RequestOption) (*Contact, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UpdateContactParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateContact(ctx, id, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Contact
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a contact.
func (s *ContactsService) Delete(ctx context.Context, id string, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DeleteContactParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DeleteContact(ctx, id, p, s.client.callEditors(cfg)...)
	})
	return err
}

// Batch creates or updates up to a batch's worth of contacts in one request,
// matched by email address. Retried safely with a reused idempotency key.
func (s *ContactsService) Batch(ctx context.Context, params ContactBatchParams, opts ...option.RequestOption) (*ContactUpsertResult, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateContactBatchParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateContactBatch(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactUpsertResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPage fetches one page of contacts. Pass the previous page's NextCursor
// as startingAfter to advance; "" starts from the most recent.
func (s *ContactsService) ListPage(ctx context.Context, params ContactListParams, startingAfter string, opts ...option.RequestOption) (*ContactList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListContacts(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every contact matching params, fetching pages lazily. Range over
// it; the second value is non-nil only on the iteration where a fetch failed.
func (s *ContactsService) List(ctx context.Context, params ContactListParams, opts ...option.RequestOption) iter.Seq2[*Contact, error] {
	return paginate(func(cursor string) ([]Contact, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
