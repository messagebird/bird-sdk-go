package bird

// WhatsappReactionService places and removes this workspace's emoji reaction on
// a message a contact sent, and reads the log of every change to that message's
// reactions. Reach it via Client.Whatsapp.Reaction.
type WhatsappReactionService struct {
	resource
}
