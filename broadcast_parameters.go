package bird

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

type BroadcastsParameterTemplate struct {
	ID         *string                     `json:"id,omitempty"`
	Language   Nullable[string]            `json:"language,omitempty"`
	Parameters *map[string]json.RawMessage `json:"parameters,omitempty"`
	// The complete per-language override map; a pointer to an empty map clears every override.
	ParameterOverrides *map[string]map[string]json.RawMessage `json:"parameter_overrides,omitempty"`
}

type BroadcastsValidateParams struct {
	Template Nullable[BroadcastsParameterTemplate] `json:"template,omitempty"`
	From     string                                `json:"from,omitempty"`
	ReplyTo  Nullable[[]string]                    `json:"reply_to,omitempty"`
	Headers  *map[string]string                    `json:"headers,omitempty"`
}

type BroadcastsPreviewParams struct {
	BroadcastsValidateParams
	ContactID string `json:"contact_id,omitempty"`
}

type BroadcastsTestParams struct {
	BroadcastsPreviewParams
	To                 string `json:"to"`
	PreviewFingerprint string `json:"preview_fingerprint"`
}

func (p BroadcastsValidateParams) body() map[string]any {
	body := map[string]any{}
	if p.Template.IsSpecified() {
		if p.Template.IsNull() {
			body["template"] = nil
		} else {
			selected := p.Template.MustGet()
			template := map[string]any{}
			if selected.ID != nil {
				template["id"] = *selected.ID
			}
			if selected.Language.IsSpecified() {
				if selected.Language.IsNull() {
					template["language"] = nil
				} else {
					template["language"] = selected.Language.MustGet()
				}
			}
			if selected.Parameters != nil {
				template["parameters"] = *emptyIfNil(selected.Parameters)
			}
			if selected.ParameterOverrides != nil {
				template["parameter_overrides"] = broadcastParameterOverridesWire(*selected.ParameterOverrides)
			}
			body["template"] = template
		}
	}
	if p.From != "" {
		body["from"] = addressInput(p.From)
	}
	if p.ReplyTo.IsSpecified() {
		if p.ReplyTo.IsNull() {
			body["reply_to"] = nil
		} else {
			addresses := addressInputs(p.ReplyTo.MustGet())
			if addresses == nil {
				addresses = []oapi.EmailAddressInput{}
			}
			body["reply_to"] = addresses
		}
	}
	if p.Headers != nil {
		body["headers"] = *emptyIfNil(p.Headers)
	}
	return body
}

func (s *BroadcastsService) Validate(ctx context.Context, broadcastID string, params BroadcastsValidateParams, opts ...option.RequestOption) (*EmailBroadcastValidationResult, error) {
	requestBody := params.body()
	raw, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	editors, err := s.client.credentialEditors(cfg, nil)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		op := &oapi.ValidateEmailBroadcastParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.ValidateEmailBroadcastWithBody(ctx, broadcastID, op, "application/json", bytes.NewReader(raw), editors...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailBroadcastValidationResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *BroadcastsService) Preview(ctx context.Context, broadcastID string, params BroadcastsPreviewParams, opts ...option.RequestOption) (*EmailBroadcastPreviewResult, error) {
	requestBody := params.body()
	if params.ContactID != "" {
		requestBody["contact_id"] = params.ContactID
	}
	raw, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	editors, err := s.client.credentialEditors(cfg, nil)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		op := &oapi.PreviewEmailBroadcastParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.PreviewEmailBroadcastWithBody(ctx, broadcastID, op, "application/json", bytes.NewReader(raw), editors...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailBroadcastPreviewResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *BroadcastsService) Test(ctx context.Context, broadcastID string, params BroadcastsTestParams, opts ...option.RequestOption) (*EmailMessage, error) {
	requestBody := params.body()
	if params.ContactID != "" {
		requestBody["contact_id"] = params.ContactID
	}
	requestBody["to"] = params.To
	requestBody["preview_fingerprint"] = params.PreviewFingerprint
	raw, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	if cfg.IdempotencyKey == "" {
		return nil, errors.New("bird: broadcasts.test requires a nonempty caller idempotency key")
	}
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	editors, err := s.client.credentialEditors(cfg, nil)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		op := &oapi.SendEmailBroadcastTestParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = idempotencyKey
		}
		return s.client.oapi.SendEmailBroadcastTestWithBody(ctx, broadcastID, op, "application/json", bytes.NewReader(raw), editors...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailMessage
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
