package bird

// EmailTemplatesVersionsService manages a template's editable draft and the
// immutable versions submitting it produces. Reach it via
// Client.Email.Templates.Versions.
type EmailTemplatesVersionsService struct {
	resource

	// Languages authors the per-language content a version carries.
	Languages *EmailTemplatesVersionsLanguagesService
}
