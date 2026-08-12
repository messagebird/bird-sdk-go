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

	msg, err := client.Whatsapp.Send(context.Background(), bird.WhatsappSendParams{
		To:       "+15551234567",
		Template: "bird_delivery_update",
		Components: []bird.WhatsAppMessageTemplateComponent{{
			Type: "body",
			Parameters: &[]bird.WhatsAppMessageTemplateComponentParameter{
				{Type: "text", Name: bird.String("ref"), Text: bird.String("A1B2C3D4")},
				{Type: "text", Name: bird.String("date"), Text: bird.String("10 Jul 2026")},
			},
		}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}
