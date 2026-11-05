package bird_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
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

func TestNumberRouteRequiresExplicitValue(t *testing.T) {
	requests := make(chan map[string]any, 8)
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/voice/numbers/number-id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		requests <- body
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	client := newClient(t, server, option.WithMaxRetries(2))
	params := bird.VoiceNumbersUpdateParams{InboundConfiguration: &bird.VoiceInboundConfigurationPut{}}
	_, err := client.Voice.Numbers.Update(context.Background(), "number-id", params)
	var invalid *bird.RequestValidationError
	if !errors.As(err, &invalid) || invalid.Field != "VoiceInboundConfigurationPut.route" {
		t.Fatalf("unset route error = %v", err)
	}
	if _, ok := err.(*bird.RequestValidationError); !ok {
		t.Fatalf("local validation was classified as %T", err)
	}
	if len(requests) != 0 {
		t.Fatalf("unset route sent %d requests", len(requests))
	}
	var route bird.VoiceCallRouteWritable
	if err := json.Unmarshal([]byte(`{"type":"reject"}`), &route); err != nil {
		t.Fatal(err)
	}
	for _, state := range []struct {
		name  string
		route bird.Nullable[bird.VoiceCallRouteWritable]
		want  string
	}{
		{"null", bird.Null[bird.VoiceCallRouteWritable](), `{"inbound_configuration":{"route":null}}`},
		{"set", bird.Value(route), `{"inbound_configuration":{"route":{"type":"reject"}}}`},
	} {
		t.Run(state.name, func(t *testing.T) {
			params.InboundConfiguration.Route = state.route
			if _, err := client.Voice.Numbers.Update(context.Background(), "number-id", params); err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(<-requests)
			if err != nil || string(got) != state.want {
				t.Fatalf("wire body = %s, error = %v; want %s", got, err, state.want)
			}
		})
	}
}

func TestOtherRequiredNullableVoiceFields(t *testing.T) {
	requests := make(chan map[string]any, 8)
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		requests <- body
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	client := newClient(t, server, option.WithMaxRetries(2))
	for _, test := range []struct {
		name  string
		field string
		body  any
	}{
		{"workspace_limit", "VoiceDailySpendLimitUpdate.workspace_limit", &bird.VoiceDailySpendLimitUpdate{}},
		{"limit", "VoiceDestinationDailySpendLimitUpdate.limit", &bird.VoiceDestinationDailySpendLimitUpdate{}},
		{"orig", "VoiceSequenceSavedExecutionCallSample.orig", &bird.VoiceSequenceSavedExecutionCallSample{Id: "call", SessionId: "session", Dest: bird.Null[bird.VoiceSequenceSavedExecutionParty]()}},
		{"dest", "VoiceSequenceSavedExecutionCallSample.dest", &bird.VoiceSequenceSavedExecutionCallSample{Id: "call", SessionId: "session", Orig: bird.Null[bird.VoiceSequenceSavedExecutionParty]()}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var send func() error
			if limit, ok := test.body.(*bird.VoiceDailySpendLimitUpdate); ok {
				send = func() error {
					_, err := client.Voice.Settings.Update(context.Background(), bird.VoiceSettingsUpdateParams{DailySpendLimit: limit})
					return err
				}
			} else if limit, ok := test.body.(*bird.VoiceDestinationDailySpendLimitUpdate); ok {
				send = func() error {
					_, err := client.Voice.Destinations.Update(context.Background(), bird.VoiceDestinationsUpdateParams{Destinations: []bird.DestinationSetting{{CountryCode: "NL", DailySpendLimit: limit}}})
					return err
				}
			} else {
				var sequence bird.CreateVoiceCallSequenceRequest
				if err := json.Unmarshal([]byte(`{"entry_node_id":"start","trigger_data":{},"definition":{"schema_version":1,"nodes":[],"presentation":{"preview":{"execution_sample":{"id":"run","started_at":"2026-11-03T00:00:00Z","call":{"id":"call","session_id":"session","orig":null,"dest":null}},"trigger_data":{},"trigger_node_id":"start"}}}}`), &sequence); err != nil {
					t.Fatal(err)
				}
				sample := test.body.(*bird.VoiceSequenceSavedExecutionCallSample)
				send = func() error {
					sequence.Definition.Presentation.Preview.ExecutionSample.Call = *sample
					_, err := client.Voice.Calls.Create(context.Background(), bird.VoiceCallsCreateParams{From: "+31612345678", To: "+31687654321", Sequence: sequence})
					return err
				}
			}
			var invalid *bird.RequestValidationError
			if err := send(); !errors.As(err, &invalid) || invalid.Field != test.field {
				t.Fatalf("missing field error = %v", err)
			}
			if len(requests) != 0 {
				t.Fatal("missing field reached HTTP transport")
			}
			for _, value := range []string{"null", `{"amount":"1.00","currency_code":"EUR"}`} {
				if test.name == "orig" || test.name == "dest" {
					if value != "null" {
						value = `{"endpoint":{"type":"phone"},"address":"+31612345678"}`
					}
				}
				if err := json.Unmarshal([]byte(`{"`+test.name+`":`+value+`}`), test.body); err != nil {
					t.Fatal(err)
				}
				if err := send(); err != nil {
					t.Fatal(err)
				}
				body := <-requests
				encoded, err := json.Marshal(body)
				var want any
				if err := json.Unmarshal([]byte(value), &want); err != nil {
					t.Fatal(err)
				}
				canonical, _ := json.Marshal(want)
				if err != nil || !strings.Contains(string(encoded), `"`+test.name+`":`+string(canonical)) {
					t.Fatalf("wire body = %s, error = %v; missing %s", encoded, err, value)
				}
			}
		})
	}
}
