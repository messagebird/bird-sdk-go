package bird

type EmailInboxInsightsService struct {
	resource
	Benchmarks       *EmailInboxInsightsBenchmarksService
	DomainMonitoring *EmailInboxInsightsDomainMonitoringService
	Domains          *EmailInboxInsightsDomainsService
}

type EmailInboxInsightsBenchmarksService struct {
	resource
}

type EmailInboxInsightsDomainMonitoringService struct {
	resource
}

type EmailInboxInsightsDomainsService struct {
	resource
}
