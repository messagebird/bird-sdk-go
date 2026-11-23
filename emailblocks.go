package bird

// EmailBlocksService manages reusable email blocks: self-contained sections a
// template links instead of copying, filled with the block's latest published
// content when mail is sent. Reach it via Client.Email.Blocks.
type EmailBlocksService struct {
	resource

	// Languages edits a block's per-language drafts.
	Languages *EmailBlocksLanguagesService
}

// EmailBlocksLanguagesService edits the draft each block language keeps.
// Reach it via Client.Email.Blocks.Languages.
type EmailBlocksLanguagesService struct {
	resource

	// Revisions reads a block language's published revisions.
	Revisions *EmailBlocksLanguagesRevisionsService
}

// EmailBlocksLanguagesRevisionsService reads the revisions each publication of
// a block language adds. Reach it via Client.Email.Blocks.Languages.Revisions.
type EmailBlocksLanguagesRevisionsService struct{ resource }
