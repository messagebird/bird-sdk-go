package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// SMSTemplatesService reads the SMS templates available to a workspace — Bird's
// built-in templates and any the workspace authored. Reach it via
// Client.SMSTemplates. The catalogue is read-only through this SDK.
type SMSTemplatesService struct{ client *Client }

// SMSTemplateScope filters templates by origin: Bird's built-in templates
// (system) or the workspace's own (workspace).
type SMSTemplateScope string

const (
	SMSTemplateScopeSystem    SMSTemplateScope = "system"
	SMSTemplateScopeWorkspace SMSTemplateScope = "workspace"
)

// SMSTemplateCategory filters templates by content classification.
type SMSTemplateCategory string

const (
	SMSTemplateCategoryTransactional  SMSTemplateCategory = "transactional"
	SMSTemplateCategoryMarketing      SMSTemplateCategory = "marketing"
	SMSTemplateCategoryAuthentication SMSTemplateCategory = "authentication"
	SMSTemplateCategoryService        SMSTemplateCategory = "service"
)

// SMSTemplateListParams filters the template list. Zero-value fields are omitted.
type SMSTemplateListParams struct {
	Scope    SMSTemplateScope    // filter by origin (system or workspace)
	Category SMSTemplateCategory // filter by content classification
	Language string              // keep only templates available in this BCP-47 language tag
}

func (p SMSTemplateListParams) toWire() *oapi.ListSMSTemplatesParams {
	w := &oapi.ListSMSTemplatesParams{}
	if p.Scope != "" {
		scope := oapi.ListSMSTemplatesParamsScope(p.Scope)
		w.Scope = &scope
	}
	if p.Category != "" {
		category := oapi.ListSMSTemplatesParamsCategory(p.Category)
		w.Category = &category
	}
	if p.Language != "" {
		language := p.Language
		w.Language = &language
	}
	return w
}

// List returns the SMS templates available to the workspace, filtered by params.
// The catalogue is small and returned in full — this list is not paginated.
func (s *SMSTemplatesService) List(ctx context.Context, params SMSTemplateListParams, opts ...option.RequestOption) (*SMSTemplateList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListSMSTemplates(ctx, params.toWire(), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out SMSTemplateList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single SMS template by its name or id, including its body and
// the variables it expects.
func (s *SMSTemplatesService) Get(ctx context.Context, templateRef string, opts ...option.RequestOption) (*SMSTemplate, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetSMSTemplate(ctx, templateRef, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out SMSTemplate
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
