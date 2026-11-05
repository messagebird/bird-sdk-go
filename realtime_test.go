package bird_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

// A Realtime call without app credentials must fail before the request is sent,
// so a misconfigured client never reaches the network with an unauthenticated
// publish.
func TestRealtimeWithoutCredentials(t *testing.T) {
	var hits atomic.Int32
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	client := newClient(t, server)

	_, err := client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
		Event: "ping", Channels: []string{"orders"},
	})
	if err == nil {
		t.Fatal("Publish without credentials: got nil error, want a configuration error")
	}
	if !strings.Contains(err.Error(), "option.WithRealtimeCredentials") {
		t.Errorf("error should name the option to set: %v", err)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server saw %d requests; the call must fail before any network call", n)
	}
}

// Record requests at the HTTP client boundary so these cases exercise the real
// public facades and generated wire encoding without contacting a server.
type realtimePresenceRecorder struct {
	bodies []map[string]any
}

func (r *realtimePresenceRecorder) RoundTrip(request *http.Request) (*http.Response, error) {
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		return nil, err
	}
	if err := request.Body.Close(); err != nil {
		return nil, err
	}
	r.bodies = append(r.bodies, body)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"data":[]}`)),
		Request:    request,
	}, nil
}

func TestRealtimePublicMethodsPreserveDataPresence(t *testing.T) {
	methods := []struct {
		name string
		call func(*bird.Client, bird.Nullable[any]) error
	}{
		{"publish", func(client *bird.Client, data bird.Nullable[any]) error {
			_, err := client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
				Event: "changed", Channels: []string{"orders"}, Data: data,
			})
			return err
		}},
		{"publish_batch", func(client *bird.Client, data bird.Nullable[any]) error {
			_, err := client.Realtime.PublishBatch(context.Background(), "rap_123", bird.RealtimePublishBatchParams{
				Events: []bird.RealtimeBatchEventParams{{Event: "changed", Channel: "orders", Data: data}},
			})
			return err
		}},
		{"members.send", func(client *bird.Client, data bird.Nullable[any]) error {
			return client.Realtime.Members.Send(context.Background(), "rap_123", "member_123", bird.RealtimeMemberSendParams{
				Event: "changed", Data: data,
			})
		}},
	}
	states := []struct {
		name    string
		data    bird.Nullable[any]
		present bool
		want    string
	}{
		{"unset", nil, false, ""},
		{"null", bird.Null[any](), true, "null"},
		{"object with nested null", bird.Value[any](map[string]any{"items": []any{nil, map[string]any{"value": nil}}}), true, `{"items":[null,{"value":null}]}`},
		{"array", bird.Value[any]([]any{nil, "value"}), true, `[null,"value"]`},
		{"scalar", bird.Value[any]("value"), true, `"value"`},
		{"false", bird.Value[any](false), true, "false"},
		{"zero", bird.Value[any](0), true, "0"},
	}
	for _, method := range methods {
		for _, state := range states {
			t.Run(method.name+"/"+state.name, func(t *testing.T) {
				recorder := &realtimePresenceRecorder{}
				client, err := bird.NewClient(
					option.WithAPIKey("bk_eu1_test"),
					option.WithHTTPClient(&http.Client{Transport: recorder}),
					option.WithMaxRetries(0),
					option.WithRealtimeCredentials("rk_test", "rs_test"),
				)
				if err != nil {
					t.Fatal(err)
				}
				if err := method.call(client, state.data); err != nil {
					t.Fatal(err)
				}
				if len(recorder.bodies) != 1 {
					t.Fatalf("recorded %d requests, want one", len(recorder.bodies))
				}
				body := recorder.bodies[0]
				if method.name == "publish_batch" {
					events, ok := body["events"].([]any)
					if !ok || len(events) != 1 {
						t.Fatalf("batch events = %#v", body["events"])
					}
					body, ok = events[0].(map[string]any)
					if !ok {
						t.Fatalf("batch event = %#v", events[0])
					}
				}
				value, present := body["data"]
				if present != state.present {
					t.Fatalf("data present = %t, want %t; body = %#v", present, state.present, body)
				}
				if present {
					encoded, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					if string(encoded) != state.want {
						t.Fatalf("data = %s, want %s", encoded, state.want)
					}
				}
			})
		}
	}
}

// The credentials travel as X-Realtime-Key/X-Realtime-Secret, and a per-call
// WithRealtimeCredentials overrides the client's — one client, several apps.
// option.WithHeader cannot forge either header, since both are SDK-owned.
func TestRealtimeCredentialHeaders(t *testing.T) {
	var gotKey, gotSecret string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Realtime-Key")
		gotSecret = r.Header.Get("X-Realtime-Secret")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	client := newClient(t, server, option.WithRealtimeCredentials("rk_client", "rs_client"))

	_, err := client.Realtime.Channels.List(context.Background(), "rap_123", bird.RealtimeChannelListParams{},
		option.WithRealtimeCredentials("rk_call", "rs_call"),
		option.WithHeader("X-Realtime-Key", "rk_forged"))
	if err != nil {
		t.Fatalf("Channels.List: %v", err)
	}
	if gotKey != "rk_call" || gotSecret != "rs_call" {
		t.Errorf("credentials on the wire: got %q/%q, want rk_call/rs_call", gotKey, gotSecret)
	}
}

// The app credentials are scoped to Realtime calls. A client configured with them
// must not put the app secret on an unrelated request, where it would reach
// proxies and logs that have no business seeing it.
func TestRealtimeCredentialsNotSentOnOtherResources(t *testing.T) {
	var gotKey, gotSecret string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Realtime-Key")
		gotSecret = r.Header.Get("X-Realtime-Secret")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"em_1"}`))
	})
	client := newClient(t, server, option.WithRealtimeCredentials("rk_client", "rs_client"))

	_, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From: "hello@acme.com", To: []string{"customer@example.com"}, Subject: "hi", HTML: "<p>hi</p>",
	})
	if err != nil {
		t.Fatalf("Email.Send: %v", err)
	}
	if gotKey != "" || gotSecret != "" {
		t.Errorf("email send carried Realtime credentials: key=%q secret=%q", gotKey, gotSecret)
	}
}
