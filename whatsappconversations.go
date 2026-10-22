package bird

// WhatsappConversationsService reads the workspace's WhatsApp conversations and
// triages them: assignment, labels, inbox status and read state. Who holds a
// conversation, the Business Agent or your workspace, is changed in the
// dashboard inbox, not here. Reach it via Client.Whatsapp.Conversations.
type WhatsappConversationsService struct{ resource }
