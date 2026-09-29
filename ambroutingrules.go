package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

type AmbRoutingRulesCreateParams struct {
	BusinessAccountID string                  `json:"business_account_id"`
	MatchKind         AMBRoutingRuleMatchKind `json:"match_kind"`
	MatchIntentID     Nullable[string]        `json:"match_intent_id,omitempty"`
	MatchGroupID      Nullable[string]        `json:"match_group_id,omitempty"`
	Queue             string                  `json:"queue"`
	Precedence        *int                    `json:"precedence,omitempty"`
	IsDefault         *bool                   `json:"is_default,omitempty"`
}

func (s *AmbRoutingRulesService) Create(ctx context.Context, params AmbRoutingRulesCreateParams, opts ...option.RequestOption) (*AMBRoutingRule, error) {
	wire, err := ambWire[oapi.AMBRoutingRuleCreate](params)
	if err != nil {
		return nil, err
	}
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreateAMBRoutingRuleParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateAMBRoutingRule(ctx, op, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out AMBRoutingRule
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type AmbRoutingRulesUpdateParams struct {
	Queue      *string `json:"queue,omitempty"`
	Precedence *int    `json:"precedence,omitempty"`
	IsDefault  *bool   `json:"is_default,omitempty"`
}

func (s *AmbRoutingRulesService) Update(ctx context.Context, routingRuleId string, params AmbRoutingRulesUpdateParams, opts ...option.RequestOption) (*AMBRoutingRule, error) {
	wire, err := ambWire[oapi.AMBRoutingRuleUpdate](params)
	if err != nil {
		return nil, err
	}
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.UpdateAMBRoutingRuleParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateAMBRoutingRule(ctx, oapi.AMBRoutingRuleID(routingRuleId), op, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out AMBRoutingRule
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
