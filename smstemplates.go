package bird

// SmsTemplatesService reads the SMS templates available to a workspace. Reach
// it via Client.SmsTemplates.
type SmsTemplatesService struct {
	resource

	// Versions reads the content and variables stored under one template.
	Versions *SmsTemplatesVersionsService
}
