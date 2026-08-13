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

	result, err := client.Verify.Verifications.Check(context.Background(), bird.VerifyVerificationsCheckParams{
		To:   bird.VerificationTo{Email: bird.Email("user@example.com")},
		Code: "123456",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*result.Success)
}
