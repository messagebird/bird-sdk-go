package bird

type VoiceService struct {
	Legs  *VoiceLegsService
	Calls *VoiceCallsService
}

type VoiceLegsService struct{ resource }

type VoiceCallsService struct{ resource }
