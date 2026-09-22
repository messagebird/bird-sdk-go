package bird

// WhatsappGroupsService creates and administers WhatsApp groups: one chat a
// business number shares with up to 8 people, who join by opening its invite
// link. Reach it via Client.Whatsapp.Groups.
type WhatsappGroupsService struct {
	resource

	// InviteLink rotates the link people join through.
	InviteLink *WhatsappGroupsInviteLinkService

	// JoinRequests decides who gets into a group that asks for approval.
	JoinRequests *WhatsappGroupsJoinRequestsService

	// Participants removes people from a group. Nothing adds them.
	Participants *WhatsappGroupsParticipantsService

	// Pins holds the up-to-3 messages kept at the top of the group's chat.
	Pins *WhatsappGroupsPinsService
}

// WhatsappGroupsInviteLinkService rotates a group's invite link, which is the
// only way anyone joins. Reach it via Client.Whatsapp.Groups.InviteLink.
type WhatsappGroupsInviteLinkService struct{ resource }

// WhatsappGroupsJoinRequestsService reads and decides the requests raised when
// someone opens the invite link of a group that asks for approval. Reach it via
// Client.Whatsapp.Groups.JoinRequests.
type WhatsappGroupsJoinRequestsService struct{ resource }

// WhatsappGroupsParticipantsService removes people from a group. There is no
// add: WhatsApp lets nobody be added directly. Reach it via
// Client.Whatsapp.Groups.Participants.
type WhatsappGroupsParticipantsService struct{ resource }

// WhatsappGroupsPinsService pins and unpins the messages kept at the top of a
// group's chat. Reach it via Client.Whatsapp.Groups.Pins.
type WhatsappGroupsPinsService struct{ resource }
