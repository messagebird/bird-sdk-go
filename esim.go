package bird

type EsimService struct {
	resource
	Assignment             *EsimAssignmentService
	Credentials            *EsimCredentialsService
	Deliveries             *EsimDeliveriesService
	InstallLinks           *EsimInstallLinksService
	Offers                 *EsimOffersService
	Orders                 *EsimOrdersService
	Packages               *EsimPackagesService
	RecurringSubscriptions *EsimRecurringSubscriptionsService
	Settings               *EsimSettingsService
	Subscribers            *EsimSubscribersService
	Zones                  *EsimZonesService
}

type EsimAssignmentService struct {
	resource
}

type EsimCredentialsService struct {
	resource
}

type EsimDeliveriesService struct {
	resource
}

type EsimInstallLinksService struct {
	resource
}

type EsimOffersService struct {
	resource
}

type EsimOrdersService struct {
	resource
}

type EsimPackagesService struct {
	resource
}

type EsimRecurringSubscriptionsService struct {
	resource
}

type EsimSettingsService struct {
	resource
}

type EsimSubscribersService struct {
	resource
}

type EsimZonesService struct {
	resource
}
