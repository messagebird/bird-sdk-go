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

	verification, err := client.Verify.Verifications.Create(context.Background(), bird.VerifyVerificationsCreateParams{
		To: bird.VerificationTo{EmailAddress: bird.Email("user@example.com")},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(verification.Id, *verification.Status)
}
