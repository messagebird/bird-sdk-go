package bird

// OrganizationService manages resources in the credential's organization.
type OrganizationService struct {
	Workspaces *OrganizationWorkspacesService
}

// OrganizationWorkspacesService manages the organization's workspaces.
type OrganizationWorkspacesService struct {
	resource

	// WhatsappNumbers connects numbers a multi-partner solution onboarded into a workspace.
	WhatsappNumbers *OrganizationWorkspacesWhatsappNumbersService
}

// OrganizationWorkspacesWhatsappNumbersService connects a number your own
// Embedded Signup onboarded through your multi-partner solution with Bird into
// one of your workspaces. Reach it via Client.Organization.Workspaces.WhatsappNumbers.
type OrganizationWorkspacesWhatsappNumbersService struct{ resource }
