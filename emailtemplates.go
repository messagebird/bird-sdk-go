package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// EmailTemplatesService manages the workspace's reusable email templates and
// Bird's built-in catalogue. Reach it via Client.Email.Templates.
//
// Create is hand-written because its body carries the draft's initial content as
// a language-tag map the facade generator drops.
type EmailTemplatesService struct {
	resource

	// Versions manages a template's drafts and published versions.
	Versions *EmailTemplatesVersionsService

	// Broadcasts lists the broadcasts that block deleting a template.
	Broadcasts *EmailTemplatesBroadcastsService
}

// EmailTemplatesCreateParams is a template to create, together with its first
// draft. Nil optional fields are omitted from the request.
type EmailTemplatesCreateParams struct {
	// The template's workspace-unique handle, and a stable alternative to the template ID when sending by template. It can contain lowercase letters, numbers, hyphens, and underscores. It is fixed at creation, so pick it deliberately. Two prefixes are rejected: `bird_`, reserved for our built-in templates, and `emt_`, the template ID format, which a slug could never be distinguished from.
	Slug string
	// The template's display name, shown wherever the template is listed. You can change it any time. It defaults to the slug if you do not set one.
	Name *string
	// What the template is for, in your own words.
	Description *string
	// Whether the template is for `transactional` email or `marketing` email.
	Category EmailTemplateCategory
	// The authoring format to create the template in. `html` is finished markup you provide, optionally personalized with Liquid; a format the API cannot author yet is refused.
	Source EmailTemplateSourceWrite
	// The initial draft's content, keyed by language tag in BCP-47 form such as `en` or `pt-BR`. A template holds up to 25 languages, and a send picks one of them. Leave it empty to create an empty draft and add content later.
	Languages map[string]EmailTemplateLanguageContent
	// The language a send uses when it does not name one, and the last resort when a requested language is not available. It has to be one of the languages you supply. If you leave it out, we default to `en`, unless you supply exactly one language, in which case we use that one instead. So if you supply two or more languages and `en` is not among them, you have to set this yourself.
	DefaultLanguage *string
	// What a send does when it asks for a language this template does not carry. Defaults to `fallback` on email.
	OnMissingLanguage TemplateOnMissingLanguage
	// LanguageSourceRequired is a pointer because the server default is false — a
	// nil leaves the default, true rejects a send that names no language.
	LanguageSourceRequired *bool
}

func (p EmailTemplatesCreateParams) toWire() oapi.EmailTemplateCreate {
	body := oapi.EmailTemplateCreate{
		Slug:                   p.Slug,
		Name:                   p.Name,
		Description:            p.Description,
		Category:               p.Category,
		Source:                 p.Source,
		DefaultLanguage:        p.DefaultLanguage,
		LanguageSourceRequired: p.LanguageSourceRequired,
	}
	if len(p.Languages) > 0 {
		languages := p.Languages
		body.Languages = &languages
	}
	if p.OnMissingLanguage != "" {
		onMissing := p.OnMissingLanguage
		body.OnMissingLanguage = &onMissing
	}
	return body
}

// Create saves a template and its first editable draft, optionally with
// per-language content. The display name defaults to Slug, and a slug already
// used in the workspace is refused with a 409. Submit the draft before sending
// it. Retried safely with a reused idempotency key.
func (s *EmailTemplatesService) Create(ctx context.Context, params EmailTemplatesCreateParams, opts ...option.RequestOption) (*EmailTemplate, error) {
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreateEmailTemplateParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateEmailTemplate(ctx, op, params.toWire(), cfg...)
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
