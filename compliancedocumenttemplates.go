package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// ComplianceDocumentTemplatesService downloads the blank forms registration
// requirements name. Reach it via Client.ComplianceDocumentTemplates.
type ComplianceDocumentTemplatesService struct{ resource }

// Get downloads the blank form a requirement's template_url names, as a PDF,
// Word document, JPEG or PNG. Complete it and upload the result with
// Client.ComplianceAttachments.Create; export a Word form to PDF first.
func (s *ComplianceDocumentTemplatesService) Get(ctx context.Context, templateID string, opts ...option.RequestOption) (*ComplianceFile, error) {
	body, meta, err := s.getFile(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetComplianceDocumentTemplate(ctx, oapi.ComplianceTemplateID(templateID), nil, cfg...)
	})
	if err != nil {
		return nil, err
	}
	return complianceFile(body, meta.Header), nil
}
