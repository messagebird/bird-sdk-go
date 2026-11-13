package bird_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

func TestAudioFollowsTheRedirectWithoutCredentials(t *testing.T) {
	t.Parallel()
	api, seen := redirectingAPI(t, http.StatusOK, http.Header{"Content-Type": {"audio/wav"}})

	audio, err := newClient(t, api).Voice.Legs.Recordings.Audio(context.Background(), "vcl_1", "vrc_1")
	if err != nil {
		t.Fatalf("Audio: %v", err)
	}
	if audio.ContentType != "audio/wav" || string(audio.Data) != string(mediaPNG) || audio.ContentLength != int64(len(mediaPNG)) {
		t.Errorf("got %q/%v/%d, want audio/wav and the stored bytes", audio.ContentType, audio.Data, audio.ContentLength)
	}
	// The presigned URL carries its own credential and refuses a second auth
	// mechanism, so a Bird header reaching storage is both a leak and a broken request.
	if got := seen.Get("Authorization"); got != "" {
		t.Errorf("storage received Authorization %q, want none", got)
	}
}

// The conformance corpus cannot script a 302, so this is the branch the
// voice.legs.recordings.audio vector actually drives.
func TestAudioAcceptsADirect2xx(t *testing.T) {
	t.Parallel()
	api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "audio/wav")
		_, _ = w.Write(mediaPNG)
	})

	audio, err := newClient(t, api).Voice.Legs.Recordings.Audio(context.Background(), "vcl_1", "vrc_1")
	if err != nil {
		t.Fatalf("Audio: %v", err)
	}
	if audio.ContentType != "audio/wav" || string(audio.Data) != string(mediaPNG) {
		t.Errorf("got %q/%v, want audio/wav and the stored bytes", audio.ContentType, audio.Data)
	}
}

func TestAudioRefusedLinkNamesTheRecovery(t *testing.T) {
	t.Parallel()
	api, _ := redirectingAPI(t, http.StatusForbidden, nil)

	_, err := newClient(t, api).Voice.Legs.Recordings.Audio(context.Background(), "vcl_1", "vrc_1")
	if err == nil || !strings.Contains(err.Error(), "Audio again") {
		t.Fatalf("err = %v, want one naming the recovery", err)
	}
	var apiErr *bird.APIError
	if errors.As(err, &apiErr) {
		t.Errorf("err = %T, want a connection error rather than an API error", err)
	}
}

// A leg that is still recording is the API's own 425, which the caller needs
// as an API error to tell "wait" from "gone".
func TestAudioSurfacesStillRecording(t *testing.T) {
	t.Parallel()
	api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooEarly)
		_, _ = w.Write([]byte(`{"error":{"type":"too_early_error","code":"E21083"}}`))
	})

	_, err := newClient(t, api).Voice.Legs.Recordings.Audio(context.Background(), "vcl_1", "vrc_1")
	var apiErr *bird.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v (%T), want *bird.APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusTooEarly || apiErr.Type != bird.ErrorTypeTooEarly || apiErr.Code != "E21083" {
		t.Errorf("got %d %s %s, want 425 too_early_error E21083", apiErr.StatusCode, apiErr.Type, apiErr.Code)
	}
}
