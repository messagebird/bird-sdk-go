package bird

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

type AmbService struct {
	resource
	BusinessAccounts *AmbBusinessAccountsService
	Conversations    *AmbConversationsService
	RoutingRules     *AmbRoutingRulesService
	Stats            *AmbStatsService
	Suppressions     *AmbSuppressionsService
}

type AmbBusinessAccountsService struct {
	resource
	Events      *AmbBusinessAccountsEventsService
	Settings    *AmbBusinessAccountsSettingsService
	Submissions *AmbBusinessAccountsSubmissionsService
}

type AmbConversationsService struct {
	resource
}

type AmbRoutingRulesService struct {
	resource
}

type AmbStatsService struct {
	resource
	Conversations *AmbStatsConversationsService
	Inbound       *AmbStatsInboundService
}

type AmbSuppressionsService struct {
	resource
}

type AmbBusinessAccountsSettingsService struct {
	resource
}

type AmbBusinessAccountsEventsService struct {
	resource
}

type AmbBusinessAccountsSubmissionsService struct {
	resource
}

type AmbStatsConversationsService struct {
	resource
}

type AmbStatsInboundService struct {
	resource
}

type AmbSendParams struct {
	From     string
	To       string
	Content  map[string]any
	Source   *AMBMessageSource
	Category *string
	Metadata map[string]any
	Tags     []Tag
	Group    *string
	Intent   *string
	Locale   *string
}

func (s *AmbService) Send(ctx context.Context, params AmbSendParams, opts ...option.RequestOption) (*AMBMessage, error) {
	content, err := json.Marshal(params.Content)
	if err != nil {
		return nil, err
	}
	wire := oapi.AMBMessageSendRequest{From: params.From, To: params.To, Source: params.Source, Category: params.Category, Group: params.Group, Intent: params.Intent, Locale: params.Locale}
	if err := json.Unmarshal(content, &wire.Content); err != nil {
		return nil, err
	}
	if params.Metadata != nil {
		wire.Metadata = &params.Metadata
	}
	if params.Tags != nil {
		wire.Tags = &params.Tags
	}
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		p := &oapi.CreateAMBMessageParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateAMBMessage(ctx, p, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out AMBMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func ambWire[T any](params any) (T, error) {
	var wire T
	data, err := json.Marshal(params)
	if err != nil {
		return wire, err
	}
	err = json.Unmarshal(data, &wire)
	return wire, err
}
