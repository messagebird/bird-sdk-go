package bird

import (
	"encoding/json"
	"testing"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
)

// TestAddressInputIsTheStringArmVerbatim pins that an address reaches the wire as a
// JSON string, unparsed. The union's other arm is an object, so a client that
// "helpfully" split a display name would send a shape the server reads differently —
// and the whole reason this conversion is non-fallible is that it does not inspect
// the value at all.
func TestAddressInputIsTheStringArmVerbatim(t *testing.T) {
	t.Parallel()

	for _, addr := range []string{
		"jane@example.com",
		`Jane Doe <jane@example.com>`,
		`"Doe, Jane" <jane@example.com>`,
		"not-an-address-at-all",
		"",
	} {
		got, err := json.Marshal(addressInput(addr))
		if err != nil {
			t.Fatalf("marshal %q: %v", addr, err)
		}
		want, err := json.Marshal(addr)
		if err != nil {
			t.Fatalf("marshal want %q: %v", addr, err)
		}
		if string(got) != string(want) {
			t.Errorf("addressInput(%q) marshalled to %s, want %s", addr, got, want)
		}
	}
}

// TestAddressInputsPreservesOrder pins the slice form. Recipient order is
// observable — the batch result is returned in submission order — so a conversion
// that reordered or dropped one would misattribute every per-recipient result.
func TestAddressInputsPreservesOrder(t *testing.T) {
	t.Parallel()

	in := []string{"a@x.com", "b@x.com", "c@x.com"}
	got := addressInputs(in)
	if len(got) != len(in) {
		t.Fatalf("got %d addresses, want %d", len(got), len(in))
	}
	for i, a := range in {
		want, _ := json.Marshal(a)
		have, _ := json.Marshal(got[i])
		if string(have) != string(want) {
			t.Errorf("index %d: got %s, want %s", i, have, want)
		}
	}
	if addressInputs(nil) == nil {
		t.Error("a nil slice must convert to an empty non-nil slice, since To is a required wire field")
	}
}

// TestToWireCarriesEveryAddressField pins that all four address fields go through
// the conversion and land on the wire. Cc/Bcc/ReplyTo are pointers, so an omitted
// one must stay nil rather than becoming an empty array the server would read as an
// explicit empty list.
func TestToWireCarriesEveryAddressField(t *testing.T) {
	t.Parallel()

	body := EmailSendParams{
		From:    "Sender <from@x.com>",
		To:      []string{"to@x.com"},
		Cc:      []string{"cc1@x.com", "cc2@x.com"},
		Bcc:     []string{"bcc@x.com"},
		ReplyTo: []string{"reply@x.com"},
		Subject: "s",
	}.toWire()

	if got := marshalString(t, body.From); got != "Sender <from@x.com>" {
		t.Errorf("From = %q", got)
	}
	if len(body.To) != 1 {
		t.Fatalf("To has %d entries, want 1", len(body.To))
	}
	if body.Cc == nil || len(*body.Cc) != 2 {
		t.Error("Cc must carry both addresses")
	}
	if body.Bcc == nil || len(*body.Bcc) != 1 {
		t.Error("Bcc must carry its address")
	}
	if body.ReplyTo == nil || len(*body.ReplyTo) != 1 {
		t.Error("ReplyTo must carry its address")
	}

	bare := EmailSendParams{To: []string{"to@x.com"}}.toWire()
	if bare.Cc != nil || bare.Bcc != nil || bare.ReplyTo != nil {
		t.Error("an unset optional address list must stay nil, not become an empty array")
	}
}

// TestDefaultsFillAddressesRatherThanDroppingThem is the behaviour change worth
// pinning. Both default paths used to swallow the conversion error, so a configured
// From or ReplyTo could vanish with no signal. The conversion cannot fail, so the
// value now always lands — and a per-send value still wins.
func TestDefaultsFillAddressesRatherThanDroppingThem(t *testing.T) {
	t.Parallel()

	defaults := requestconfig.EmailDefaults{
		From:    `Default <default@x.com>`,
		ReplyTo: []string{"default-reply@x.com"},
	}

	filled := EmailSendParams{To: []string{"to@x.com"}}.toWire()
	applyEmailDefaults(&filled, defaults)
	if got := marshalString(t, filled.From); got != `Default <default@x.com>` {
		t.Errorf("default From did not land, got %q", got)
	}
	if filled.ReplyTo == nil || len(*filled.ReplyTo) != 1 {
		t.Fatal("default ReplyTo did not land")
	}

	kept := EmailSendParams{
		From:    "explicit@x.com",
		To:      []string{"to@x.com"},
		ReplyTo: []string{"explicit-reply@x.com"},
	}.toWire()
	applyEmailDefaults(&kept, defaults)
	if got := marshalString(t, kept.From); got != "explicit@x.com" {
		t.Errorf("per-send From must win over the default, got %q", got)
	}
	if got := marshalString(t, (*kept.ReplyTo)[0]); got != "explicit-reply@x.com" {
		t.Errorf("per-send ReplyTo must win over the default, got %q", got)
	}
}

// TestMailboxComposeSharesTheAdapter pins that the compose path converts addresses
// the same way. It had its own copy that discarded the error, which is how one
// conversion came to be described three different ways.
func TestMailboxComposeSharesTheAdapter(t *testing.T) {
	t.Parallel()

	body := EmailMailboxesMessagesCreateParams{
		To:      []string{"to@x.com"},
		CC:      []string{`Cc Person <cc@x.com>`},
		ReplyTo: []string{"reply@x.com"},
	}.toWire()

	if len(body.To) != 1 {
		t.Fatalf("To has %d entries, want 1", len(body.To))
	}
	if body.Cc == nil || len(*body.Cc) != 1 {
		t.Fatal("Cc did not reach the wire")
	}
	if got := marshalString(t, (*body.Cc)[0]); got != `Cc Person <cc@x.com>` {
		t.Errorf("compose Cc = %q, want the display-name form verbatim", got)
	}
}

// marshalString reads back the string an EmailAddressInput carries, which is the
// only way to observe the union's contents from outside the generated package.
func marshalString(t *testing.T, in oapi.EmailAddressInput) string {
	t.Helper()

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal address: %v", err)
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("address did not marshal to a JSON string (%s): %v", raw, err)
	}
	return s
}
