package bird

import (
	"encoding/json"
	"os"
	"testing"
)

// TestDetectCaller_GoldenVectors runs the shared cross-language fixtures
// (clients/caller-detection-cases.json) so this SDK's detector stays in lockstep
// with the CLI and the other SDKs.
func TestDetectCaller_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/caller-detection-cases.json")
	if err != nil {
		t.Fatalf("read cases: %v", err)
	}
	var doc struct {
		Cases []struct {
			Name        string            `json:"name"`
			Env         map[string]string `json:"env"`
			Want        string            `json:"want"`
			Source      string            `json:"source"`
			Execution   string            `json:"execution"`
			Model       string            `json:"model"`
			ModelSource string            `json:"model_source"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse cases: %v", err)
	}
	if len(doc.Cases) == 0 {
		t.Fatal("no golden cases loaded")
	}
	for _, tc := range doc.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			got := detectCallerInfo(func(k string) string { return tc.Env[k] }).name
			if got != tc.Want {
				t.Fatalf("detectCaller(%v) = %q, want %q", tc.Env, got, tc.Want)
			}
			if tc.Source != "" {
				info := detectCallerInfo(func(k string) string { return tc.Env[k] })
				if info.name != tc.Want || info.source != tc.Source || info.execution != tc.Execution || info.model != tc.Model || info.modelSource != tc.ModelSource {
					t.Fatalf("caller evidence = %+v, want %s/%s/%s", info, tc.Want, tc.Source, tc.Execution)
				}
			}
		})
	}
}
