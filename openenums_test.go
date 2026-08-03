package bird_test

import (
	"encoding/json"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
)

// An open-enum field is a defined type carrying its known values, and it must still
// decode a value this SDK version has never heard of. Closing the enum would break
// every client the first time the server adds one, so both halves are asserted
// together.
func TestOpenEnumFieldStaysOpen(t *testing.T) {
	t.Parallel()

	var event struct {
		Type bird.EmailEventType `json:"type"`
	}

	if err := json.Unmarshal([]byte(`{"type":"email.invented_in_2030"}`), &event); err != nil {
		t.Fatalf("a value added by a newer server must still decode: %v", err)
	}
	if got := string(event.Type); got != "email.invented_in_2030" {
		t.Fatalf("unknown value not preserved: %q", got)
	}
	if event.Type.Valid() {
		t.Fatal("Valid() must report an unrecognized value as unknown")
	}

	if err := json.Unmarshal([]byte(`{"type":"email.bounced"}`), &event); err != nil {
		t.Fatal(err)
	}
	if event.Type != bird.EmailEventTypeEmailBounced {
		t.Fatalf("known value should equal its constant, got %q", event.Type)
	}
	if !event.Type.Valid() {
		t.Fatal("Valid() must accept a known value")
	}
}

// The re-exported constant carries the facade's enum type, not a bare string —
// asserted at compile time, since that is where the distinction lives.
var _ bird.VerificationChannel = bird.VerificationChannelSms

// The constants are re-exported from the internal wire layer, which a caller cannot
// import — so the facade alias is what makes them nameable at all.
func TestOpenEnumConstantsAreUsableThroughTheFacade(t *testing.T) {
	t.Parallel()

	channel := bird.VerificationChannelSms
	if string(channel) != "sms" {
		t.Fatalf("constant value: %q", channel)
	}
	if !channel.Valid() {
		t.Fatal("a re-exported constant must be a valid member")
	}
}
