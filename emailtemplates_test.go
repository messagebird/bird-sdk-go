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

// The populated body of the email.templates.parameters conformance vector. The
// corpus only proves a surface decodes it without error; these tests prove the
// records survive the decode.
const templateParametersBody = `{
	"template_id": "emt_01krdgeqcxet5s7t44vh8rt9mg",
	"observed_version_id": "emv_01krdgeqcxet5s7t44vh8rt9mh",
	"requested_language": "en",
	"effective_language": "en",
	"requirements_fingerprint": "sha256:98603c0e927a4aac43b86f6bac4d2d3f72d6b79260aa1d3da33d18e0fea82522",
	"parameters": [
		{"name": "items", "type": "text", "value_domain": "scalar", "required": true, "sensitive": false,
			"use_constraints": [{"kind": "iterable", "site": "html", "path": "items", "operation": "for"}]},
		{"name": "offer", "type": "text", "value_domain": "scalar", "required": true, "sensitive": false,
			"use_constraints": [
				{"kind": "value", "site": "subject", "path": "offer", "operation": "output"},
				{"kind": "value", "site": "html", "path": "offer", "operation": "if"}
			]},
		{"name": "order_total", "type": "amount", "value_domain": "scalar", "required": false, "sensitive": false,
			"use_constraints": [{"kind": "value", "site": "text", "path": "order_total", "operation": "output"}]}
	],
	"variables": [
		{"path": "bird.contact.first_name", "source": "contact", "source_type": "string",
			"use_constraints": [{"kind": "value", "site": "subject", "path": "bird.contact.first_name", "operation": "output"}]},
		{"path": "bird.unsubscribe_url", "source": "bird",
			"use_constraints": [{"kind": "value", "site": "html", "path": "bird.unsubscribe_url", "operation": "output"}]}
	]
}`

func serveTemplateParameters(t *testing.T, body string) (*bird.Client, *string) {
	t.Helper()
	var query string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	})
	return newClient(t, server), &query
}

func TestEmailTemplateParametersKeepsEveryRecord(t *testing.T) {
	t.Parallel()
	client, query := serveTemplateParameters(t, templateParametersBody)

	out, err := client.Email.Templates.Parameters(context.Background(), "emt_01krdgeqcxet5s7t44vh8rt9mg", bird.EmailTemplatesParametersParams{Language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if *query != "language=en" {
		t.Errorf("query = %q", *query)
	}

	// Re-encoding the decoded result must give back the served body: a field the
	// struct dropped or mapped to the wrong key would be missing or moved here.
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(templateParametersBody), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("decoded result re-encodes as\n%s\nwant the served body", encoded)
	}

	if len(out.Parameters) != 3 || len(out.Variables) != 2 {
		t.Fatalf("records: %d parameters, %d variables", len(out.Parameters), len(out.Variables))
	}
	items, offer, total := out.Parameters[0], out.Parameters[1], out.Parameters[2]
	if items.UseConstraints[0].Kind != bird.EmailTemplateUseKindIterable || items.UseConstraints[0].Site != bird.EmailTemplateUseSiteHtml {
		t.Errorf("items use: %+v", items.UseConstraints[0])
	}
	if len(offer.UseConstraints) != 2 || offer.UseConstraints[0].Site != bird.EmailTemplateUseSiteSubject || *offer.UseConstraints[1].Operation != "if" {
		t.Errorf("offer uses: %+v", offer.UseConstraints)
	}
	if total.Type != bird.EmailTemplateParameterTypeAmount || total.ValueDomain != bird.EmailTemplateParameterValueDomainScalar || total.Required {
		t.Errorf("order_total: %+v", total)
	}
	contact, unsubscribe := out.Variables[0], out.Variables[1]
	if contact.Source != bird.EmailTemplateVariableSourceContact || contact.SourceType == nil || *contact.SourceType != "string" {
		t.Errorf("contact variable: %+v", contact)
	}
	if unsubscribe.Source != bird.EmailTemplateVariableSourceBird || unsubscribe.SourceType != nil || unsubscribe.UseConstraints == nil || (*unsubscribe.UseConstraints)[0].Site != bird.EmailTemplateUseSiteHtml {
		t.Errorf("bird variable: %+v", unsubscribe)
	}
}

// Every enum in the result is open: a value a newer server adds must decode,
// keep its wire spelling, and report itself as unknown.
func TestEmailTemplateParametersKeepsUnknownEnumValues(t *testing.T) {
	t.Parallel()
	client, _ := serveTemplateParameters(t, `{
		"template_id": "emt_01krdgeqcxet5s7t44vh8rt9mg",
		"observed_version_id": "emv_01krdgeqcxet5s7t44vh8rt9mh",
		"effective_language": "en",
		"requirements_fingerprint": "sha256:98603c0e927a4aac43b86f6bac4d2d3f72d6b79260aa1d3da33d18e0fea82522",
		"parameters": [{"name": "link", "type": "url", "value_domain": "list", "required": true, "sensitive": false,
			"use_constraints": [{"kind": "partial", "site": "amp", "path": "link"}]}],
		"variables": [{"path": "bird.workspace.name", "source": "workspace"}]
	}`)

	out, err := client.Email.Templates.Parameters(context.Background(), "emt_01krdgeqcxet5s7t44vh8rt9mg", bird.EmailTemplatesParametersParams{})
	if err != nil {
		t.Fatalf("a value added by a newer server must still decode: %v", err)
	}
	parameter, use, variable := out.Parameters[0], out.Parameters[0].UseConstraints[0], out.Variables[0]
	for _, tc := range []struct {
		got   string
		valid bool
		want  string
	}{
		{string(parameter.Type), parameter.Type.Valid(), "url"},
		{string(parameter.ValueDomain), parameter.ValueDomain.Valid(), "list"},
		{string(use.Kind), use.Kind.Valid(), "partial"},
		{string(use.Site), use.Site.Valid(), "amp"},
		{string(variable.Source), variable.Source.Valid(), "workspace"},
	} {
		if tc.got != tc.want || tc.valid {
			t.Errorf("unknown value %q decoded as %q, Valid() = %v", tc.want, tc.got, tc.valid)
		}
	}
}
