package bird

// EmailStatsService reads aggregated email statistics. Reach it via
// Client.Email.Stats. Every method is a read; each takes a params struct whose
// fields are all optional (zero values are omitted, and the server applies its
// own defaults for the window, sort, and limit).
type EmailStatsService struct{ resource }
