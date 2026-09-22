package bird

// WhatsappSuppressionsService reads and edits the addresses the workspace will
// not message on WhatsApp. A record is one period of suppression rather than a
// current state, so an address suppressed, ended and suppressed again has two
// of them. Reach it via Client.Whatsapp.Suppressions.
type WhatsappSuppressionsService struct{ resource }
