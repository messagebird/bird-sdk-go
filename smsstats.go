package bird

// SmsStatsService reads aggregated statistics over the workspace's own SMS
// traffic. Reach it via Client.Sms.Stats. Every method is a read; each takes a
// params struct whose fields are all optional (zero values are omitted, and the
// server applies its own defaults for the window, sort, and limit).
type SmsStatsService struct {
	resource

	// Inbound reads the same shapes for messages the workspace's numbers received.
	Inbound *SmsStatsInboundService
}

// SmsStatsInboundService reads how many messages the workspace's own numbers
// received. Reach it via Client.Sms.Stats.Inbound.
type SmsStatsInboundService struct{ resource }
