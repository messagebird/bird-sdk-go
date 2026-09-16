package bird_test

import (
	"context"
	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestInsightsMonitoringRequiresExplicitValue(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client, err := bird.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("bk_eu1_conformance"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Email.InboxInsights.Domains.Update(context.Background(), "example.com", bird.EmailInboxInsightsDomainsUpdateParams{})
	if err == nil || result != nil || requests.Load() != 0 {
		t.Fatalf("missing monitored: result=%v err=%v requests=%d", result, err, requests.Load())
	}
}
