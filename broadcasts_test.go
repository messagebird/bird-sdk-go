package bird_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
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
