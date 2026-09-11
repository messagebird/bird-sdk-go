package bird

// WhatsappNumbersService reads the WhatsApp numbers this workspace can send
// from, and the state WhatsApp reports for each. Reach it via
// Client.Whatsapp.Numbers.
type WhatsappNumbersService struct {
	resource

	// Profile reads the business profile WhatsApp shows to people a number messages.
	Profile *WhatsappNumbersProfileService
}

// WhatsappNumbersProfileService reads one number's business profile. Reach it
// via Client.Whatsapp.Numbers.Profile.
type WhatsappNumbersProfileService struct{ resource }
