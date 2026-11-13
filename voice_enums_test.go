package bird_test

import (
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

func wantStream(t *testing.T, state bird.VoiceLegMediaState, channel bird.VoiceLegChannel, transcript bird.VoiceLegTranscriptState) {
	t.Helper()
	if state != bird.VoiceLegMediaStateAvailable || channel != bird.VoiceLegChannelRx || transcript != bird.VoiceLegTranscriptStateFailed {
		t.Fatalf("got %s %s %s", state, channel, transcript)
	}
}

func forwardedAs(p *bird.VoiceParty) bird.VoiceInboundForwardAs {
	return bridgedAs(p.Endpoint.BridgePstn)
}

func bridgedAs(pstn *bird.VoicePartyBridgePSTNEndpoint) bird.VoiceInboundForwardAs {
	return pstn.ForwardAs
}

// Outside the module internal/oapi cannot be imported, so the recording
// inventory's and transcript's nested types are usable only through these
// bird names.
func TestVoiceRecordingTypesAreNameable(t *testing.T) {
	t.Parallel()
	party := bird.VoiceParty{Endpoint: &bird.VoicePartyEndpoint{
		BridgePstn: &bird.VoicePartyBridgePSTNEndpoint{ForwardAs: bird.VoiceInboundForwardAsDialedNumber},
		BridgeSip:  &bird.VoicePartyBridgeSIPEndpoint{},
		Sip:        &bird.VoicePartySIPEndpoint{},
	}}
	inv := bird.VoiceLegRecordings{
		Media:       []bird.VoiceLegRecordingMedia{{State: bird.VoiceLegMediaStateAvailable, SourceParty: &party}},
		Recordings:  []bird.VoiceLegRecording{{Channel: bird.VoiceLegChannelRx}},
		Transcripts: []bird.VoiceLegTranscript{{State: bird.VoiceLegTranscriptStateFailed, Channel: bird.VoiceLegChannelTx}},
	}
	wantStream(t, inv.Media[0].State, inv.Recordings[0].Channel, inv.Transcripts[0].State)

	transcript := bird.VoiceCallTranscript{Parties: []bird.VoiceParty{party}}
	if got := forwardedAs(&transcript.Parties[0]); got != bird.VoiceInboundForwardAsDialedNumber {
		t.Fatalf("forward_as = %s", got)
	}
	if got := forwardedAs(inv.Media[0].SourceParty); got != bird.VoiceInboundForwardAsDialedNumber {
		t.Fatalf("forward_as = %s", got)
	}
}
