package apierror

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
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

// The refusing policy limit is read from the body, never parsed out of the
// message, and is absent when the API sent none.
func TestFromResponseSurfacesRateLimitContext(t *testing.T) {
	t.Parallel()
	body := `{"error":{"type":"rate_limit_error","code":"E01003","message":"x","rate_limit":{"policy_key":"email_send","scope":"workspace","window_seconds":60,"limit":300}}}`
	err := FromResponse(http.StatusTooManyRequests, []byte(body), http.Header{"Retry-After": []string{"12"}})
	var limited *RateLimitError
	if !errors.As(err, &limited) {
		t.Fatalf("got %T, want *RateLimitError", err)
	}
	want := RateLimitContext{PolicyKey: "email_send", Scope: "workspace", WindowSeconds: 60, Limit: 300}
	if limited.RateLimit == nil || *limited.RateLimit != want {
		t.Fatalf("RateLimit = %+v, want %+v", limited.RateLimit, want)
	}
	bare := FromResponse(http.StatusTooManyRequests, []byte(`{"error":{"type":"rate_limit_error","code":"E01003","message":"x"}}`), http.Header{})
	if !errors.As(bare, &limited) || limited.RateLimit != nil {
		t.Fatalf("got %+v, want a RateLimitError without RateLimit", bare)
	}
}

// The refusing quota limit is read from the body, never parsed out of the
// message, and is absent when the API sent none.
func TestFromResponseSurfacesQuotaContext(t *testing.T) {
	t.Parallel()
	body := `{"error":{"type":"validation_error","code":"E06009","message":"x","quota":{"quota_key":"webhook_subscriptions_max","scope":"workspace"}}}`
	err := FromResponse(http.StatusUnprocessableEntity, []byte(body), http.Header{})
	var refused *ValidationError
	if !errors.As(err, &refused) {
		t.Fatalf("got %T, want *ValidationError", err)
	}
	want := QuotaContext{QuotaKey: "webhook_subscriptions_max", Scope: "workspace"}
	if refused.Quota == nil || *refused.Quota != want {
		t.Fatalf("Quota = %+v, want %+v", refused.Quota, want)
	}
	bare := FromResponse(http.StatusUnprocessableEntity, []byte(`{"error":{"type":"validation_error","code":"E06009","message":"x"}}`), http.Header{})
	if !errors.As(bare, &refused) || refused.Quota != nil {
		t.Fatalf("got %+v, want a ValidationError without Quota", bare)
	}
}

// ErrorType is hand-listed, so a type the spec adds to ErrorBody.type reaches
// callers only as an unnamed string; too_early_error shipped that way.
func TestErrorTypeCoversTheWireEnum(t *testing.T) {
	t.Parallel()
	facade := stringConsts(t, "apierror.go", "ErrorType")
	wire := stringConsts(t, filepath.Join("..", "oapi", "oapi.gen.go"), "ErrorBodyType")
	if len(wire) == 0 {
		t.Fatal("found no oapi.ErrorBodyType constants; the source scan is broken")
	}
	for v := range wire {
		if !facade[v] {
			t.Errorf("ErrorBody.type %q has no ErrorType constant; add it in apierror.go and re-export it from errors.go", v)
		}
	}

	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "..", "errors.go"), nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	exported := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "apierror" {
				exported[sel.Sel.Name] = true
			}
		}
		return true
	})
	for _, name := range constNames(t, "apierror.go", "ErrorType") {
		if !exported[name] {
			t.Errorf("%s is not re-exported from the bird package; add it to errors.go", name)
		}
	}
}

func constNames(t *testing.T, path, typeName string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			if ident, ok := vs.Type.(*ast.Ident); ok && ident.Name == typeName {
				out = append(out, vs.Names[0].Name)
			}
		}
	}
	return out
}

func TestInferTypeFromBodylessStatus(t *testing.T) {
	t.Parallel()
	for status, want := range map[int]ErrorType{
		http.StatusGone:     ErrorTypeGone,
		http.StatusTooEarly: ErrorTypeTooEarly,
		499:                 ErrorTypeClientClosed,
	} {
		var apiErr *APIError
		if !errors.As(FromResponse(status, []byte("upstream"), http.Header{}), &apiErr) || apiErr.Type != want {
			t.Errorf("status %d: got %+v, want type %s", status, apiErr, want)
		}
	}
}

func stringConsts(t *testing.T, path, typeName string) map[string]bool {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			ident, ok := vs.Type.(*ast.Ident)
			if !ok || ident.Name != typeName || len(vs.Values) != 1 {
				continue
			}
			if lit, ok := vs.Values[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				v, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatal(err)
				}
				out[v] = true
			}
		}
	}
	return out
}
