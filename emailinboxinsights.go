package bird

type EmailInboxInsightsService struct {
	resource
	SeedTests        *EmailInboxInsightsSeedTestsService
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

type EmailInboxInsightsSeedTestsService struct {
	resource
	Configuration *EmailInboxInsightsSeedTestsConfigurationService
}

type EmailInboxInsightsSeedTestsConfigurationService struct{ resource }
