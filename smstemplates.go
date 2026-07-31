package bird

// SmsTemplatesService reads the SMS templates available to a workspace — Bird's
// built-in templates and any the workspace authored. Reach it via
// Client.SmsTemplates. The catalogue is read-only through this SDK.
type SmsTemplatesService struct{ resource }
