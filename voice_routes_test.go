package bird_test

import (
	"encoding/json"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

func TestRingGroupRoutesDecode(t *testing.T) {
	const groupID = "vrg_01krdgeqcxet5s7t44vh8rt9mg"
	var number bird.VoiceNumber
	var settings bird.VoiceSettings
	var leg bird.VoiceLeg
	for _, item := range []struct {
		body string
		out  any
	}{
		{`{"inbound_configuration":{"route":{"type":"ring_group","ring_group_id":"` + groupID + `"}}}`, &number},
		{`{"inbound_configuration":{"route":{"type":"ring_group","ring_group_id":"` + groupID + `"}}}`, &settings},
		{`{"route":{"type":"ring_group","ring_group_id":"` + groupID + `"}}`, &leg},
	} {
		if err := json.Unmarshal([]byte(item.body), item.out); err != nil {
			t.Fatal(err)
		}
	}
	if number.InboundConfiguration.Route == nil || leg.Route == nil {
		t.Fatal("ring group route was lost")
	}
	for _, route := range []bird.VoiceCallRoute{*number.InboundConfiguration.Route, settings.InboundConfiguration.Route} {
		value, err := route.ValueByDiscriminator()
		if err != nil {
			t.Fatal(err)
		}
		group, ok := value.(bird.VoiceCallRouteRingGroup)
		if !ok || group.Type != "ring_group" || group.RingGroupId != groupID {
			t.Fatalf("number route = %#v", value)
		}
	}
	value, err := leg.Route.ValueByDiscriminator()
	if err != nil {
		t.Fatal(err)
	}
	group, ok := value.(bird.VoiceLegInboundRouteRingGroup)
	if !ok || group.Type != "ring_group" || group.RingGroupId != groupID {
		t.Fatalf("leg route = %#v", value)
	}
}

func TestPublicWritableRouteDoesNotAdmitSequence(t *testing.T) {
	var route bird.VoiceCallRouteWritable
	if err := json.Unmarshal([]byte(`{"type":"sequence"}`), &route); err != nil {
		t.Fatal(err)
	}
	if _, err := route.ValueByDiscriminator(); err == nil {
		t.Fatal("public writable route unexpectedly admitted sequence")
	}
}
