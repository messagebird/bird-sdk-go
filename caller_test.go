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
			Name string            `json:"name"`
			Env  map[string]string `json:"env"`
			Want string            `json:"want"`
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
			got := detectCaller(func(k string) string { return tc.Env[k] })
			if got != tc.Want {
				t.Fatalf("detectCaller(%v) = %q, want %q", tc.Env, got, tc.Want)
			}
		})
	}
}
