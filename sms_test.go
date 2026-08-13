package bird_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

// TestSmsSendSmartEncoding pins the option through the facade: it is nil-omitted, so a
// caller who says nothing sends no `options` and takes Bird's default, and both explicit
// values reach the wire under `options.smart_encoding`.
func TestSmsSendSmartEncoding(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		param *bool
		want  any // the decoded body's "options", nil when the key is absent
	}{
		{"omitted", nil, nil},
		{"opted in", boolPtr(true), map[string]any{"smart_encoding": true}},
		{"explicitly off", boolPtr(false), map[string]any{"smart_encoding": false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got map[string]any
			server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &got)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"sms_01kzv55h4efwqb9jct5011gcnw","status":"accepted"}`))
			})

			_, err := newClient(t, server).Sms.Send(t.Context(), bird.SmsSendParams{
				To:            "+15551234567",
				Text:          "hi",
				Category:      bird.SMSCategoryTransactional,
				SmartEncoding: tc.param,
			})
			if err != nil {
				t.Fatalf("Send: %v", err)
			}

			opts, ok := got["options"]
			if tc.want == nil {
				if ok {
					t.Fatalf("options present on an unset param: %v", opts)
				}
				return
			}
			if !ok {
				t.Fatal("options absent")
			}
			if want, gotJSON := jsonOf(t, tc.want), jsonOf(t, opts); want != gotJSON {
				t.Fatalf("options = %s, want %s", gotJSON, want)
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }

func jsonOf(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}
