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

	http.HandleFunc("POST /send", func(w http.ResponseWriter, r *http.Request) {
		msg, err := client.Whatsapp.Send(r.Context(), bird.WhatsappSendParams{
			To:       "+15551234567",
			Template: "bird_otp",
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
