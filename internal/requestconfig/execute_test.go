package requestconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/messagebird/bird-sdk-go/internal/apierror"
)

func TestExecuteDoesNotRetryLocalValidation(t *testing.T) {
	cfg := Config{MaxRetries: 2}
	invalid := &apierror.RequestValidationError{Field: "VoiceInboundConfigurationPut.route"}
	attempts := 0
	_, err := cfg.Execute(context.Background(), true, func(context.Context, string) (*http.Response, error) {
		attempts++
		return nil, &json.MarshalerError{Err: invalid}
	})
	if err != invalid || attempts != 1 {
		t.Fatalf("error = %v; attempts = %d", err, attempts)
	}
}
