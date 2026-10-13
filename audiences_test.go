package bird_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

// Conformance drivers discard decoded responses.
func TestAudiencesMembershipKeepsUnavailableDistinctFromZero(t *testing.T) {
	t.Parallel()
	membership := func(body string) *bird.AudienceMembershipResponse {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}))
		defer server.Close()
		client, err := bird.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("bk_eu1_conformance"))
		if err != nil {
			t.Fatal(err)
		}
		result, err := client.Audiences.Membership(context.Background(), "adn_01krdgeqcxet5s7t44vh8rt201")
		if err != nil {
			t.Fatal(err)
		}
		return result
	}

	if got := membership(`{}`).Membership; got != nil {
		t.Errorf("unavailable membership: got %+v, want absent", *got)
	}
	got := membership(`{"membership":{"total":0,"data_as_of":"2026-10-12T09:41:07.512843Z"}}`).Membership
	want := time.Date(2026, 10, 12, 9, 41, 7, 512843000, time.UTC)
	if got == nil || got.Total != 0 || !got.DataAsOf.Equal(want) {
		t.Errorf("empty audience: got %+v, want total 0 as of %s", got, want)
	}
}
