package bird_test

import (
	"context"
	"net/http"
	"testing"
)

// TestPreferencesDeleteRefusal locks in the reason Delete is hand-written: a
// refused delete is a normal 200 response, not an error, and its
// applied:false plus the surviving Preference must reach the caller rather
// than being swallowed by a void return.
func TestPreferencesDeleteRefusal(t *testing.T) {
	t.Parallel()

	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"applied":false,"preference":{"id":"prf_01krdgeqcxet5s7t44vh8rt9mg","channel":"sms","handle":"+15550001234","status":"revoked"}}`))
	})

	result, err := newClient(t, server).Preferences.Delete(context.Background(), "prf_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if result == nil || result.Applied == nil || *result.Applied || result.Preference == nil || result.Preference.Id != "prf_01krdgeqcxet5s7t44vh8rt9mg" {
		t.Fatalf("result = %+v, want applied=false with surviving preference", result)
	}
}
