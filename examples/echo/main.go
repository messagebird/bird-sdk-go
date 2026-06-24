package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.POST("/send", func(c echo.Context) error {
		msg, err := client.Email.Send(c.Request().Context(), bird.EmailSendParams{
			From:    "onboarding@messagebird.dev",
			To:      []string{"delivered@messagebird.dev"},
			Subject: "Hello from Bird",
			HTML:    "<p>My first Bird email.</p>",
		})
		if err != nil {
			var apiErr *bird.APIError
			if errors.As(err, &apiErr) {
				return c.JSON(apiErr.StatusCode, map[string]string{"error": apiErr.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusAccepted, msg)
	})

	log.Fatal(e.Start(":3000"))
}
