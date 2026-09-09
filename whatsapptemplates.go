package bird

// WhatsappTemplatesService reads the workspace's WhatsApp template registry:
// the templates a send can name, and what each one currently offers. Reach it
// via Client.Whatsapp.Templates.
type WhatsappTemplatesService struct {
	resource

	// Versions reads one template's submissions, where its content lives.
	Versions *WhatsappTemplatesVersionsService
}
