package bird

// WhatsappTemplatesVersionsService reads one template's versions. A version is
// where content lives; the template above it carries only the handle. Reach it
// via Client.Whatsapp.Templates.Versions.
type WhatsappTemplatesVersionsService struct {
	resource

	// Languages reads one version's per-language content and review outcome.
	Languages *WhatsappTemplatesVersionsLanguagesService
}

// WhatsappTemplatesVersionsLanguagesService reads one version's languages.
// Reach it via Client.Whatsapp.Templates.Versions.Languages.
type WhatsappTemplatesVersionsLanguagesService struct{ resource }
