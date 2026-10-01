package bird_test

import (
	"encoding/json"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

// Go decodes a JSON null into a string as "", so only a pointer keeps an
// inline call's missing sequence ID distinct from a saved one.
func TestAcceptedCallDistinguishesAnInlineSequence(t *testing.T) {
	var inline, saved bird.VoiceCall
	if err := json.Unmarshal([]byte(`{"sequence":{"id":null,"run_id":"vsr_01krdgeqcxet5s7t44vh8rt9mg"}}`), &inline); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"sequence":{"id":"vsq_01krdgeqcxet5s7t44vh8rt9mg","run_id":"vsr_01krdgeqcxet5s7t44vh8rt9mg"}}`), &saved); err != nil {
		t.Fatal(err)
	}
	if inline.Sequence == nil || inline.Sequence.Id != nil {
		t.Errorf("inline call sequence = %+v, want a nil ID", inline.Sequence)
	}
	if saved.Sequence == nil || saved.Sequence.Id == nil || *saved.Sequence.Id != "vsq_01krdgeqcxet5s7t44vh8rt9mg" {
		t.Errorf("saved call sequence = %+v, want its ID", saved.Sequence)
	}
}
