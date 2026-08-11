package bird

// VoiceService reads a workspace's call log — the record Bird writes for every
// call, in flight or settled. Reach it via Client.Voice. Calls are placed by the
// customer's own SIP equipment rather than through the API, so this is a read
// surface with no send verb.
type VoiceService struct{ resource }
