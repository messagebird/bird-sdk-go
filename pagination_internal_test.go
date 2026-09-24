package bird

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
	"github.com/messagebird/bird-sdk-go/option"
)

func TestBodyPageOptionsKeysAcrossRetries(t *testing.T) {
	opts := []option.RequestOption{option.WithIdempotencyKey("first-page"), option.WithMaxRetries(1)}
	var keys []string
	for _, cursor := range []string{"", "next", "last"} {
		cfg := requestconfig.Config{}
		for _, opt := range bodyPageOptions(opts, cursor) {
			if err := opt(&cfg); err != nil {
				t.Fatal(err)
			}
		}
		attempts := 0
		_, err := cfg.Execute(context.Background(), true, func(_ context.Context, key string) (*http.Response, error) {
			keys = append(keys, key)
			status := http.StatusOK
			if attempts == 0 {
				status = http.StatusServiceUnavailable
			}
			attempts++
			return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": []string{"0"}}, Body: io.NopCloser(strings.NewReader("{}"))}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(keys) != 6 || keys[0] != "first-page" || keys[0] != keys[1] || keys[2] != keys[3] || keys[4] != keys[5] || keys[2] == "" || keys[2] == keys[0] || keys[4] == keys[2] {
		t.Fatalf("page/retry keys = %v", keys)
	}
	cfg := requestconfig.Config{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			t.Fatal(err)
		}
	}
	if cfg.IdempotencyKey != "first-page" || cfg.MaxRetries != 1 {
		t.Fatalf("original options changed: %+v", cfg)
	}
}
