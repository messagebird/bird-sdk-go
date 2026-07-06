package apierror

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
)

// jsonTags returns the set of JSON field names on a struct type, dropping "-" and
// tag options like ",omitempty".
func jsonTags(t reflect.Type) map[string]bool {
	out := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if name != "" && name != "-" {
			out[name] = true
		}
	}
	return out
}

// The SDK error facade is hand-maintained (no generator emits wireError/APIError),
// so this is the guard: every field on the generated oapi.ErrorBody must be read
// by the facade's wireError. A new wire field (e.g. a future recovery field) fails
// here until it is surfaced in apierror.go.
func TestWireErrorCoversErrorBody(t *testing.T) {
	t.Parallel()
	wire := jsonTags(reflect.TypeOf(wireError{}))
	for name := range jsonTags(reflect.TypeOf(oapi.ErrorBody{})) {
		if !wire[name] {
			t.Errorf("oapi.ErrorBody field %q is not read by wireError — surface it in apierror.go", name)
		}
	}
}

// FromResponse surfaces the wire recovery (remediation + next, ADR-0073) on the
// typed error.
func TestFromResponseSurfacesRecovery(t *testing.T) {
	t.Parallel()
	body := `{"error":{"type":"conflict_error","code":"E11003","message":"pool not empty",` +
		`"remediation":"Reassign the pool's dedicated IPs, then delete it.",` +
		`"next":[{"operation":"assignDedicatedIp","description":"Assign a dedicated IP","scope":"email:write"}]}}`
	err := FromResponse(http.StatusConflict, []byte(body), http.Header{})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("FromResponse returned %T, want *APIError", err)
	}
	if apiErr.Remediation != "Reassign the pool's dedicated IPs, then delete it." {
		t.Errorf("remediation not surfaced: %q", apiErr.Remediation)
	}
	if len(apiErr.Next) != 1 {
		t.Fatalf("next not surfaced: %+v", apiErr.Next)
	}
	if n := apiErr.Next[0]; n.Operation != "assignDedicatedIp" || n.Description != "Assign a dedicated IP" || n.Scope != "email:write" {
		t.Errorf("next action wrong: %+v", n)
	}
}
