package bird

import (
	"context"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// EmailTemplatesService manages workspace email templates: create a template and
// its editable draft, edit the draft, publish immutable numbered versions, and
// read templates and versions back. Reach it via Client.EmailTemplates.
type EmailTemplatesService struct{ client *Client }

// EmailTemplateSource is the authoring format a template is written in, fixed at
// creation. `liquid` currently supports variable substitution only.
type EmailTemplateSource string

const (
	EmailTemplateSourceLiquid     EmailTemplateSource = "liquid"
	EmailTemplateSourceHandlebars EmailTemplateSource = "handlebars"
	EmailTemplateSourceHTML       EmailTemplateSource = "html"
)

// EmailTemplateCreateParams creates a template and its initial draft. Optional
// fields are omitted from the request when left at their zero value.
type EmailTemplateCreateParams struct {
	Name        string              // required; workspace-unique slug handle for send-by-template
	Category    Category            // transactional or marketing
	Source      EmailTemplateSource // authoring format, fixed at creation
	Description string              // optional description of the template's purpose
	Subject     string              // initial draft subject line
	HTML        string              // initial draft HTML body
	Text        string              // optional initial draft plain-text body
	BrandKitID  string              // optional brand kit to apply to the draft
}

func (p EmailTemplateCreateParams) toWire() oapi.EmailTemplateCreate {
	body := oapi.EmailTemplateCreate{
		Name:     p.Name,
		Category: oapi.EmailTemplateCategory(p.Category),
		Source:   oapi.EmailTemplateSource(p.Source),
	}
	if p.Description != "" {
		description := p.Description
		body.Description = &description
	}
	if p.Subject != "" {
		subject := p.Subject
		body.Subject = &subject
	}
	if p.HTML != "" {
		html := p.HTML
		body.Html = &html
	}
	if p.Text != "" {
		text := p.Text
		body.Text = &text
	}
	if p.BrandKitID != "" {
		brandKitID := p.BrandKitID
		body.BrandKitId = &brandKitID
	}
	return body
}

// EmailTemplateUpdateParams is a partial update of a template's metadata and its
// draft content. Revision is required — pass the draft revision you last read so
// concurrent edits are detected (a stale value returns a conflict). Every other
// field is a pointer: nil leaves it unchanged.
type EmailTemplateUpdateParams struct {
	Revision    int     // required: the draft revision you last read
	Name        *string // nil leaves the name unchanged
	Description *string
	Subject     *string
	HTML        *string
	Text        *string
	BrandKitID  *string
}

func (p EmailTemplateUpdateParams) toWire() oapi.EmailTemplateUpdate {
	return oapi.EmailTemplateUpdate{
		Revision:    p.Revision,
		Name:        p.Name,
		Description: p.Description,
		Subject:     p.Subject,
		Html:        p.HTML,
		Text:        p.Text,
		BrandKitId:  p.BrandKitID,
	}
}

// EmailTemplateListParams filters the template list. Zero-value fields are omitted.
type EmailTemplateListParams struct {
	Category Category            // filter by category
	Source   EmailTemplateSource // filter by authoring format
	Name     string              // filter by name prefix (case-insensitive)
	Limit    int
}

func (p EmailTemplateListParams) toWire(startingAfter string) *oapi.ListEmailTemplatesParams {
	w := &oapi.ListEmailTemplatesParams{}
	if p.Category != "" {
		category := oapi.EmailTemplateCategory(p.Category)
		w.Category = &category
	}
	if p.Source != "" {
		source := oapi.EmailTemplateSource(p.Source)
		w.Source = &source
	}
	if p.Name != "" {
		name := p.Name
		w.Name = &name
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

// Create creates a template and its initial editable draft. Retried safely: a
// single idempotency key is reused across attempts. Provide your own key with
// option.WithIdempotencyKey.
func (s *EmailTemplatesService) Create(ctx context.Context, params EmailTemplateCreateParams, opts ...option.RequestOption) (*EmailTemplate, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateEmailTemplateParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateEmailTemplate(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailTemplate
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single template with its current draft content.
func (s *EmailTemplatesService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*EmailTemplate, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetEmailTemplate(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailTemplate
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update edits a template's metadata and draft content. Only the fields set in
// params change. Retried safely with a reused idempotency key.
func (s *EmailTemplatesService) Update(ctx context.Context, id string, params EmailTemplateUpdateParams, opts ...option.RequestOption) (*EmailTemplate, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UpdateEmailTemplateParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateEmailTemplate(ctx, id, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailTemplate
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a template and all its versions. Its name becomes available for
// reuse in the workspace.
func (s *EmailTemplatesService) Delete(ctx context.Context, id string, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DeleteEmailTemplateParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DeleteEmailTemplate(ctx, id, p, s.client.callEditors(cfg)...)
	})
	return err
}

// Publish publishes the current draft as a new immutable, numbered version and
// makes it the live version used by sends. The draft stays editable. Returns the
// newly published version.
func (s *EmailTemplatesService) Publish(ctx context.Context, id string, opts ...option.RequestOption) (*EmailTemplateVersion, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.PublishEmailTemplateParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.PublishEmailTemplate(ctx, id, p, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailTemplateVersion
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListVersions returns every version of a template — the current draft plus all
// published versions — newest first, in Data. The list is not paginated.
func (s *EmailTemplatesService) ListVersions(ctx context.Context, id string, opts ...option.RequestOption) (*EmailTemplateVersionList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListEmailTemplateVersions(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailTemplateVersionList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetVersion returns a single version of a template.
func (s *EmailTemplatesService) GetVersion(ctx context.Context, id, versionID string, opts ...option.RequestOption) (*EmailTemplateVersion, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetEmailTemplateVersion(ctx, id, versionID, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailTemplateVersion
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPage fetches one page of templates. Pass the previous page's NextCursor as
// startingAfter to advance; "" starts from the most recent.
func (s *EmailTemplatesService) ListPage(ctx context.Context, params EmailTemplateListParams, startingAfter string, opts ...option.RequestOption) (*EmailTemplateList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListEmailTemplates(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailTemplateList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every template matching params, fetching pages lazily. Range over
// it; the second value is non-nil only on the iteration where a fetch failed.
func (s *EmailTemplatesService) List(ctx context.Context, params EmailTemplateListParams, opts ...option.RequestOption) iter.Seq2[*EmailTemplateSummary, error] {
	return paginate(func(cursor string) ([]EmailTemplateSummary, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
