package bird_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func TestSMSMediaDownload(t *testing.T) {
	t.Parallel()
	var storageHeaders http.Header
	storage := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		storageHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(mediaPNG)
	})
	api := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sms/messages/sms_1/media/smf_1" || r.Header.Get("Authorization") == "" {
			t.Errorf("unexpected API request: %s, authenticated=%v", r.URL.Path, r.Header.Get("Authorization") != "")
		}
		w.Header().Set("Location", storage.URL+"/photo")
		w.WriteHeader(http.StatusFound)
	})
	media, err := newClient(t, api).Sms.Messages.Media(t.Context(), "sms_1", "smf_1")
	if err != nil {
		t.Fatal(err)
	}
	if string(media.Data) != string(mediaPNG) || media.ContentType != "image/png" || media.ContentLength != int64(len(mediaPNG)) {
		t.Fatalf("unexpected media: %+v", media)
	}
	if storageHeaders.Get("Authorization") != "" || storageHeaders.Get("Cookie") != "" {
		t.Fatal("storage received credentials")
	}
	for name := range storageHeaders {
		if strings.HasPrefix(strings.ToLower(name), "bird-") {
			t.Errorf("storage received %s", name)
		}
	}
}

// A presigned URL's query string is its credential, so a storage hop must not
// learn the previous URL from Referer, which net/http forwards by default.
func TestSMSMediaStorageRedirectDropsReferer(t *testing.T) {
	t.Parallel()
	var referer string
	cdn := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		referer = r.Header.Get("Referer")
		_, _ = w.Write(mediaPNG)
	})
	storage := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cdn.URL+"/photo", http.StatusFound)
	})
	api := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", storage.URL+"/photo?X-Amz-Signature=secret")
		w.WriteHeader(http.StatusFound)
	})
	media, err := newClient(t, api).Sms.Messages.Media(t.Context(), "sms_1", "smf_1")
	if err != nil {
		t.Fatal(err)
	}
	if string(media.Data) != string(mediaPNG) {
		t.Fatalf("unexpected media: %q", media.Data)
	}
	if referer != "" {
		t.Fatalf("redirect target received Referer %q", referer)
	}
}

// The storage leg keeps the request's timeout, so a stalled storage host ends
// in a TimeoutError rather than holding a context.Background() call forever.
func TestSMSMediaStorageLegKeepsTimeout(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	storage := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})
	api := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", storage.URL+"/photo")
		w.WriteHeader(http.StatusFound)
	})
	start := time.Now()
	_, err := newClient(t, api).Sms.Messages.Media(t.Context(), "sms_1", "smf_1", option.WithTimeout(200*time.Millisecond))
	var timeout *bird.TimeoutError
	if !errors.As(err, &timeout) {
		t.Fatalf("want *TimeoutError, got %T: %v", err, err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("storage leg ran %s past a 200ms timeout", elapsed)
	}
}

type stampingTransport struct{}

func (stampingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("X-Caller-Auth", "secret")
	return http.DefaultTransport.RoundTrip(r)
}

// An injected client carries the caller's own middleware and redirect policy. The
// API leg runs on a copy of it, so the middleware applies there and the policy is
// left as the caller set it; the storage leg runs on neither, so a header the
// middleware stamps never reaches a presigned URL.
func TestSMSMediaInjectedClientStaysOnTheAPILeg(t *testing.T) {
	t.Parallel()
	var apiStamp, storageStamp string
	storage := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		storageStamp = r.Header.Get("X-Caller-Auth")
		_, _ = w.Write(mediaPNG)
	})
	api := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		apiStamp = r.Header.Get("X-Caller-Auth")
		w.Header().Set("Location", storage.URL+"/photo")
		w.WriteHeader(http.StatusFound)
	})
	errCallerPolicy := errors.New("caller policy")
	injected := &http.Client{
		Transport:     stampingTransport{},
		CheckRedirect: func(*http.Request, []*http.Request) error { return errCallerPolicy },
	}
	media, err := newClient(t, api, option.WithHTTPClient(injected)).Sms.Messages.Media(t.Context(), "sms_1", "smf_1")
	if err != nil {
		t.Fatal(err)
	}
	if string(media.Data) != string(mediaPNG) {
		t.Fatalf("unexpected media: %q", media.Data)
	}
	if apiStamp != "secret" {
		t.Fatal("API leg skipped the injected transport")
	}
	if storageStamp != "" {
		t.Fatal("storage leg ran through the injected transport")
	}
	if injected.CheckRedirect == nil || !errors.Is(injected.CheckRedirect(nil, nil), errCallerPolicy) {
		t.Fatal("caller's redirect policy was replaced")
	}
}

// A host that sends headers and then stalls mid-body hits the same deadline, and
// reports it the same way, however the interrupted read words its error.
func TestSMSMediaStalledBodyIsATimeout(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	storage := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write(mediaPNG[:4])
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})
	api := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", storage.URL+"/photo")
		w.WriteHeader(http.StatusFound)
	})
	_, err := newClient(t, api).Sms.Messages.Media(t.Context(), "sms_1", "smf_1", option.WithTimeout(200*time.Millisecond))
	var timeout *bird.TimeoutError
	if !errors.As(err, &timeout) {
		t.Fatalf("want *TimeoutError, got %T: %v", err, err)
	}
}
