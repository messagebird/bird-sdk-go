package bird_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/crypto/nacl/secretbox"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/internal/realtimecrypto"
	"github.com/messagebird/bird-sdk-go/option"
)

// The shared cross-SDK vectors for encrypted channels. Every SDK replays this
// file verbatim, so a divergence here is a divergence from the other SDKs and
// from the browser clients that have to decrypt what we publish.
const vectorsPath = "testdata/realtime-encryption-vectors.json"

type encryptionVectors struct {
	DeriveSharedSecret []struct {
		ID           string `json:"id"`
		Channel      string `json:"channel"`
		MasterKey    string `json:"master_key"`
		SharedSecret string `json:"shared_secret"`
	} `json:"derive_shared_secret"`
	Encrypt []struct {
		ID         string `json:"id"`
		Channel    string `json:"channel"`
		MasterKey  string `json:"master_key"`
		Plaintext  string `json:"plaintext"`
		Nonce      string `json:"nonce"`
		Ciphertext string `json:"ciphertext"`
	} `json:"encrypt"`
	Decrypt []struct {
		ID         string `json:"id"`
		Channel    string `json:"channel"`
		MasterKey  string `json:"master_key"`
		Nonce      string `json:"nonce"`
		Ciphertext string `json:"ciphertext"`
		Plaintext  string `json:"plaintext"`
		Result     string `json:"result"`
	} `json:"decrypt"`
	AuthorizeChannel []struct {
		ID           string `json:"id"`
		Key          string `json:"key"`
		Secret       string `json:"secret"`
		MasterKey    string `json:"master_key"`
		ConnectionID string `json:"connection_id"`
		Channel      string `json:"channel"`
		MemberData   string `json:"member_data"`
		Auth         string `json:"auth"`
		SharedSecret string `json:"shared_secret"`
	} `json:"authorize_channel"`
}

func loadEncryptionVectors(t *testing.T) encryptionVectors {
	t.Helper()
	data, err := os.ReadFile(vectorsPath)
	if err != nil {
		t.Fatalf("reading %s: %v", vectorsPath, err)
	}
	var vectors encryptionVectors
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatalf("parsing %s: %v", vectorsPath, err)
	}
	return vectors
}

func decodeMasterKey(t *testing.T, encoded string) [realtimecrypto.MasterKeyLen]byte {
	t.Helper()
	key, err := realtimecrypto.DecodeMasterKey(encoded)
	if err != nil {
		t.Fatalf("decoding the vector's master key: %v", err)
	}
	return key
}

func decodeBase64(t *testing.T, what, encoded string) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decoding the vector's %s: %v", what, err)
	}
	return raw
}

func TestRealtimeVectorsDeriveSharedSecret(t *testing.T) {
	vectors := loadEncryptionVectors(t)
	if len(vectors.DeriveSharedSecret) == 0 {
		t.Fatal("no derive_shared_secret vectors loaded")
	}
	for _, v := range vectors.DeriveSharedSecret {
		t.Run(v.ID, func(t *testing.T) {
			secret := realtimecrypto.DeriveSharedSecret(v.Channel, decodeMasterKey(t, v.MasterKey))
			if got := base64.StdEncoding.EncodeToString(secret[:]); got != v.SharedSecret {
				t.Errorf("shared secret for %s: got %s, want %s", v.Channel, got, v.SharedSecret)
			}
		})
	}
}

func TestRealtimeVectorsEncrypt(t *testing.T) {
	vectors := loadEncryptionVectors(t)
	if len(vectors.Encrypt) == 0 {
		t.Fatal("no encrypt vectors loaded")
	}
	for _, v := range vectors.Encrypt {
		t.Run(v.ID, func(t *testing.T) {
			var nonce [24]byte
			copy(nonce[:], decodeBase64(t, "nonce", v.Nonce))
			// The vector's plaintext is the payload's JSON serialization, so it
			// goes in as raw JSON to reach the cipher unchanged.
			envelope, err := realtimecrypto.SealWithNonce(
				v.Channel, json.RawMessage(v.Plaintext), nonce, decodeMasterKey(t, v.MasterKey))
			if err != nil {
				t.Fatalf("SealWithNonce: %v", err)
			}
			if envelope.Nonce != v.Nonce {
				t.Errorf("nonce: got %s, want %s", envelope.Nonce, v.Nonce)
			}
			if envelope.Ciphertext != v.Ciphertext {
				t.Errorf("ciphertext: got %s, want %s", envelope.Ciphertext, v.Ciphertext)
			}
		})
	}
}

// The publish side never decrypts, but the vectors pin both directions: opening
// them here proves the key we derive is the one a subscriber will use, and that a
// tampered or stale-key envelope fails authentication instead of yielding bytes.
func TestRealtimeVectorsDecrypt(t *testing.T) {
	vectors := loadEncryptionVectors(t)
	if len(vectors.Decrypt) == 0 {
		t.Fatal("no decrypt vectors loaded")
	}
	for _, v := range vectors.Decrypt {
		t.Run(v.ID, func(t *testing.T) {
			var nonce [24]byte
			copy(nonce[:], decodeBase64(t, "nonce", v.Nonce))
			key := realtimecrypto.DeriveSharedSecret(v.Channel, decodeMasterKey(t, v.MasterKey))
			plaintext, ok := secretbox.Open(nil, decodeBase64(t, "ciphertext", v.Ciphertext), &nonce, &key)

			if v.Result != "valid" {
				if ok {
					t.Fatalf("a %s envelope opened to %q; it must fail authentication", v.Result, plaintext)
				}
				return
			}
			if !ok {
				t.Fatal("the canonical envelope failed to open")
			}
			if string(plaintext) != v.Plaintext {
				t.Errorf("plaintext: got %s, want %s", plaintext, v.Plaintext)
			}
		})
	}
}

func TestRealtimeVectorsAuthorizeChannel(t *testing.T) {
	vectors := loadEncryptionVectors(t)
	if len(vectors.AuthorizeChannel) == 0 {
		t.Fatal("no authorize_channel vectors loaded")
	}
	for _, v := range vectors.AuthorizeChannel {
		t.Run(v.ID, func(t *testing.T) {
			opts := []option.RequestOption{
				option.WithAPIKey("bk_eu1_test"),
				option.WithRealtimeCredentials(v.Key, v.Secret),
			}
			if v.MasterKey != "" {
				opts = append(opts, option.WithRealtimeEncryptionMasterKey(v.MasterKey))
			}
			client, err := bird.NewClient(opts...)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			auth, err := client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
				ConnectionID: v.ConnectionID,
				ChannelName:  v.Channel,
				MemberData:   v.MemberData,
			})
			if err != nil {
				t.Fatalf("AuthorizeChannel: %v", err)
			}
			if auth.Auth != v.Auth {
				t.Errorf("auth: got %s, want %s", auth.Auth, v.Auth)
			}
			if auth.MemberData != v.MemberData {
				t.Errorf("member_data: got %q, want %q — it is signed and echoed byte-identical", auth.MemberData, v.MemberData)
			}
			if auth.SharedSecret != v.SharedSecret {
				t.Errorf("shared_secret: got %q, want %q", auth.SharedSecret, v.SharedSecret)
			}
		})
	}
}

// The vectors' master key, so a behavior test's envelope can be opened with the
// key the SDK sealed under.
const testMasterKey = "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="

// capturePublish serves a publish and hands back the request body the SDK sent.
func capturePublish(t *testing.T, opts ...option.RequestOption) (*bird.Client, *map[string]any) {
	t.Helper()
	var sent map[string]any
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading the request body: %v", err)
		}
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Errorf("request body is not JSON: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	base := []option.RequestOption{option.WithRealtimeCredentials("rk_test", "rs_test")}
	return newClient(t, server, append(base, opts...)...), &sent
}

// openEnvelope decodes a captured {nonce, ciphertext} envelope and returns the
// plaintext it opens to, proving what the SDK actually encrypted.
func openEnvelope(t *testing.T, channel string, data any) []byte {
	t.Helper()
	envelope, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("data is %T, want the {nonce, ciphertext} envelope", data)
	}
	if len(envelope) != 2 {
		t.Errorf("envelope carries %v, want exactly nonce and ciphertext", envelope)
	}
	nonceRaw, ok := envelope["nonce"].(string)
	if !ok {
		t.Fatalf("envelope has no nonce: %v", envelope)
	}
	ciphertextRaw, ok := envelope["ciphertext"].(string)
	if !ok {
		t.Fatalf("envelope has no ciphertext: %v", envelope)
	}
	var nonce [24]byte
	copy(nonce[:], decodeBase64(t, "nonce", nonceRaw))
	key := realtimecrypto.DeriveSharedSecret(channel, decodeMasterKey(t, testMasterKey))
	plaintext, ok := secretbox.Open(nil, decodeBase64(t, "ciphertext", ciphertextRaw), &nonce, &key)
	if !ok {
		t.Fatal("the published envelope does not open under the channel's derived key")
	}
	return plaintext
}

// The payload must be sealed before the request is built, so the plaintext never
// reaches the wire — and the rest of the body must survive unchanged, since the
// edge still routes on the event name and channel.
func TestRealtimePublishSealsEncryptedChannel(t *testing.T) {
	client, sent := capturePublish(t, option.WithRealtimeEncryptionMasterKey(testMasterKey))

	_, err := client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
		Event:    "order.updated",
		Channels: []string{"private-encrypted-orders"},
		Data:     map[string]any{"order_id": "ord_123", "status": "shipped"},
	})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	body := *sent
	if body["event"] != "order.updated" {
		t.Errorf("event on the wire: got %v, want order.updated", body["event"])
	}
	plaintext := openEnvelope(t, "private-encrypted-orders", body["data"])
	if want := `{"order_id":"ord_123","status":"shipped"}`; string(plaintext) != want {
		t.Errorf("sealed plaintext: got %s, want %s", plaintext, want)
	}
}

// A plain channel is not encrypted even with a master key configured, so a
// workspace can hold one key and still publish ordinary events.
func TestRealtimePublishLeavesPlainChannelsUntouched(t *testing.T) {
	client, sent := capturePublish(t, option.WithRealtimeEncryptionMasterKey(testMasterKey))

	_, err := client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
		Event:    "order.created",
		Channels: []string{"orders", "presence-lobby"},
		Data:     map[string]any{"order_id": "ord_1"},
	})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	data, ok := (*sent)["data"].(map[string]any)
	if !ok {
		t.Fatalf("data on the wire is %T, want the payload object", (*sent)["data"])
	}
	if data["order_id"] != "ord_1" {
		t.Errorf("a plain channel's payload was rewritten: %v", data)
	}
}

// Fanning one encrypted publish across channels must be refused locally: each
// channel derives its own key, so the others would receive an envelope their
// subscribers cannot open — undecryptable garbage rather than a visible failure.
func TestRealtimePublishRejectsEncryptedFanOut(t *testing.T) {
	var hits atomic.Int32
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	client := newClient(t, server,
		option.WithRealtimeCredentials("rk_test", "rs_test"),
		option.WithRealtimeEncryptionMasterKey(testMasterKey))

	_, err := client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
		Event:    "order.updated",
		Channels: []string{"private-encrypted-orders", "orders"},
		Data:     map[string]any{"order_id": "ord_123"},
	})
	if err == nil {
		t.Fatal("publishing an encrypted channel alongside another: got nil error")
	}
	if !strings.Contains(err.Error(), "private-encrypted-") {
		t.Errorf("error should explain the one-channel rule: %v", err)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server saw %d requests; the call must fail before any network call", n)
	}
}

// Publishing to an encrypted channel with no master key configured must fail
// rather than send the payload in the clear on a channel whose subscribers will
// try to decrypt it.
func TestRealtimePublishEncryptedWithoutMasterKey(t *testing.T) {
	var hits atomic.Int32
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	client := newClient(t, server, option.WithRealtimeCredentials("rk_test", "rs_test"))

	_, err := client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
		Event:    "order.updated",
		Channels: []string{"private-encrypted-orders"},
		Data:     map[string]any{"order_id": "ord_123"},
	})
	if err == nil {
		t.Fatal("publishing to an encrypted channel without a master key: got nil error")
	}
	if !strings.Contains(err.Error(), "option.WithRealtimeEncryptionMasterKey") {
		t.Errorf("error should name the option to set: %v", err)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server saw %d requests; the call must fail before any network call", n)
	}
}

// A batch event names one channel, so encrypted and plain events can share a
// batch and each encrypted one seals under its own channel's key. The caller's
// slice must come back untouched — params travel by value, but the events slice
// shares its backing array.
func TestRealtimePublishBatchSealsPerEvent(t *testing.T) {
	client, sent := capturePublish(t, option.WithRealtimeEncryptionMasterKey(testMasterKey))

	events := []bird.RealtimeBatchEventParams{
		{Event: "order.created", Channel: "orders", Data: map[string]any{"id": 1}},
		{Event: "order.updated", Channel: "private-encrypted-orders", Data: map[string]any{"id": 2}},
		{Event: "price.changed", Channel: "private-encrypted-cache-prices", Data: map[string]any{"id": 3}},
	}
	_, err := client.Realtime.PublishBatch(context.Background(), "rap_123", bird.RealtimePublishBatchParams{
		Events: events,
	})
	if err != nil {
		t.Fatalf("PublishBatch: %v", err)
	}

	wire, ok := (*sent)["events"].([]any)
	if !ok || len(wire) != 3 {
		t.Fatalf("events on the wire: got %v, want 3 items", (*sent)["events"])
	}
	plain, ok := wire[0].(map[string]any)["data"].(map[string]any)
	if !ok || plain["id"] != float64(1) {
		t.Errorf("the plain event's payload was rewritten: %v", wire[0])
	}
	// Each envelope opens only under its own channel's derived key, which is what
	// makes the fan-out rule on Publish necessary.
	if got := openEnvelope(t, "private-encrypted-orders", wire[1].(map[string]any)["data"]); string(got) != `{"id":2}` {
		t.Errorf("sealed plaintext: got %s, want {\"id\":2}", got)
	}
	if got := openEnvelope(t, "private-encrypted-cache-prices", wire[2].(map[string]any)["data"]); string(got) != `{"id":3}` {
		t.Errorf("sealed plaintext: got %s, want {\"id\":3}", got)
	}

	for i, e := range events {
		if _, sealed := e.Data.(realtimecrypto.Envelope); sealed {
			t.Errorf("event %d: the caller's Data was replaced with the envelope", i)
		}
	}
}

// A batch naming an encrypted channel without a master key must fail whole rather
// than send that item's payload in the clear alongside the plain ones.
func TestRealtimePublishBatchEncryptedWithoutMasterKey(t *testing.T) {
	var hits atomic.Int32
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	client := newClient(t, server, option.WithRealtimeCredentials("rk_test", "rs_test"))

	_, err := client.Realtime.PublishBatch(context.Background(), "rap_123", bird.RealtimePublishBatchParams{
		Events: []bird.RealtimeBatchEventParams{
			{Event: "order.created", Channel: "orders", Data: map[string]any{"id": 1}},
			{Event: "order.updated", Channel: "private-encrypted-orders", Data: map[string]any{"id": 2}},
		},
	})
	if err == nil {
		t.Fatal("batching an encrypted event without a master key: got nil error")
	}
	if !strings.Contains(err.Error(), "option.WithRealtimeEncryptionMasterKey") {
		t.Errorf("error should name the option to set: %v", err)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server saw %d requests; the call must fail before any network call", n)
	}
}

// The response is marshaled straight into an auth endpoint's reply, so its JSON
// spelling is the contract with the browser client — and the optional fields must
// stay absent on a channel that has no use for them.
func TestRealtimeAuthorizeChannelWireShape(t *testing.T) {
	client, err := bird.NewClient(
		option.WithAPIKey("bk_eu1_test"),
		option.WithRealtimeCredentials("rk_test", "rs_test"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	auth, err := client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
		ConnectionID: "26896.319537",
		ChannelName:  "private-room-1",
	})
	if err != nil {
		t.Fatalf("AuthorizeChannel: %v", err)
	}
	body, err := json.Marshal(auth)
	if err != nil {
		t.Fatalf("marshaling the authorization: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("unmarshaling the authorization: %v", err)
	}
	if _, ok := fields["auth"]; !ok {
		t.Errorf("no auth field in %s", body)
	}
	if len(fields) != 1 {
		t.Errorf("a private channel's authorization carries only auth; got %s", body)
	}
}

// Authorizing without app credentials must fail, and name the option to set —
// signing with an empty secret would produce an auth string the edge rejects,
// which is a much harder failure to read at the browser.
func TestRealtimeAuthorizeChannelWithoutCredentials(t *testing.T) {
	client, err := bird.NewClient(option.WithAPIKey("bk_eu1_test"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
		ConnectionID: "26896.319537", ChannelName: "private-room-1",
	})
	if err == nil {
		t.Fatal("AuthorizeChannel without credentials: got nil error, want a configuration error")
	}
	if !strings.Contains(err.Error(), "option.WithRealtimeCredentials") {
		t.Errorf("error should name the option to set: %v", err)
	}
}

// An encrypted channel additionally needs the master key: without it there is no
// shared_secret to return, and a client that subscribes without one receives
// events it cannot open.
func TestRealtimeAuthorizeChannelEncryptedWithoutMasterKey(t *testing.T) {
	client, err := bird.NewClient(
		option.WithAPIKey("bk_eu1_test"),
		option.WithRealtimeCredentials("rk_test", "rs_test"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
		ConnectionID: "26896.319537", ChannelName: "private-encrypted-orders",
	})
	if err == nil {
		t.Fatal("AuthorizeChannel on an encrypted channel without a master key: got nil error")
	}
	if !strings.Contains(err.Error(), "option.WithRealtimeEncryptionMasterKey") {
		t.Errorf("error should name the option to set: %v", err)
	}
}

// A malformed master key must be refused where it is configured, so the caller
// sees the option's name rather than a cipher-length failure at publish time.
func TestRealtimeEncryptionMasterKeyValidation(t *testing.T) {
	for _, tc := range []struct{ name, key string }{
		{"not base64", "not-base64!!"},
		{"too short", base64.StdEncoding.EncodeToString(make([]byte, 16))},
		{"too long", base64.StdEncoding.EncodeToString(make([]byte, 48))},
		{"empty", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := bird.NewClient(
				option.WithAPIKey("bk_eu1_test"),
				option.WithRealtimeEncryptionMasterKey(tc.key),
			)
			if err == nil {
				t.Fatal("got nil error, want a rejected master key")
			}
			if !strings.Contains(err.Error(), "option.WithRealtimeEncryptionMasterKey") {
				t.Errorf("error should name the option: %v", err)
			}
		})
	}
}
