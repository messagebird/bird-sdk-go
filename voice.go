package bird

import "github.com/messagebird/bird-sdk-go/internal/oapi"

type (
	VoiceLegRecordings            = oapi.VoiceLegRecordings
	VoiceLegRecordingMedia        = oapi.VoiceLegRecordingMedia
	VoiceLegRecording             = oapi.VoiceLegRecording
	VoiceLegTranscript            = oapi.VoiceLegTranscript
	VoiceLegTranscriptCue         = oapi.VoiceLegTranscriptCue
	VoiceCallTranscript           = oapi.VoiceCallTranscript
	VoiceCallTranscriptCue        = oapi.VoiceCallTranscriptCue
	VoiceLegMediaState            = oapi.VoiceLegMediaState
	VoiceLegChannel               = oapi.VoiceLegChannel
	VoiceLegTranscriptState       = oapi.VoiceLegTranscriptState
	VoiceParty                    = oapi.VoiceParty
	VoicePartyEndpoint            = oapi.VoicePartyEndpoint
	VoicePartySIPEndpoint         = oapi.VoicePartySIPEndpoint
	VoicePartyBridgePSTNEndpoint  = oapi.VoicePartyBridgePSTNEndpoint
	VoicePartyBridgeSIPEndpoint   = oapi.VoicePartyBridgeSIPEndpoint
	VoiceInboundForwardAs         = oapi.VoiceInboundForwardAs
	VoiceTrunk                    = oapi.VoiceTrunk
	VoiceTrunkList                = oapi.VoiceTrunkList
	VoiceTrunkSortField           = oapi.VoiceTrunkSortField
	VoiceNumber                   = oapi.VoiceNumber
	VoiceNumberList               = oapi.VoiceNumberList
	VoiceNumberProviderType       = oapi.VoiceNumberProviderType
	VoiceNumberSortField          = oapi.VoiceNumberSortField
	VoiceCallRoute                = oapi.VoiceCallRoute
	VoiceCallRouteType            = oapi.VoiceCallRouteType
	VoiceCallRouteWritable        = oapi.VoiceCallRouteWritable
	VoiceCallRouteReject          = oapi.VoiceCallRouteReject
	VoiceCallRouteTrunk           = oapi.VoiceCallRouteTrunk
	VoiceCallRouteForward         = oapi.VoiceCallRouteForward
	VoiceCallRouteRingGroup       = oapi.VoiceCallRouteRingGroup
	VoiceLegInboundRouteRingGroup = oapi.VoiceLegInboundRouteRingGroup
	VoiceVerifiedNumber           = oapi.VoiceVerifiedNumber
	VoiceVerifiedNumberList       = oapi.VoiceVerifiedNumberList
	VoiceVerifiedNumberSortField  = oapi.VoiceVerifiedNumberSortField
	VoiceDestination              = oapi.VoiceDestination
	VoiceDestinationList          = oapi.VoiceDestinationList
	VoiceSessionCredential        = oapi.VoiceSessionCredential
	VoiceSettings                 = oapi.VoiceSettings
	VoiceRingGroup                = oapi.VoiceRingGroup
	VoiceRingGroupList            = oapi.VoiceRingGroupList
	VoiceRingGroupMember          = oapi.VoiceRingGroupMember
	VoiceRingGroupMemberList      = oapi.VoiceRingGroupMemberList
)

// A recording's audio is downloadable only while its media is available; open
// means the leg is still recording, so Audio answers 425 until it closes.
const (
	VoiceLegMediaStateOpen      VoiceLegMediaState = "open"
	VoiceLegMediaStateAvailable VoiceLegMediaState = "available"
	VoiceLegMediaStateFailed    VoiceLegMediaState = "failed"
	VoiceLegMediaStatePurged    VoiceLegMediaState = "purged"
)

const (
	VoiceLegChannelRx VoiceLegChannel = "rx"
	VoiceLegChannelTx VoiceLegChannel = "tx"
)

const (
	VoiceLegTranscriptStateAvailable VoiceLegTranscriptState = "available"
	VoiceLegTranscriptStateFailed    VoiceLegTranscriptState = "failed"
)

const (
	VoiceInboundForwardAsCallingNumber VoiceInboundForwardAs = "calling_number"
	VoiceInboundForwardAsDialedNumber  VoiceInboundForwardAs = "dialed_number"
)

type VoiceService struct {
	Legs               *VoiceLegsService
	Trunks             *VoiceTrunksService
	Numbers            *VoiceNumbersService
	Settings           *VoiceSettingsService
	VerifiedNumbers    *VoiceVerifiedNumbersService
	Destinations       *VoiceDestinationsService
	SessionCredentials *VoiceSessionCredentialsService
	Calls              *VoiceCallsService
	RingGroups         *VoiceRingGroupsService
}

type VoiceLegsService struct {
	resource
	Recordings *VoiceLegsRecordingsService
}

type VoiceTrunksService struct {
	resource
	Gateways *VoiceTrunksGatewaysService
}

type VoiceTrunksGatewaysService struct{ resource }

type VoiceNumbersService struct{ resource }

type VoiceSettingsService struct{ resource }

type VoiceVerifiedNumbersService struct{ resource }

type VoiceDestinationsService struct{ resource }

type VoiceSessionCredentialsService struct{ resource }

type VoiceCallsService struct {
	resource
	Recordings *VoiceCallsRecordingsService
}

type VoiceCallsRecordingsService struct{ resource }

type VoiceRingGroupsService struct {
	resource
	Members *VoiceRingGroupsMembersService
}

type VoiceRingGroupsMembersService struct{ resource }
