package bird

import "github.com/messagebird/bird-sdk-go/internal/oapi"

type (
	VoiceTrunk                   = oapi.VoiceTrunk
	VoiceTrunkList               = oapi.VoiceTrunkList
	VoiceTrunkSortField          = oapi.VoiceTrunkSortField
	VoiceNumber                  = oapi.VoiceNumber
	VoiceNumberList              = oapi.VoiceNumberList
	VoiceNumberProviderType      = oapi.VoiceNumberProviderType
	VoiceNumberSortField         = oapi.VoiceNumberSortField
	VoiceCallRoute               = oapi.VoiceCallRoute
	VoiceCallRouteType           = oapi.VoiceCallRouteType
	VoiceCallRouteWritable       = oapi.VoiceCallRouteWritable
	VoiceCallRouteReject         = oapi.VoiceCallRouteReject
	VoiceCallRouteTrunk          = oapi.VoiceCallRouteTrunk
	VoiceCallRouteForward        = oapi.VoiceCallRouteForward
	VoiceVerifiedNumber          = oapi.VoiceVerifiedNumber
	VoiceVerifiedNumberList      = oapi.VoiceVerifiedNumberList
	VoiceVerifiedNumberSortField = oapi.VoiceVerifiedNumberSortField
	VoiceDestination             = oapi.VoiceDestination
	VoiceDestinationList         = oapi.VoiceDestinationList
	VoiceSessionCredential       = oapi.VoiceSessionCredential
	VoiceSettings                = oapi.VoiceSettings
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
}

type VoiceLegsService struct{ resource }

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
type VoiceCallsService struct{ resource }
