package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// ContactPropertiesService manages workspace contact properties: create,
// read, update, list, archive, and unarchive. Reach it via
// Client.ContactProperties.
type ContactPropertiesService struct{ resource }

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

// Create registers a contact property. Hand-written: FallbackValue is an
// arbitrary JSON value matching the declared Type, which the generator can't
// model. Retried safely: a single idempotency key is reused across attempts.
func (s *ContactPropertiesService) Create(ctx context.Context, params ContactPropertyCreateParams, opts ...option.RequestOption) (*ContactProperty, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.CreateContactPropertyParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateContactProperty(ctx, p, wire, cfg...)
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

// Update changes a contact property's fallback value. Hand-written for the same
// arbitrary-value reason as Create. Retried safely with a reused idempotency key.
func (s *ContactPropertiesService) Update(ctx context.Context, id string, params ContactPropertyUpdateParams, opts ...option.RequestOption) (*ContactProperty, error) {
	wire := params.toWire()
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.UpdateContactPropertyParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateContactProperty(ctx, oapi.ContactPropertyID(id), p, wire, cfg...)
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
