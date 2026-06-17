package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.POST("/send", func(c *gin.Context) {
		msg, err := client.Email.Send(c.Request.Context(), bird.EmailSendParams{
			From:    "onboarding@messagebird.dev",
			To:      []string{"delivered@messagebird.dev"},
			Subject: "Hello from Bird",
			HTML:    "<p>My first Bird email.</p>",
		})
		if err != nil {
			var apiErr *bird.APIError
			if errors.As(err, &apiErr) {
				c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, msg)
	})

	log.Fatal(r.Run(":3000"))
}
