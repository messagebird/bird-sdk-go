package bird_test

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func TestAMBWebhookPayload(t *testing.T) {
	t.Parallel()

	client, err := bird.NewClient(option.WithWebhookSecret(webhookSecret))
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"type":"amb.conversation_closed","timestamp":"2026-09-26T10:00:00Z","data":{"workspace_id":"ws_01krdgeqcxet5s7t44vh8rt9mg","business_account_id":"abz_01krdgeqcxet5s7t44vh8rt9mg","conversation_id":"acv_01krdgeqcxet5s7t44vh8rt9mg","open_count":2}}`)
	id := "msg_amb_conversation_closed"
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	headers := http.Header{}
	headers.Set("webhook-id", id)
	headers.Set("webhook-timestamp", ts)
	headers.Set("webhook-signature", sign(t, id, ts, payload))

	event, err := client.Webhooks.Unwrap(payload, headers)
	if err != nil {
		t.Fatal(err)
	}
	value, err := event.AsAny()
	if err != nil {
		t.Fatal(err)
	}
	closed, ok := value.(bird.AMBConversationClosedEvent)
	if !ok {
		t.Fatalf("AsAny returned %T, want AMBConversationClosedEvent", value)
	}
	if event.Type() != bird.EventTypeAmbConversationClosed || closed.Data.ConversationId != "acv_01krdgeqcxet5s7t44vh8rt9mg" || closed.Data.OpenCount != 2 {
		t.Fatalf("unexpected AMB conversation event: %#v", closed)
	}
}
