package bird

import (
	"context"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// ContactPropertiesService manages workspace contact properties: create,
// read, update, list, archive, and unarchive. Reach it via
// Client.ContactProperties.
type ContactPropertiesService struct{ client *Client }

// ContactPropertyCreateParams registers a contact property. Key and Type are
// required; Type cannot be changed after creation.
type ContactPropertyCreateParams struct {
	Key           string // required; lowercase letters, digits, underscores, starting with a letter
	Type          string // required; "string", "number", or "boolean"
	FallbackValue any    // optional; matches the declared Type
}

func (p ContactPropertyCreateParams) toWire() oapi.ContactPropertyCreateRequest {
	return oapi.ContactPropertyCreateRequest{
		Key:           p.Key,
		Type:          oapi.ContactPropertyCreateRequestType(p.Type),
		FallbackValue: p.FallbackValue,
	}
}

// ContactPropertyUpdateParams changes a contact property's fallback value.
type ContactPropertyUpdateParams struct {
	FallbackValue any // optional; nil leaves the fallback unchanged
}

func (p ContactPropertyUpdateParams) toWire() oapi.ContactPropertyUpdateRequest {
	return oapi.ContactPropertyUpdateRequest{FallbackValue: p.FallbackValue}
}

// ContactPropertyListParams filters the contact property list. Zero-value
// fields are omitted.
type ContactPropertyListParams struct {
	Limit int
}

func (p ContactPropertyListParams) toWire(startingAfter string) *oapi.ListContactPropertiesParams {
	w := &oapi.ListContactPropertiesParams{}
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

// Create registers a contact property. Retried safely: a single idempotency
// key is reused across attempts. Provide your own key with
// option.WithIdempotencyKey.
func (s *ContactPropertiesService) Create(ctx context.Context, params ContactPropertyCreateParams, opts ...option.RequestOption) (*ContactProperty, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateContactPropertyParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateContactProperty(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactProperty
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single contact property by id.
func (s *ContactPropertiesService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*ContactProperty, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetContactProperty(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactProperty
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes a contact property's fallback value. Retried safely with a
// reused idempotency key.
func (s *ContactPropertiesService) Update(ctx context.Context, id string, params ContactPropertyUpdateParams, opts ...option.RequestOption) (*ContactProperty, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UpdateContactPropertyParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateContactProperty(ctx, id, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactProperty
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Archive archives a contact property: the key stops being accepted in new
// contact writes and stops rendering in templates, but every value already
// stored on contacts is preserved. Reverse it with Unarchive. Retried safely
// with a reused idempotency key.
func (s *ContactPropertiesService) Archive(ctx context.Context, id string, opts ...option.RequestOption) (*ContactProperty, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.ArchiveContactPropertyParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.ArchiveContactProperty(ctx, id, p, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactProperty
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Unarchive reactivates an archived contact property: the key is accepted in
// contact writes and renders in templates again. Retried safely with a
// reused idempotency key.
func (s *ContactPropertiesService) Unarchive(ctx context.Context, id string, opts ...option.RequestOption) (*ContactProperty, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UnarchiveContactPropertyParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UnarchiveContactProperty(ctx, id, p, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactProperty
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPage fetches one page of contact properties. Pass the previous page's
// NextCursor as startingAfter to advance; "" starts from the most recent.
func (s *ContactPropertiesService) ListPage(ctx context.Context, params ContactPropertyListParams, startingAfter string, opts ...option.RequestOption) (*ContactPropertyList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListContactProperties(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out ContactPropertyList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every contact property matching params, fetching pages lazily.
// Range over it; the second value is non-nil only on the iteration where a
// fetch failed.
func (s *ContactPropertiesService) List(ctx context.Context, params ContactPropertyListParams, opts ...option.RequestOption) iter.Seq2[*ContactProperty, error] {
	return paginate(func(cursor string) ([]ContactProperty, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
