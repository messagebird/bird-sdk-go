package bird_test

import (
	"encoding/json"
	"net/http"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/internal/oapi"
)

func TestAMBRoutingCreateNullableUnionPresence(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, kind    string
		group, intent bird.Nullable[string]
		want          map[string]any
	}{
		{"group omitted", "intent", bird.Nullable[string]{}, bird.Value("order_status"), map[string]any{"match_intent_id": "order_status"}},
		{"intent omitted", "group", bird.Value("support"), bird.Nullable[string]{}, map[string]any{"match_group_id": "support"}},
		{"group null", "intent", bird.Null[string](), bird.Value("order_status"), map[string]any{"match_group_id": nil, "match_intent_id": "order_status"}},
		{"intent null", "group", bird.Value("support"), bird.Null[string](), map[string]any{"match_group_id": "support", "match_intent_id": nil}},
		{"both values", "both", bird.Value("support"), bird.Value("order_status"), map[string]any{"match_group_id": "support", "match_intent_id": "order_status"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var got map[string]any
			server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Errorf("request JSON: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			})
			_, err := newClient(t, server).Amb.RoutingRules.Create(t.Context(), bird.AmbRoutingRulesCreateParams{
				BusinessAccountID: "abz_01krdgeqcxet5s7t44vh8rt9mg",
				MatchKind:         bird.AMBRoutingRuleMatchKind(test.kind),
				MatchGroupID:      test.group,
				MatchIntentID:     test.intent,
				Queue:             "support",
			})
			if err != nil {
				t.Fatal(err)
			}
			test.want["business_account_id"] = "abz_01krdgeqcxet5s7t44vh8rt9mg"
			test.want["match_kind"] = test.kind
			test.want["queue"] = "support"
			if jsonOf(t, got) != jsonOf(t, test.want) {
				t.Fatalf("request = %s, want %s", jsonOf(t, got), jsonOf(t, test.want))
			}
		})
	}
}

func TestAMBRoutingUnionUnsetRemovesRetainedKey(t *testing.T) {
	var wire oapi.AMBRoutingRuleCreate
	if err := json.Unmarshal([]byte(`{"business_account_id":"abz_01krdgeqcxet5s7t44vh8rt9mg","match_kind":"both","match_group_id":"old","match_intent_id":"order_status","queue":"support"}`), &wire); err != nil {
		t.Fatal(err)
	}
	wire.MatchGroupId = bird.Nullable[string]{}
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if _, present := got["match_group_id"]; present || got["match_intent_id"] != "order_status" {
		t.Fatalf("retained union overwrote explicit omission: %s", body)
	}
}
