package bird_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

var mediaPNG = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

func redirectingAPI(t *testing.T, storageStatus int, storageHeader http.Header) (*httptest.Server, *http.Header) {
	t.Helper()
	var seen http.Header
	storage := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		// A nil entry suppresses net/http's own content sniffing, which would
		// otherwise read the PNG magic bytes and declare image/png — leaving the
		// no-content-type case untestable.
		w.Header()["Content-Type"] = nil
		for name, values := range storageHeader {
			w.Header()[name] = values
		}
		w.WriteHeader(storageStatus)
		if storageStatus/100 == 2 {
			_, _ = w.Write(mediaPNG)
		}
	})
	api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", storage.URL+"/blob.png")
		w.WriteHeader(http.StatusFound)
	})
	return api, &seen
}

func TestMediaFollowsTheRedirect(t *testing.T) {
	t.Parallel()
	api, _ := redirectingAPI(t, http.StatusOK, http.Header{"Content-Type": {"image/png"}})

	media, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
	if err != nil {
		t.Fatalf("Media: %v", err)
	}
	if media.ContentType != "image/png" {
		t.Errorf("ContentType = %q, want image/png", media.ContentType)
	}
	if string(media.Data) != string(mediaPNG) {
		t.Errorf("Data = %v, want %v", media.Data, mediaPNG)
	}
	if media.ContentLength != int64(len(mediaPNG)) {
		t.Errorf("ContentLength = %d, want %d", media.ContentLength, len(mediaPNG))
	}
}

// The presigned URL carries its own credential and refuses a second auth
// mechanism, so a Bird header reaching storage is both a leak and a broken
// request. This is the assertion the whole two-leg design exists for.
func TestMediaSendsNoCredentialsToStorage(t *testing.T) {
	t.Parallel()
	api, seen := redirectingAPI(t, http.StatusOK, http.Header{"Content-Type": {"image/png"}})

	if _, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1"); err != nil {
		t.Fatalf("Media: %v", err)
	}
	if got := seen.Get("Authorization"); got != "" {
		t.Errorf("storage received Authorization %q, want none", got)
	}
	for name := range *seen {
		if strings.HasPrefix(strings.ToLower(name), "bird-") {
			t.Errorf("storage received %s, want no Bird-* header", name)
		}
	}
}

// The conformance corpus cannot script a 302 — vector.schema.json's scripted
// responses carry only status and body, no headers — so this is the branch the
// whatsapp.messages.media vector actually drives.
func TestMediaAcceptsADirect2xx(t *testing.T) {
	t.Parallel()
	api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(mediaPNG)
	})

	media, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
	if err != nil {
		t.Fatalf("Media: %v", err)
	}
	if media.ContentType != "image/png" || string(media.Data) != string(mediaPNG) {
		t.Errorf("got %q/%v, want image/png and the png bytes", media.ContentType, media.Data)
	}
}

func TestMediaFallsBackToOctetStream(t *testing.T) {
	t.Parallel()
	api, _ := redirectingAPI(t, http.StatusOK, nil)

	media, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
	if err != nil {
		t.Fatalf("Media: %v", err)
	}
	if media.ContentType != "application/octet-stream" {
		t.Errorf("ContentType = %q, want application/octet-stream", media.ContentType)
	}
}

func TestMediaRefusedLinkNamesTheRecovery(t *testing.T) {
	t.Parallel()
	api, _ := redirectingAPI(t, http.StatusForbidden, nil)

	_, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
	if err == nil || !strings.Contains(err.Error(), "Media again") {
		t.Fatalf("err = %v, want one naming the recovery", err)
	}
	// A storage refusal is not a Bird API failure: mapping it as one would
	// report the caller's own key as lacking permission.
	var apiErr *bird.APIError
	if errors.As(err, &apiErr) {
		t.Errorf("err = %T, want a connection error rather than an API error", err)
	}
}

func TestMediaRedirectWithoutLocation(t *testing.T) {
	t.Parallel()
	api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusFound)
	})

	_, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
	if err == nil || !strings.Contains(err.Error(), "Location") {
		t.Fatalf("err = %v, want one naming the missing Location header", err)
	}
}

// The API leg keeps the core's error mapping: an expired media object is a Bird
// 410, not a storage failure, and must not be flattened into one.
func TestMediaSurfacesAnAPIError(t *testing.T) {
	t.Parallel()
	api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"error":{"type":"not_found_error","code":"E00404"}}`))
	})

	_, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
	var apiErr *bird.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v (%T), want *bird.APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusGone {
		t.Errorf("StatusCode = %d, want 410", apiErr.StatusCode)
	}
}

// WithResponseInto is the caller's own handle on transport metadata; the
// redirect-aware read borrows the same field internally, so this pins that it
// still reaches the caller.
func TestMediaFillsCallerResponseInto(t *testing.T) {
	t.Parallel()
	api, _ := redirectingAPI(t, http.StatusOK, http.Header{"Content-Type": {"image/png"}})

	var meta bird.Response
	client := newClient(t, api, option.WithResponseInto(&meta))
	if _, err := client.Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1"); err != nil {
		t.Fatalf("Media: %v", err)
	}
	if meta.Status != http.StatusFound {
		t.Errorf("Status = %d, want 302", meta.Status)
	}
}
