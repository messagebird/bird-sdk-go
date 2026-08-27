package bird

// ContactsService manages workspace contacts: create, read, update, delete, bulk
// upsert, and list. Reach it via Client.Contacts.
type ContactsService struct {
	resource

	// Preferences reads a contact's own stated messaging preferences across
	// every channel.
	Preferences *ContactsPreferencesService
}
