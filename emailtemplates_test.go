package bird_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

func TestEmailTemplateWritePresence(t *testing.T) {
	t.Parallel()
	for _, present := range []bool{false, true} {
		name := "omitted"
		if present {
			name = "explicit zero and empty"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var bodies []map[string]any
			server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				bodies = append(bodies, body)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{}`)
			})
			client := newClient(t, server)
			var revision *int
			var html, displayName, defaultLanguage *string
			if present {
				revision = bird.Ptr(0)
				html = bird.Ptr("")
				displayName = bird.Ptr("Welcome")
				defaultLanguage = bird.Ptr("en")
			}
			ctx := context.Background()
			_, err := client.Email.Templates.Create(ctx, bird.EmailTemplatesCreateParams{
				Slug: "welcome", Category: "transactional", Source: "html", Description: html, Name: displayName, DefaultLanguage: defaultLanguage,
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Email.Templates.Versions.Languages.Set(ctx, "welcome", "draft", "en", bird.EmailTemplatesVersionsLanguagesSetParams{
				Subject: "Welcome", Text: bird.Ptr("Hello"), HTML: html, Revision: revision,
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Email.Templates.Versions.Languages.Update(ctx, "welcome", "draft", "en", bird.EmailTemplatesVersionsLanguagesUpdateParams{
				Subject: bird.Ptr("Welcome"), Text: bird.Value("Hello"), HTML: html, Revision: revision,
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Email.Templates.Versions.Submit(ctx, "welcome", "draft", bird.EmailTemplatesVersionsSubmitParams{ExpectedRevision: revision})
			if err != nil {
				t.Fatal(err)
			}
			want := []map[string]any{
				{"slug": "welcome", "category": "transactional", "source": "html"},
				{"subject": "Welcome", "text": "Hello"},
				{"subject": "Welcome", "text": "Hello"},
				{},
			}
			if present {
				want[0]["description"] = ""
				want[0]["name"] = "Welcome"
				want[0]["default_language"] = "en"
				for _, body := range want[1:3] {
					body["revision"] = float64(0)
					body["html"] = ""
				}
				want[3]["expected_revision"] = float64(0)
			}
			if !reflect.DeepEqual(bodies, want) {
				t.Errorf("request bodies = %#v, want %#v", bodies, want)
			}
		})
	}
}
