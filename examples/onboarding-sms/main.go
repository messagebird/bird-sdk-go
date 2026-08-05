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

	msg, err := client.Sms.Send(context.Background(), bird.SmsSendParams{
		To:         "+15551234567",
		Template:   "bird_otp_verification",
		Parameters: map[string]any{"code": "493021"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}
