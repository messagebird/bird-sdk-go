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

	// What is this number? The base lookup always answers the country, the
	// serving network and a coarse line type, and it always bills once.
	number, err := client.Lookup.PhoneNumber(ctx, bird.LookupPhoneNumberParams{
		PhoneNumber: "+31612345678",
		Type:        []bird.LookupProperty{"porting", "score"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*number.CountryCode, *number.LineType)

	// Each requested property is billed only when it is delivered, so read the
	// status before the value. Anything other than "ok" means "not answered".
	if number.Score != nil && *number.Score.Status == "ok" {
		fmt.Println("credibility", *number.Score.Value)
	}
	if number.Porting != nil && *number.Porting.Status == "ok" {
		fmt.Println("ported", *number.Porting.Ported)
	}

	// Is this address worth sending to? Result is the field to decide on;
	// DeliveryConfidence is always present and comparable, which is what makes
	// it safe to fall back on when a new verdict is added.
	address, err := client.Lookup.Email(ctx, bird.LookupEmailParams{
		Email: "aisha.khan@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*address.Result, *address.DeliveryConfidence)

	if *address.Result == "typo" {
		fmt.Println("did you mean", *address.DidYouMean)
	}
}
