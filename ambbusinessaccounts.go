package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

type AmbBusinessAccountsCreateParams struct {
	Name            string  `json:"name"`
	AppleBusinessID *string `json:"apple_business_id,omitempty"`
}

func (s *AmbBusinessAccountsService) Create(ctx context.Context, params AmbBusinessAccountsCreateParams, opts ...option.RequestOption) (*AMBBusinessAccount, error) {
	wire, err := ambWire[oapi.AMBBusinessAccountCreate](params)
	if err != nil {
		return nil, err
	}
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreateAMBBusinessAccountParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateAMBBusinessAccount(ctx, op, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out AMBBusinessAccount
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type AmbBusinessAccountsUpdateParams struct {
	Name            *string `json:"name,omitempty"`
	AppleBusinessID *string `json:"apple_business_id,omitempty"`
}

func (s *AmbBusinessAccountsService) Update(ctx context.Context, businessAccountID string, params AmbBusinessAccountsUpdateParams, opts ...option.RequestOption) (*AMBBusinessAccount, error) {
	wire, err := ambWire[oapi.AMBBusinessAccountUpdate](params)
	if err != nil {
		return nil, err
	}
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.UpdateAMBBusinessAccountParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateAMBBusinessAccount(ctx, oapi.AMBBusinessID(businessAccountID), op, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out AMBBusinessAccount
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
