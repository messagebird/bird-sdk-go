package bird

// SmsTemplatesVersionsService reads one SMS template's versions. Reach it via
// Client.SmsTemplates.Versions.
type SmsTemplatesVersionsService struct {
	resource

	// Languages reads one version's language metadata and text.
	Languages *SmsTemplatesVersionsLanguagesService
}

// SmsTemplatesVersionsLanguagesService reads one SMS template version's
// languages. Reach it via Client.SmsTemplates.Versions.Languages.
type SmsTemplatesVersionsLanguagesService struct{ resource }
