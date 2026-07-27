package bird

import (
	"context"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// ContactsService manages workspace contacts: create, read, update, delete, bulk
// upsert, and list. Reach it via Client.Contacts. Get, Delete, and List (with
// ListPage) are generated (contacts.gen.go); the writes are hand-written here
// because they carry craft the generator does not express (clearable PATCH
// fields, an ergonomic nested-contact batch).
type ContactsService struct{ resource }

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

// Update edits a contact. Only the fields set in params change. Retried
// safely with a reused idempotency key.
func (s *ContactsService) Update(ctx context.Context, id string, params ContactUpdateParams, opts ...option.RequestOption) (*Contact, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.UpdateContactParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateContact(ctx, oapi.ContactID(id), p, wire, cfg...)
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

// Batch creates or updates up to a batch's worth of contacts in one request,
// matched by email address. Retried safely with a reused idempotency key.
func (s *ContactsService) Batch(ctx context.Context, params ContactBatchParams, opts ...option.RequestOption) (*ContactUpsertResult, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.CreateContactBatchParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateContactBatch(ctx, p, wire, cfg...)
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

