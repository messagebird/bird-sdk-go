package bird

// OrganizationService manages resources in the credential's organization.
type OrganizationService struct {
	Workspaces *OrganizationWorkspacesService
}

// OrganizationWorkspacesService manages the organization's workspaces.
type OrganizationWorkspacesService struct {
	resource
}
