package bird

// WhatsappKeywordRulesService reads what a reply to one of the workspace's
// numbers does, and configures the workspace's own overrides of it. Bird's rules
// and the workspace's share one id space, so a read answers from either and
// Scope is what tells them apart. Reach it via Client.Whatsapp.KeywordRules.
type WhatsappKeywordRulesService struct{ resource }
