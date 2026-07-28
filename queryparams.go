package bird

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Optional query-param builders: each maps a clean Go value to the pointer the
// generated wire params struct wants, returning nil for the zero value so the
// param is omitted. Every generated resource's toWire calls these.

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
