package main

import (
	"context"
	"fmt"
	"log"
	"os"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	conversation, err := client.Amb.Conversations.Get(ctx, os.Getenv("AMB_CONVERSATION_ID"))
	if err != nil {
		log.Fatal(err)
	}
	business, err := client.Amb.BusinessAccounts.Get(ctx, conversation.BusinessAccountId)
	if err != nil {
		log.Fatal(err)
	}
	if business.Status == "disconnected" || conversation.Status != "open" || business.AppleBusinessId == nil || conversation.Recipient == nil || conversation.Recipient.OpaqueUserId == nil {
		log.Fatal("A configured, connected business account and an open conversation are required.")
	}
	message, err := client.Amb.Send(ctx, bird.AmbSendParams{
		From: business.AppleBusinessId.String(), To: *conversation.Recipient.OpaqueUserId,
		Content: map[string]any{"type": "text", "body": "Your order is ready."},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(message.Id)
	events, err := client.Amb.ListEvents(ctx, message.Id, bird.AmbListEventsParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(events)
}
