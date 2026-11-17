package bird_test

import (
	"net/http"
	"strings"
	"testing"
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
