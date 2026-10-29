package bird

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

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
	Templates        *AmbTemplatesService
}

// AmbTemplatesService reads the Apple Messages templates available to a
// workspace. Reach it via Client.Amb.Templates.
type AmbTemplatesService struct {
	resource

	// Versions reads the content and variables stored under one template.
	Versions *AmbTemplatesVersionsService
}

// AmbTemplatesVersionsService reads one Apple Messages template's versions.
// Reach it via Client.Amb.Templates.Versions.
type AmbTemplatesVersionsService struct {
	resource

	// Languages reads one version's language metadata and content.
	Languages *AmbTemplatesVersionsLanguagesService
}

// AmbTemplatesVersionsLanguagesService reads one Apple Messages template
// version's languages. Reach it via Client.Amb.Templates.Versions.Languages.
type AmbTemplatesVersionsLanguagesService struct{ resource }

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
	From string
	To   string
	// Text alone is a text message; beside RichLink or Interactive it is sent
	// first as its own bubble.
	Text        *AMBText
	RichLink    *AMBRichLink
	Interactive *AMBInteractive
	Template    string         // stored template id (amt_…) or slug, instead of Text, RichLink or Interactive
	Language    string         // template language as a BCP-47 tag; template sends only
	Parameters  map[string]any // template variable values; template sends only
	Source      *AMBMessageSource
	Category    *string
	Metadata    map[string]any
	Tags        []Tag
	Group       *string
	Intent      *string
	Locale      *string
}

func (s *AmbService) Send(ctx context.Context, params AmbSendParams, opts ...option.RequestOption) (*AMBMessage, error) {
	wire := oapi.AMBMessageSendRequest{From: params.From, To: params.To, Text: params.Text, RichLink: params.RichLink, Interactive: params.Interactive, Source: params.Source, Category: params.Category, Group: params.Group, Intent: params.Intent, Locale: params.Locale}
	if params.Template != "" {
		var tmpl oapi.AMBTemplateSend
		if params.Template != "" {
			if strings.HasPrefix(params.Template, "amt_") {
				id := oapi.AMBTemplateID(params.Template)
				tmpl.Id = &id
			} else {
				slug := params.Template
				tmpl.Slug = &slug
			}
		}
		if params.Language != "" {
			language := params.Language
			tmpl.Language = &language
		}
		if params.Parameters != nil {
			parameters := params.Parameters
			tmpl.Parameters = &parameters
		}
		wire.Template = &tmpl
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
