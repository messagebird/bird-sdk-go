package bird_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

func TestEmailRecipientsRetainsAcceptedDeliveryContext(t *testing.T) {
	t.Parallel()
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"id":"er_01krdgeqcxet5s7t44vh8rt9mg","parent_id":"em_01krdgeqcxet5s7t44vh8rt9mg","recipient":"ada@example.net","role":"to","status":"canceled","subject":"Hello Ada","open_count":0,"click_count":0},{"id":"er_01krdgeqcxet5s7t44vh8rt9mh","parent_id":"em_01krdgeqcxet5s7t44vh8rt9mg","recipient":"charlie@example.net","role":"cc","status":"canceled","subject":"Hello Ada","copy_of_recipient_id":"er_01krdgeqcxet5s7t44vh8rt9mg","open_count":0,"click_count":0}],"next_cursor":null,"prev_cursor":null}`)
	})
	client := newClient(t, server)
	count := 0
	for recipient, err := range client.Email.Recipients(context.Background(), "em_01krdgeqcxet5s7t44vh8rt9mg", bird.EmailRecipientsParams{}) {
		if err != nil {
			t.Fatal(err)
		}
		count++
		if recipient.Subject == nil || *recipient.Subject != "Hello Ada" || recipient.Status == nil || string(*recipient.Status) != "canceled" {
			t.Fatalf("recipient lost its saved subject or cancellation: %#v", recipient)
		}
		if count == 1 && recipient.CopyOfRecipientId != nil {
			t.Fatalf("To recipient unexpectedly marked as a copy: %#v", recipient)
		}
		if count == 2 && (recipient.CopyOfRecipientId == nil || *recipient.CopyOfRecipientId != "er_01krdgeqcxet5s7t44vh8rt9mg") {
			t.Fatalf("copy lost its original To recipient: %#v", recipient)
		}
	}
	if count != 2 {
		t.Fatalf("got %d recipients, want 2", count)
	}
}
