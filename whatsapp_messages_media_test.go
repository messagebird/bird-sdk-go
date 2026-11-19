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
	api, seen := redirectingAPI(t, http.StatusOK, http.Header{"Content-Type": {"image/png"}})

	var meta bird.Response
	media, err := newClient(t, api, option.WithResponseInto(&meta)).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
	if err != nil {
		t.Fatalf("Media: %v", err)
	}
	if media.ContentType != "image/png" || string(media.Data) != string(mediaPNG) || media.ContentLength != int64(len(mediaPNG)) {
		t.Errorf("got %q/%v/%d, want image/png and %d bytes", media.ContentType, media.Data, media.ContentLength, len(mediaPNG))
	}
	if meta.Status != http.StatusFound {
		t.Errorf("Status = %d, want 302", meta.Status)
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

func TestMediaResponsesAndErrors(t *testing.T) {
	t.Parallel()

	t.Run("direct 2xx", func(t *testing.T) {
		api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(mediaPNG)
		})
		media, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
		if err != nil || media.ContentType != "image/png" || string(media.Data) != string(mediaPNG) {
			t.Fatalf("err=%v media=%+v", err, media)
		}
	})

	t.Run("octet-stream fallback", func(t *testing.T) {
		api, _ := redirectingAPI(t, http.StatusOK, nil)
		media, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
		if err != nil || media.ContentType != "application/octet-stream" {
			t.Fatalf("err=%v ContentType=%q", err, media.ContentType)
		}
	})

	t.Run("storage refusal names recovery", func(t *testing.T) {
		api, _ := redirectingAPI(t, http.StatusForbidden, nil)
		_, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
		var apiErr *bird.APIError
		if err == nil || !strings.Contains(err.Error(), "Media again") || errors.As(err, &apiErr) {
			t.Fatalf("err = %v (%T)", err, err)
		}
	})

	t.Run("redirect without Location", func(t *testing.T) {
		api := newServer(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusFound) })
		if _, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1"); err == nil || !strings.Contains(err.Error(), "Location") {
			t.Fatalf("err = %v, want missing Location", err)
		}
	})

	t.Run("API leg surfaces APIError", func(t *testing.T) {
		api := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			_, _ = w.Write([]byte(`{"error":{"type":"not_found_error","code":"E00404"}}`))
		})
		_, err := newClient(t, api).Whatsapp.Messages.Media(context.Background(), "wam_1", "waf_1")
		var apiErr *bird.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusGone {
			t.Fatalf("err = %v (%T)", err, err)
		}
	})
}
