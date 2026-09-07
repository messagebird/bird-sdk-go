package bird

// EmailTemplatesService reads the email templates available to a workspace.
// Reach it via Client.Email.Templates. Read-only through this SDK: authoring
// stays on the dashboard, CLI and MCP surfaces.
type EmailTemplatesService struct{ resource }
