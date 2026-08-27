package bird_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

// TestPreferencesCreateOmitsZeroFields pins toWire: Coverage/SenderScope/Source/
// ConsentedAt are optional and must be absent from the wire body at their zero
// value rather than sent as "" or the epoch.
func TestPreferencesCreateOmitsZeroFields(t *testing.T) {
	t.Parallel()

	var got map[string]any
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"applied":true}`))
	})

	_, err := newClient(t, server).Preferences.Create(context.Background(), bird.PreferencesCreateParams{
		Channel: bird.PreferenceChannelSms,
		Handle:  "+15550001234",
		Status:  bird.PreferenceStatusRevoked,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, field := range []string{"coverage", "sender_scope", "source", "consented_at"} {
		if _, ok := got[field]; ok {
			t.Errorf("%s present at zero value: %v", field, got[field])
		}
	}
	if got["channel"] != "sms" || got["handle"] != "+15550001234" || got["status"] != "revoked" {
		t.Errorf("required fields wrong: %v", got)
	}
}

// TestPreferencesDeleteRefusal locks in the reason Delete is hand-written: a
// refused delete is a normal 200 response, not an error, and its
// applied:false plus the surviving Preference must reach the caller rather
// than being swallowed by a void return.
func TestPreferencesDeleteRefusal(t *testing.T) {
	t.Parallel()

	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"applied": false,
			"preference": {
				"id": "prf_01krdgeqcxet5s7t44vh8rt9mg",
				"channel": "sms",
				"handle": "+15550001234",
				"status": "revoked"
			}
		}`))
	})

	result, err := newClient(t, server).Preferences.Delete(context.Background(), "prf_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if result == nil {
		t.Fatal("Delete returned a nil result alongside a nil error")
	}
	if result.Applied == nil || *result.Applied {
		t.Errorf("Applied = %v, want false", result.Applied)
	}
	if result.Preference == nil || result.Preference.Id != "prf_01krdgeqcxet5s7t44vh8rt9mg" {
		t.Errorf("Preference = %v, want the surviving statement", result.Preference)
	}
}
