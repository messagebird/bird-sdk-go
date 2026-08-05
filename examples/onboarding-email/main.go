// The first send a customer makes from the dashboard's onboarding step. Unlike
// quickstart-email this carries the key inline: the dashboard fills it with the
// workspace's real key, so the placeholder is what a reader sees before it is
// substituted, not advice to hardcode a secret.
package main

import (
	"context"
	"fmt"
	"log"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey("bk_XXXXXXXXXXXXXXXXXXXXXXXX"))
	if err != nil {
		log.Fatal(err)
	}

	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "onboarding@messagebird.dev",
		To:      []string{"delivered@messagebird.dev"},
		Subject: "Hello World",
		HTML:    "<p>You made your <strong>first email fly</strong>. Congratulations!</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}
