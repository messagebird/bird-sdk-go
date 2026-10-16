package bird_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

// TestBroadcastsUpdateClearsCollections pins the three-state semantics on
// Headers, Tags and Metadata: an empty collection is how the update body
// clears them, so a nil has to stay off the wire while a pointer to an empty
// one has to reach it. Sending nothing for an intended clear returns a 200
// with the field still set, which the caller cannot tell from success.
func TestBroadcastsUpdateClearsCollections(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		params bird.BroadcastsUpdateParams
		want   map[string]any
		absent []string
	}{
		{
			name: "an empty collection is sent so the server clears it",
			params: bird.BroadcastsUpdateParams{
				Headers:  &map[string]string{},
				Tags:     &[]bird.EmailTag{},
				Metadata: &map[string]any{},
			},
			want: map[string]any{
				"headers":  map[string]any{},
				"tags":     []any{},
				"metadata": map[string]any{},
			},
		},
		{
			name:   "a nil collection stays absent so the server leaves it alone",
			params: bird.BroadcastsUpdateParams{AudienceID: "adn_01krdgeqcxet5s7t44vh8rt9mg"},
			absent: []string{"headers", "tags", "metadata"},
		},
		{
			// The Go idiom for building a clear is to declare the collection and
			// never append to it, which leaves the pointer non-nil over a nil
			// value. encoding/json writes that as null, and the request schema
			// has no null branch for either field, so it was a 400 rather than a
			// clear.
			name: "a pointer to a nil collection still clears rather than sending null",
			params: func() bird.BroadcastsUpdateParams {
				var headers map[string]string
				var tags []bird.EmailTag
				var metadata map[string]any
				return bird.BroadcastsUpdateParams{Headers: &headers, Tags: &tags, Metadata: &metadata}
			}(),
			want: map[string]any{
				"headers":  map[string]any{},
				"tags":     []any{},
				"metadata": map[string]any{},
			},
		},
		{
			name: "a populated collection is sent through unchanged",
			params: bird.BroadcastsUpdateParams{
				Headers: &map[string]string{"x-campaign-id": "spring"},
				Tags:    &[]bird.EmailTag{{Name: "team", Value: "growth"}},
			},
			want: map[string]any{
				"headers": map[string]any{"x-campaign-id": "spring"},
				"tags":    []any{map[string]any{"name": "team", "value": "growth"}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got map[string]any
			server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &got)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"eb_01krdgeqcxet5s7t44vh8rt9mg"}`))
			})

			_, err := newClient(t, server).Broadcasts.Update(
				context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg", tc.params)
			if err != nil {
				t.Fatalf("Update: %v", err)
			}

			for field, want := range tc.want {
				gotJSON, _ := json.Marshal(got[field])
				wantJSON, _ := json.Marshal(want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("%s: got %s, want %s", field, gotJSON, wantJSON)
				}
			}
			for _, field := range tc.absent {
				if _, ok := got[field]; ok {
					t.Errorf("%s present when unset: %v", field, got[field])
				}
			}
		})
	}
}

// A language, parameters and overrides travel inside the template reference,
// so clearing the template in the same call leaves them nowhere to go; the SDK
// refuses, naming the input, rather than dropping it on the way to the wire.
func TestBroadcastsUpdateRefusesTemplateInputsOnAClearedTemplate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		params bird.BroadcastsUpdateParams
		want   string
	}{
		{"language", bird.BroadcastsUpdateParams{Language: bird.Value("nl")}, "Language needs a Template"},
		{"parameters", bird.BroadcastsUpdateParams{Parameters: &map[string]json.RawMessage{}}, "Parameters need a Template"},
		{"parameter overrides", bird.BroadcastsUpdateParams{ParameterOverrides: &map[string]map[string]json.RawMessage{}}, "ParameterOverrides need a Template"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
				t.Error("request reached the server")
				w.WriteHeader(http.StatusInternalServerError)
			})

			tc.params.Template = bird.Null[string]()
			_, err := newClient(t, server).Broadcasts.Update(
				context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg", tc.params)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Update: got %v, want a %s error", err, tc.want)
			}
		})
	}
}

// Overrides travel inside the template reference, so a create without one has
// nowhere to put them; the SDK refuses rather than sending a legacy draft.
func TestBroadcastsCreateRefusesParameterOverridesWithoutATemplate(t *testing.T) {
	t.Parallel()

	for name, overrides := range map[string]map[string]map[string]json.RawMessage{
		"empty":     {},
		"populated": {"es": {"first_name": json.RawMessage(`"Ana"`)}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
				t.Error("request reached the server")
				w.WriteHeader(http.StatusInternalServerError)
			})

			_, err := newClient(t, server).Broadcasts.Create(context.Background(), bird.BroadcastsCreateParams{
				AudienceID:         "adn_01krdgeqcxet5s7t44vh8rt9mg",
				ParameterOverrides: overrides,
			})
			if err == nil || !strings.Contains(err.Error(), "ParameterOverrides need a Template") {
				t.Fatalf("Create: got %v, want a ParameterOverrides need a Template error", err)
			}
		})
	}
}

// Clearing the template already clears its language, so a Null language sent
// beside a Null template is the same request twice, and the wire carries one
// null template.
func TestBroadcastsUpdateClearsTemplateAndLanguageTogether(t *testing.T) {
	t.Parallel()

	var got map[string]json.RawMessage
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"eb_01krdgeqcxet5s7t44vh8rt9mg","status":"draft"}`))
	})

	if _, err := newClient(t, server).Broadcasts.Update(
		context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg",
		bird.BroadcastsUpdateParams{Template: bird.Null[string](), Language: bird.Null[string]()}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if string(got["template"]) != "null" {
		t.Fatalf("template: got %s, want null", got["template"])
	}
}

func TestBroadcastParameterNumericWire(t *testing.T) {
	t.Parallel()
	parameters := map[string]json.RawMessage{
		"integer":  json.RawMessage(`9007199254740993`),
		"decimal":  json.RawMessage(`1.2300`),
		"exponent": json.RawMessage(`1e+30`),
		"nested":   json.RawMessage(`[{"literal":9007199254740993},null,false,{},[]]`),
	}
	calls := 0
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Idempotency-Key") == "" {
			t.Error("logical POST key missing")
		}
		if strings.HasSuffix(r.URL.Path, "/test") && r.Header.Get("Idempotency-Key") != "broadcast-member-test" {
			t.Errorf("Test intent key: got %q", r.Header.Get("Idempotency-Key"))
		}
		var body struct {
			Template struct {
				Parameters         map[string]json.RawMessage            `json:"parameters"`
				ParameterOverrides map[string]map[string]json.RawMessage `json:"parameter_overrides"`
			} `json:"template"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if got := string(body.Template.ParameterOverrides["it"]["integer"]); got != "9007199254740995" {
			t.Errorf("override raw token: got %s", got)
		}
		if es, ok := body.Template.ParameterOverrides["es"]; !ok || es == nil {
			t.Errorf("a nil language map must reach the wire as an object, got %v", body.Template.ParameterOverrides)
		}
		if len(body.Template.Parameters) != len(parameters) {
			t.Fatalf("parameter count: %d", len(body.Template.Parameters))
		}
		for key, want := range parameters {
			var actual, expected strings.Builder
			var decoded any
			decoder := json.NewDecoder(strings.NewReader(string(body.Template.Parameters[key])))
			decoder.UseNumber()
			if err := decoder.Decode(&decoded); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			actual.Write(encoded)
			decoder = json.NewDecoder(strings.NewReader(string(want)))
			decoder.UseNumber()
			if err := decoder.Decode(&decoded); err != nil {
				t.Fatal(err)
			}
			encoded, err = json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			expected.Write(encoded)
			if actual.String() != expected.String() {
				t.Errorf("%s raw tokens: got %s want %s", key, actual.String(), expected.String())
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{}`))
	})
	client := newClient(t, server)
	template := "emt_01krdgeqcxet5s7t44vh8rt9mg"
	overrides := map[string]map[string]json.RawMessage{"it": {"integer": json.RawMessage(`9007199254740995`)}, "es": nil}
	selected := bird.BroadcastsValidateParams{Template: bird.Value(bird.BroadcastsParameterTemplate{ID: &template, Parameters: &parameters, ParameterOverrides: &overrides})}
	_, err := client.Broadcasts.Create(context.Background(), bird.BroadcastsCreateParams{Template: template, Parameters: parameters, ParameterOverrides: overrides})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Broadcasts.Update(context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg", bird.BroadcastsUpdateParams{Template: bird.Value(template), Parameters: &parameters, ParameterOverrides: &overrides})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Broadcasts.Validate(context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg", selected)
	if err != nil {
		t.Fatal(err)
	}
	preview := bird.BroadcastsPreviewParams{BroadcastsValidateParams: selected}
	_, err = client.Broadcasts.Preview(context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg", preview)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Broadcasts.Test(context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg", bird.BroadcastsTestParams{BroadcastsPreviewParams: preview, To: "member@example.com", PreviewFingerprint: "matching"}, option.WithIdempotencyKey("broadcast-member-test"))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 5 {
		t.Fatalf("captured calls: %d", calls)
	}
}

func TestBroadcastTestRequiresCallerIntentKey(t *testing.T) {
	t.Parallel()
	for _, opts := range [][]option.RequestOption{nil, {option.WithIdempotencyKey("")}} {
		calls := 0
		server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"id":"em_1"}`))
		})
		client := newClient(t, server)
		_, err := client.Broadcasts.Test(context.Background(), "eb_01krdgeqcxet5s7t44vh8rt9mg", bird.BroadcastsTestParams{To: "member@example.com", PreviewFingerprint: "matching"}, opts...)
		if err == nil || !strings.Contains(err.Error(), "idempotency") {
			t.Errorf("expected caller intent key error, got %v", err)
		}
		if calls != 0 {
			t.Errorf("sent %d requests without caller intent key", calls)
		}
	}
}
