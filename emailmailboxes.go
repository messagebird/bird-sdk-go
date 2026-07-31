package bird

// EmailMailboxesService manages agent mailboxes — durable inboxes on inbox.ai or
// your own domain that receive, store, and send email. Reach it via
// Client.Email.Mailboxes.
type EmailMailboxesService struct {
	resource

	// Messages sends new messages from the mailbox's own address.
	Messages *EmailMailboxesMessagesService

	// ReceiveRules manages per-sender allow/block rules on the mailbox.
	ReceiveRules *EmailMailboxesReceiveRulesService
}
