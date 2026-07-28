package bird

// AudiencesService manages static audiences and their contact membership:
// create, read, update, delete, list, and add/remove contacts. Reach it via
// Client.Audiences.
type AudiencesService struct{ resource }
