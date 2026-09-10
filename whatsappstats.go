package bird

// WhatsappStatsService reads aggregated statistics over the workspace's own
// WhatsApp traffic. Reach it via Client.Whatsapp.Stats. Every method is a read;
// each takes a params struct whose fields are all optional (zero values are
// omitted, and the server applies its own defaults for the window and limit).
type WhatsappStatsService struct {
	resource

	// Inbound reads the same shapes for messages the workspace's numbers received.
	Inbound *WhatsappStatsInboundService
}

// WhatsappStatsInboundService reads how many messages the workspace's own
// numbers received. Reach it via Client.Whatsapp.Stats.Inbound.
type WhatsappStatsInboundService struct{ resource }
