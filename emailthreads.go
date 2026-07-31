package bird

// EmailThreadsService reads and manages email conversation threads stored in
// mailboxes. Reach it via Client.Email.Threads.
type EmailThreadsService struct {
	resource

	// Messages reads and replies to the messages in a conversation.
	Messages *EmailThreadsMessagesService
}
