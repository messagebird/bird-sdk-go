package bird

import "github.com/messagebird/bird-sdk-go/internal/oapi"

type (
	VoiceTrunk             = oapi.VoiceTrunk
	VoiceTrunkList         = oapi.VoiceTrunkList
	VoiceTrunkSortField    = oapi.VoiceTrunkSortField
	VoiceNumber            = oapi.VoiceNumber
	VoiceNumberList        = oapi.VoiceNumberList
	VoiceNumberSortField   = oapi.VoiceNumberSortField
	VoiceCallRoute         = oapi.VoiceCallRoute
	VoiceCallRouteWritable = oapi.VoiceCallRouteWritable
	VoiceCallRouteReject   = oapi.VoiceCallRouteReject
	VoiceCallRouteTrunk    = oapi.VoiceCallRouteTrunk
	VoiceCallRouteForward  = oapi.VoiceCallRouteForward
	VoiceCallerID          = oapi.VoiceCallerID
	VoiceCallerIDList      = oapi.VoiceCallerIDList
	VoiceCallerIDSortField = oapi.VoiceCallerIDSortField
	VoiceDestination       = oapi.VoiceDestination
	VoiceDestinationList   = oapi.VoiceDestinationList
	VoiceSessionCredential = oapi.VoiceSessionCredential
)

type VoiceService struct {
	Legs               *VoiceLegsService
	Trunks             *VoiceTrunksService
	Numbers            *VoiceNumbersService
	CallerIDs          *VoiceCallerIDsService
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

type VoiceCallerIDsService struct{ resource }

type VoiceDestinationsService struct{ resource }

type VoiceSessionCredentialsService struct{ resource }
type VoiceCallsService struct{ resource }
