package bird_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

// Example constructs a client and sends an email. The region is taken from the
// API key's prefix; pass option.WithBaseURL or option.WithRegion to override.
func Example() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "onboarding@messagebird.dev",
		To:      []string{"delivered@messagebird.dev"},
		Subject: "Hello from Bird",
		HTML:    "<p>My first Bird email.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id)
}

func ExampleEmailService_Send() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "onboarding@messagebird.dev",
		To:      []string{"delivered@messagebird.dev"},
		Subject: "Hello from Bird",
		HTML:    "<p>My first Bird email.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// A richer send: cc/bcc, reply-to, tags, metadata, opt-out of click tracking,
// and an idempotency key (safe to retry — the server dedupes).
func ExampleEmailService_Send_rich() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Email.Send(context.Background(), bird.EmailSendParams{
		From:        "hello@acme.com",
		To:          []string{"a@example.com", "b@example.com"},
		Cc:          []string{"manager@example.com"},
		ReplyTo:     []string{"support@acme.com"},
		Subject:     "Your March invoice",
		HTML:        "<p>Attached.</p>",
		Tags:        []bird.EmailTag{{Name: "category", Value: "billing"}},
		Metadata:    map[string]any{"invoice_id": "inv_123"},
		TrackClicks: bird.Bool(false),
	}, option.WithIdempotencyKey("invoice-march/cust_1"))
	if err != nil {
		log.Fatal(err)
	}
}

// Send with display names: "Name <addr>" syntax in From and To.
func ExampleEmailService_Send_displayNames() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "Bird Support <support@acme.com>",
		To:      []string{"Jane Doe <jane@example.com>", "bob@example.com"},
		Subject: "Your order is confirmed",
		HTML:    "<p>Thanks for your order!</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
}

// SendBatch queues several emails in one request and returns one result item
// per message, in submission order.
func ExampleEmailService_SendBatch() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	batch, err := client.Email.SendBatch(context.Background(), bird.EmailSendBatchParams{
		Messages: []bird.EmailSendParams{
			{
				From:    "onboarding@messagebird.dev",
				To:      []string{"alice@example.com"},
				Subject: "Hello, Alice",
				HTML:    "<p>Welcome!</p>",
			},
			{
				From:    "onboarding@messagebird.dev",
				To:      []string{"bob@example.com"},
				Subject: "Hello, Bob",
				HTML:    "<p>Welcome!</p>",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range batch.Data {
		fmt.Println(item.Id)
	}
}

// Branch on the typed error hierarchy. The SDK already retries transient
// failures (timeouts, 429, 5xx), so a returned error is terminal — most callers
// just propagate it; branch only to act on a category.
func ExampleEmailService_Send_errors() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Email.Send(context.Background(), bird.EmailSendParams{
		From: "onboarding@messagebird.dev", To: []string{"delivered@messagebird.dev"}, Subject: "Hello from Bird", HTML: "<p>My first Bird email.</p>",
	})
	if err != nil {
		var rle *bird.RateLimitError
		var ve *bird.ValidationError
		var ae *bird.APIError
		switch {
		case errors.As(err, &rle):
			fmt.Println("rate limited; retry after", rle.RetryAfter)
		case errors.As(err, &ve):
			for _, d := range ve.Details {
				fmt.Printf("%s: %s\n", d.Param, d.Message)
			}
		case errors.As(err, &ae):
			fmt.Printf("API error %s (status %d, request %s)\n", ae.Code, ae.StatusCode, ae.RequestID)
		default:
			log.Print(err) // transport: *bird.ConnectionError or *bird.TimeoutError
		}
	}
}

func ExampleEmailService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Get(context.Background(), "em_abc123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*msg.Status, *msg.DeliveredCount)
}

// List auto-paginates: it lazily fetches each page and yields every matching
// message across all of them.
func ExampleEmailService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for msg, err := range client.Email.List(context.Background(), bird.EmailListParams{Status: bird.EmailStatusBounced}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg.Id)
	}
	page, err := client.Email.ListPage(context.Background(), bird.EmailListParams{}, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(page.Data)) // page.NextCursor carries the next starting_after
}

// EmailDefaults set common send fields once; a per-send value always wins.
func ExampleEmailDefaults() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithEmailDefaults(bird.EmailDefaults{
			From:     "hello@acme.com",
			Category: bird.CategoryTransactional,
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	// From is filled from the default.
	if _, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		To: []string{"customer@example.com"}, Subject: "Hi", HTML: "<p>hi</p>",
	}); err != nil {
		log.Fatal(err)
	}
}

// Unwrap verifies the Standard Webhooks signature over the raw request body and
// returns a typed event to dispatch on.
func ExampleWebhookService_Unwrap() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithWebhookSecret(os.Getenv("BIRD_WEBHOOK_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/webhooks/bird", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		event, err := client.Webhooks.Unwrap(body, r.Header)
		if err != nil {
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent) // ack fast, then process

		payload, _ := event.AsAny()
		switch p := payload.(type) {
		case bird.EmailDeliveredEvent:
			fmt.Println("delivered:", p.Data.EmailId, p.Data.Recipient)
		case bird.EmailBouncedEvent:
			fmt.Println("bounced:", p.Type)
		}
	})
}

// The verb methods reach endpoints outside the curated surface, decoding the
// response into a value you provide.
func ExampleClient_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	var out struct {
		Data []struct {
			Recipient string `json:"recipient"`
		} `json:"data"`
	}
	if err := client.Get(context.Background(), "/v1/email/suppressions", &out); err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(out.Data))
}
