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

// The same guard one level down: NextAction is hand-maintained beside the
// generated oapi.NextAction, and the ErrorBody guard above only sees `next` as a
// whole, so a field added inside a step passes it unnoticed. That is how `kind`,
// `params` and `url` were dropped for a release.
func TestNextActionCoversGenerated(t *testing.T) {
	t.Parallel()
	facade := jsonTags(reflect.TypeOf(NextAction{}))
	for name := range jsonTags(reflect.TypeOf(oapi.NextAction{})) {
		if !facade[name] {
			t.Errorf("oapi.NextAction field %q is not read by NextAction — surface it in apierror.go", name)
		}
	}
}

// FromResponse surfaces the wire recovery (remediation + next, ADR-0073/0124) on
// the typed error, across the step kinds: an operation step with the params that
// address it, and a bare external step that carries no operation at all.
func TestFromResponseSurfacesRecovery(t *testing.T) {
	t.Parallel()
	body := `{"error":{"type":"conflict_error","code":"E01028","message":"resource still in use",` +
		`"remediation":"Remove or reassign the resources that still reference this one, then retry the delete.",` +
		`"next":[{"kind":"operation","operation":"assignDedicatedIp","description":"Assign a dedicated IP",` +
		`"params":{"pool_id":"pool_123"}},` +
		`{"kind":"external","description":"Publish the DKIM record at your DNS provider","url":"https://example.test/dns"}]}}`
	err := FromResponse(http.StatusConflict, []byte(body), http.Header{})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("FromResponse returned %T, want *APIError", err)
	}
	if apiErr.Remediation != "Remove or reassign the resources that still reference this one, then retry the delete." {
		t.Errorf("remediation not surfaced: %q", apiErr.Remediation)
	}
	if len(apiErr.Next) != 2 {
		t.Fatalf("next not surfaced: %+v", apiErr.Next)
	}
	op := apiErr.Next[0]
	if op.Kind != "operation" || op.Operation != "assignDedicatedIp" || op.Description != "Assign a dedicated IP" {
		t.Errorf("operation step wrong: %+v", op)
	}
	if op.Params["pool_id"] != "pool_123" {
		t.Errorf("operation step params wrong: %+v", op.Params)
	}
	ext := apiErr.Next[1]
	if ext.Kind != "external" || ext.URL != "https://example.test/dns" {
		t.Errorf("external step wrong: %+v", ext)
	}
	if ext.Operation != "" {
		t.Errorf("external step invented an operation: %q", ext.Operation)
	}
}
