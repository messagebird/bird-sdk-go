package bird

import "strings"

// detectCaller infers the environment driving the SDK, for the Bird-Caller
// usage-telemetry label, by walking the generated callerRules in order (the
// single source is clients/caller-detection.yaml, shared with the CLI and the
// other SDKs). Best-effort and non-authoritative: it only labels traffic, never
// gates behavior. getenv is injected so the rules are table-tested against the
// shared golden vectors.
func detectCaller(getenv func(string) string) string {
	for _, r := range callerRules {
		v := getenv(r.env)
		if v == "" || (r.equals != "" && v != r.equals) {
			continue
		}
		if !r.passthrough {
			return r.name
		}
		if c := sanitizeCaller(v); c != "" {
			return c
		}
	}
	return callerDefault
}

// sanitizeCaller lowercases and bounds a passthrough (AGENT=<name>) value the
// same charset+length way as the other Bird-* labels, dropping boolean-ish
// values that carry no harness identity (e.g. OpenCode sets AGENT=1).
func sanitizeCaller(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" || len(v) > 32 {
		return ""
	}
	if _, skip := callerBooleanishSkip[v]; skip {
		return ""
	}
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
		default:
			return ""
		}
	}
	return v
}
