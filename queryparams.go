package bird

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Optional query-param builders: each maps a clean Go value to the pointer the
// wire params struct wants, returning nil for the zero value so the param is omitted.

// optDate renders a calendar-day query param; a zero time is omitted.
func optDate(t time.Time) *openapi_types.Date {
	if t.IsZero() {
		return nil
	}
	return &openapi_types.Date{Time: t}
}

// optTime renders an instant query param; a zero time is omitted.
func optTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// optEmail renders an email-format query param; the wire type is a distinct
// string type, so a plain *string does not assign.
func optEmail(s string) *openapi_types.Email {
	if s == "" {
		return nil
	}
	e := openapi_types.Email(s)
	return &e
}

func optStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optInt(n int) *int {
	if n <= 0 {
		return nil
	}
	return &n
}

func optBool(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}

// optEnum casts a caller-supplied string to a named wire enum (or string alias)
// query param; an empty string is omitted.
func optEnum[T ~string](s string) *T {
	if s == "" {
		return nil
	}
	v := T(s)
	return &v
}

// optZero omits a query param at its zero value. Used where the params field is
// already the wire enum type, so no cast is needed — only the omission.
func optZero[T ~string](v T) *T {
	if v == "" {
		return nil
	}
	return &v
}

// optSlice passes a repeated query param through; an empty slice is omitted so the
// param is absent rather than sent empty. The element is already the wire type —
// either a plain string or the spec's named enum, which the params field exposes.
func optSlice[T any](ss []T) *[]T {
	if len(ss) == 0 {
		return nil
	}
	return &ss
}
