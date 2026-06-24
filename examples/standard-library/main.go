package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		msg, err := client.Email.Send(r.Context(), bird.EmailSendParams{
			From:    "onboarding@messagebird.dev",
			To:      []string{"delivered@messagebird.dev"},
			Subject: "Hello from Bird",
			HTML:    "<p>My first Bird email.</p>",
		})
		if err != nil {
			var apiErr *bird.APIError
			if errors.As(err, &apiErr) {
				http.Error(w, apiErr.Error(), apiErr.StatusCode)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(msg)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
