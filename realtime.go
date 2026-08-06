package bird

// RealtimeService publishes events to a Realtime app and inspects its live
// state. Reach it via Client.Realtime.
//
// Every call needs the app's own credentials on top of the workspace API key:
// configure them with option.WithRealtimeCredentials, at construction for a
// single app or per call when one client serves several apps. Without them a
// method fails before any request is sent.
//
// The app id is a positional argument rather than client config, so one client
// can address any app the workspace owns.
type RealtimeService struct {
	resource

	// Channels reads the app's occupied channels and their members.
	Channels *RealtimeChannelsService
	// Members acts on an app-defined member across all of its connections.
	Members *RealtimeMembersService
}

// RealtimeChannelsService reads channel state. Reach it via
// Client.Realtime.Channels.
type RealtimeChannelsService struct{ resource }

// RealtimeMembersService acts on the connections of one member. Reach it via
// Client.Realtime.Members.
type RealtimeMembersService struct{ resource }
