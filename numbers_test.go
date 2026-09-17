package bird_test

import (
	"net/http"
	"testing"
)

func TestAvailableNumberOwnershipRequirement(t *testing.T) {
	t.Parallel()
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"number":"+447700900201","country_code":"GB","number_type":"mobile","capabilities":["sms","voice"],"ownership_registration_required":true}`))
	})
	number, err := newClient(t, server).Numbers.Available.Get(t.Context(), "+447700900201")
	if err != nil {
		t.Fatal(err)
	}
	if !number.OwnershipRegistrationRequired {
		t.Fatal("decoded number lost its ownership requirement")
	}
}
