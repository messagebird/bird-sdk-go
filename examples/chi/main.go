package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Post("/send", func(w http.ResponseWriter, req *http.Request) {
		msg, err := client.Email.Send(req.Context(), bird.EmailSendParams{
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

	log.Fatal(http.ListenAndServe(":3000", r))
}
