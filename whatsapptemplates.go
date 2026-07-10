package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/option"
)

// WhatsAppTemplatesService reads the WhatsApp templates available to a
// workspace — Bird's built-in templates and any the workspace authored. Reach
// it via Client.WhatsappTemplates. The catalogue is read-only through this SDK.
type WhatsAppTemplatesService struct{ client *Client }

// List returns the WhatsApp templates available to the workspace. The
// catalogue is small and returned in full — this list is not paginated.
func (s *WhatsAppTemplatesService) List(ctx context.Context, opts ...option.RequestOption) (*WhatsAppTemplateList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListWhatsAppTemplates(ctx, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out WhatsAppTemplateList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
