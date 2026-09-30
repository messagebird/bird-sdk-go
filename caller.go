package bird

import "strings"

type callerInfo struct {
	model       string
	modelSource string
	name        string
	source      string
	execution   string
}

func detectCallerInfo(getenv func(string) string) callerInfo {
	for _, r := range callerRules {
		v := getenv(r.env)
		_, falseLike := callerFalseLike[strings.ToLower(strings.TrimSpace(v))]
		if strings.TrimSpace(v) == "" || falseLike || (r.equals != "" && v != r.equals) {
			continue
		}
		caller := r.name
		if r.passthrough {
			caller = sanitizeCaller(v)
		}
		if caller == "" {
			continue
		}
		evidence := r.execution
		if r.verification == "unverified" {
			evidence = "unknown"
		}
		signal := "env:" + r.env
		info := callerInfo{name: caller, source: signal, execution: evidence}
		if modelEnv := callerModelEnv[caller]; modelEnv != "" {
			info.model = normalizeModel(getenv(modelEnv))
			if info.model != "" {
				info.modelSource = "env:" + modelEnv
			}
		}
		return info
	}
	return callerInfo{name: callerDefault, source: "fallback", execution: "unknown"}
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

func normalizeModel(raw string) string {
	model := strings.ToLower(strings.TrimSpace(raw))
	if model == "" {
		return ""
	}
	if _, ok := callerPublicModels[model]; ok {
		return model
	}
	return "other"
}

func clientEnrichmentDisabled(getenv func(string) string) bool {
	return getenv("DO_NOT_TRACK") != "" || getenv("BIRD_TELEMETRY") == "0" || getenv("BIRD_CLIENT_ENRICHMENT") == "0"
}
