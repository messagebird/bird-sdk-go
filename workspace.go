package bird

// WorkspaceService reads the workspace the client's credential is scoped to.
// Reach it via Client.Workspace.
type WorkspaceService struct {
	resource

	// RateLimits reads the rate limits that apply to the workspace.
	RateLimits *WorkspaceRateLimitsService
	// Quotas reads the resource and value quotas that apply to the workspace.
	Quotas *WorkspaceQuotasService
}
