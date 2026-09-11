package bird

// EmailTemplatesBroadcastsService lists the broadcasts that block deleting a
// template: the scheduled and accepted ones, which have not pinned their
// content yet. Reach it via Client.Email.Templates.Broadcasts.
type EmailTemplatesBroadcastsService struct{ resource }
