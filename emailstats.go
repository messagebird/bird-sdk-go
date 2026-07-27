package bird

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// EmailStatsService reads aggregated email statistics. Reach it via
// Client.Email.Stats. Every method is a read; each takes a params struct whose
// fields are all optional (zero values are omitted, and the server applies its
// own defaults for the window, sort, and limit).
type EmailStatsService struct{ resource }

// statsDate renders a calendar-day query param; a zero time is omitted.
func statsDate(t time.Time) *openapi_types.Date {
	if t.IsZero() {
		return nil
	}
	return &openapi_types.Date{Time: t}
}

// statsTime renders an instant query param; a zero time is omitted.
func statsTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func statsStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func statsInt(n int) *int {
	if n <= 0 {
		return nil
	}
	return &n
}

func statsBool(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}

// statsEnum casts a caller-supplied string to a named wire enum (or string
// alias) query param; an empty string is omitted.
func statsEnum[T ~string](s string) *T {
	if s == "" {
		return nil
	}
	v := T(s)
	return &v
}
