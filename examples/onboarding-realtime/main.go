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
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials("your-app-key", "your-app-secret"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := client.Realtime.Publish(context.Background(), "rap_01krdgeqcxet5s7t44vh8rt9mg", bird.RealtimePublishParams{
		Event:    "order-updated",
		Channels: []string{"orders"},
		Data:     map[string]any{"id": 42, "status": "shipped"},
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("published to orders")
}
