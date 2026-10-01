package bird

// WhatsappAgentsService reaches the WhatsApp Business Agent on one of the
// workspace's numbers. The agent itself is onboarded and configured in the
// dashboard; what is public is telling it about things that happen to your
// contacts. Reach it via Client.Whatsapp.Agents.
type WhatsappAgentsService struct {
	resource

	// Notifications tells a number's agent that something happened for one
	// contact, and reads back what came of it.
	Notifications *WhatsappAgentsNotificationsService
}

// WhatsappAgentsNotificationsService sends a number's agent notifications and
// reads them back. A notification settles asynchronously, so Create answers at
// status accepted and Get or List shows whether the agent acted on it. Reach it
// via Client.Whatsapp.Agents.Notifications.
type WhatsappAgentsNotificationsService struct{ resource }
