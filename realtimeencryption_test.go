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
		if strings.Contains(r.URL.Path, "/members/") {
			w.WriteHeader(http.StatusNoContent)
			return
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
		Data:     bird.Value[any](map[string]any{"order_id": "ord_123", "status": "shipped"}),
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
		Data:     bird.Value[any](map[string]any{"order_id": "ord_1"}),
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

func TestRealtimeEncryptedDataPreservesPresence(t *testing.T) {
	for _, state := range []struct {
		name string
		data bird.Nullable[any]
		want string
	}{
		{"unset", nil, ""},
		{"null", bird.Null[any](), "null"},
		{"value", bird.Value[any](map[string]any{"items": []any{nil, "value"}}), `{"items":[null,"value"]}`},
	} {
		for _, method := range []string{"publish", "publish_batch", "members.send"} {
			t.Run(method+"/"+state.name, func(t *testing.T) {
				client, sent := capturePublish(t, option.WithRealtimeEncryptionMasterKey(testMasterKey))
				var err error
				if method == "publish_batch" {
					_, err = client.Realtime.PublishBatch(context.Background(), "rap_123", bird.RealtimePublishBatchParams{
						Events: []bird.RealtimeBatchEventParams{{Event: "changed", Channel: "private-encrypted-orders", Data: state.data}},
					})
				}
				if method == "members.send" {
					err = client.Realtime.Members.Send(context.Background(), "rap_123", "member_1", bird.RealtimeMemberSendParams{Event: "changed", Data: state.data})
				}
				if method == "publish" {
					_, err = client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
						Event: "changed", Channels: []string{"private-encrypted-orders"}, Data: state.data,
					})
				}
				if err != nil {
					t.Fatal(err)
				}
				body := *sent
				if method == "publish_batch" {
					events, ok := body["events"].([]any)
					if !ok || len(events) != 1 {
						t.Fatalf("events = %#v", body["events"])
					}
					body, ok = events[0].(map[string]any)
					if !ok {
						t.Fatalf("event = %#v", events[0])
					}
				}
				if state.name == "unset" {
					if _, present := body["data"]; present {
						t.Fatal("encrypted publish synthesized an omitted payload")
					}
					return
				}
				if method == "members.send" {
					if _, present := body["data"]; !present {
						t.Fatal("member payload was omitted despite being supplied")
					}
					got, err := json.Marshal(body["data"])
					if err != nil || string(got) != state.want {
						t.Fatalf("member payload = %s, want %s (error %v)", got, state.want, err)
					}
					return
				}
				if got := openEnvelope(t, "private-encrypted-orders", body["data"]); string(got) != state.want {
					t.Fatalf("sealed plaintext = %s, want %s", got, state.want)
				}
			})
		}
	}
}

func TestRealtimePublishEncryptionRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts []option.RequestOption
		call func(*bird.Client) error
		want string
	}{
		{
			name: "fan-out with encrypted channel",
			opts: []option.RequestOption{option.WithRealtimeEncryptionMasterKey(testMasterKey)},
			call: func(c *bird.Client) error {
				_, err := c.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
					Event: "order.updated", Channels: []string{"private-encrypted-orders", "orders"},
					Data: bird.Value[any](map[string]any{"order_id": "ord_123"}),
				})
				return err
			},
			want: "private-encrypted-",
		},
		{
			name: "publish encrypted without master key",
			call: func(c *bird.Client) error {
				_, err := c.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
					Event: "order.updated", Channels: []string{"private-encrypted-orders"},
					Data: bird.Value[any](map[string]any{"order_id": "ord_123"}),
				})
				return err
			},
			want: "option.WithRealtimeEncryptionMasterKey",
		},
		{
			name: "publish batch encrypted without master key",
			call: func(c *bird.Client) error {
				_, err := c.Realtime.PublishBatch(context.Background(), "rap_123", bird.RealtimePublishBatchParams{
					Events: []bird.RealtimeBatchEventParams{
						{Event: "order.created", Channel: "orders", Data: bird.Value[any](map[string]any{"id": 1})},
						{Event: "order.updated", Channel: "private-encrypted-orders", Data: bird.Value[any](map[string]any{"id": 2})},
					},
				})
				return err
			},
			want: "option.WithRealtimeEncryptionMasterKey",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
				hits.Add(1)
				w.WriteHeader(http.StatusOK)
			})
			opts := append([]option.RequestOption{option.WithRealtimeCredentials("rk_test", "rs_test")}, tc.opts...)
			err := tc.call(newClient(t, server, opts...))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want substring %q", err, tc.want)
			}
			if n := hits.Load(); n != 0 {
				t.Errorf("server saw %d requests; the call must fail before any network call", n)
			}
		})
	}
}

// A batch event names one channel, so encrypted and plain events can share a
// batch and each encrypted one seals under its own channel's key. The caller's
// slice must come back untouched — params travel by value, but the events slice
// shares its backing array.
func TestRealtimePublishBatchSealsPerEvent(t *testing.T) {
	client, sent := capturePublish(t, option.WithRealtimeEncryptionMasterKey(testMasterKey))

	events := []bird.RealtimeBatchEventParams{
		{Event: "order.created", Channel: "orders", Data: bird.Value[any](map[string]any{"id": 1})},
		{Event: "order.updated", Channel: "private-encrypted-orders", Data: bird.Value[any](map[string]any{"id": 2})},
		{Event: "price.changed", Channel: "private-encrypted-cache-prices", Data: bird.Value[any](map[string]any{"id": 3})},
	}
	if _, err := client.Realtime.PublishBatch(context.Background(), "rap_123", bird.RealtimePublishBatchParams{Events: events}); err != nil {
		t.Fatalf("PublishBatch: %v", err)
	}

	wire, ok := (*sent)["events"].([]any)
	if !ok || len(wire) != 3 {
		t.Fatalf("events on the wire: got %v, want 3 items", (*sent)["events"])
	}
	if plain, ok := wire[0].(map[string]any)["data"].(map[string]any); !ok || plain["id"] != float64(1) {
		t.Errorf("the plain event's payload was rewritten: %v", wire[0])
	}
	if got := openEnvelope(t, "private-encrypted-orders", wire[1].(map[string]any)["data"]); string(got) != `{"id":2}` {
		t.Errorf("sealed plaintext: got %s, want {\"id\":2}", got)
	}
	if got := openEnvelope(t, "private-encrypted-cache-prices", wire[2].(map[string]any)["data"]); string(got) != `{"id":3}` {
		t.Errorf("sealed plaintext: got %s, want {\"id\":3}", got)
	}
	for i, e := range events {
		if data, _ := e.Data.Get(); func() bool { _, sealed := data.(realtimecrypto.Envelope); return sealed }() {
			t.Errorf("event %d: the caller's Data was replaced with the envelope", i)
		}
	}
}

func TestRealtimeAuthorizeChannelWireShapeAndRefusals(t *testing.T) {
	client, err := bird.NewClient(
		option.WithAPIKey("bk_eu1_test"),
		option.WithRealtimeCredentials("rk_test", "rs_test"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	auth, err := client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
		ConnectionID: "26896.319537", ChannelName: "private-room-1",
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
	if _, ok := fields["auth"]; !ok || len(fields) != 1 {
		t.Errorf("a private channel's authorization carries only auth; got %s", body)
	}

	if _, err := client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
		ConnectionID: "26896.319537", ChannelName: "private-encrypted-orders",
	}); err == nil || !strings.Contains(err.Error(), "option.WithRealtimeEncryptionMasterKey") {
		t.Errorf("encrypted channel without master key: got %v, want option.WithRealtimeEncryptionMasterKey", err)
	}

	noCreds, err := bird.NewClient(option.WithAPIKey("bk_eu1_test"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := noCreds.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
		ConnectionID: "26896.319537", ChannelName: "private-room-1",
	}); err == nil || !strings.Contains(err.Error(), "option.WithRealtimeCredentials") {
		t.Errorf("without credentials: got %v, want option.WithRealtimeCredentials", err)
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
