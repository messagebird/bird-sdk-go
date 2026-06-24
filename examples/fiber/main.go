package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	app.Post("/send", func(c *fiber.Ctx) error {
		msg, err := client.Email.Send(c.Context(), bird.EmailSendParams{
			From:    "onboarding@messagebird.dev",
			To:      []string{"delivered@messagebird.dev"},
			Subject: "Hello from Bird",
			HTML:    "<p>My first Bird email.</p>",
		})
		if err != nil {
			var apiErr *bird.APIError
			if errors.As(err, &apiErr) {
				return c.Status(apiErr.StatusCode).JSON(fiber.Map{"error": apiErr.Error()})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusAccepted).JSON(msg)
	})

	log.Fatal(app.Listen(":3000"))
}
