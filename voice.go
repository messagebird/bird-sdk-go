package bird

type VoiceService struct {
	Legs *VoiceLegsService
}

type VoiceLegsService struct{ resource }
