package bird_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

func TestSmsSendSmartEncoding(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		param *bool
		want  string
	}{
		{"omitted", nil, ""},
		{"opted in", bird.Bool(true), `{"smart_encoding":true}`},
		{"explicitly off", bird.Bool(false), `{"smart_encoding":false}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got map[string]any
			server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &got)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"sms_01kzv55h4efwqb9jct5011gcnw","status":"accepted"}`))
			})

			if _, err := newClient(t, server).Sms.Send(t.Context(), bird.SmsSendParams{
				To: "+15551234567", Text: "hi", Category: bird.SMSCategoryTransactional, SmartEncoding: tc.param,
			}); err != nil {
				t.Fatalf("Send: %v", err)
			}

			opts, ok := got["options"]
			if tc.want == "" {
				if ok {
					t.Fatalf("options present on an unset param: %v", opts)
				}
				return
			}
			if !ok || jsonOf(t, opts) != tc.want {
				t.Fatalf("options = %v, want %s", opts, tc.want)
			}
		})
	}
}

func jsonOf(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}
