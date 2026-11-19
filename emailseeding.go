package bird

// EmailSeedingService groups automatic email seeding: workspace settings, the
// trigger address, and the retained runs and sources it produces.
type EmailSeedingService struct {
	resource
	Settings *EmailSeedingSettingsService
	Address  *EmailSeedingAddressService
	Runs     *EmailSeedingRunsService
	Sources  *EmailSeedingSourcesService
}

type EmailSeedingSettingsService struct{ resource }

type EmailSeedingAddressService struct{ resource }

type EmailSeedingRunsService struct{ resource }

type EmailSeedingSourcesService struct{ resource }
