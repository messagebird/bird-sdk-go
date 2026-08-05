package bird_test

import (
	"context"
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
