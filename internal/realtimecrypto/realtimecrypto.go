// Package realtimecrypto implements the wire crypto for Realtime end-to-end
// encrypted channels (private-encrypted-…): the per-channel key derivation, the
// event envelope, and the channel-auth signature.
//
// It lives in internal/ because the master key never leaves the process — the
// facade uses these to seal a publish and to derive the shared_secret it hands a
// browser client, and nothing here is part of the SDK's public surface.
package realtimecrypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/nacl/secretbox"
)

// ChannelPrefix marks a channel whose events are encrypted end to end.
const ChannelPrefix = "private-encrypted-"

// MasterKeyLen is the required decoded length of the encryption master key.
const MasterKeyLen = 32

// IsEncryptedChannel reports whether a channel's events are encrypted end to end.
func IsEncryptedChannel(name string) bool {
	return strings.HasPrefix(name, ChannelPrefix)
}

// Envelope is what an encrypted event carries as its data. The field names are
// the wire contract every Realtime client decodes.
type Envelope struct {
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

// ErrNoMasterKey reports that no encryption master key is configured. The facade
// turns it into a message naming the option to set.
var ErrNoMasterKey = errors.New("no encryption master key configured")

// DecodeMasterKey decodes the configured master key. Validated up front so a bad
// key fails against the config that holds it rather than as a length panic deep
// in the cipher at publish time.
func DecodeMasterKey(encoded string) ([MasterKeyLen]byte, error) {
	var key [MasterKeyLen]byte
	if encoded == "" {
		return key, ErrNoMasterKey
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return key, fmt.Errorf("must be 32 bytes, base64-encoded: %w", err)
	}
	if len(decoded) != MasterKeyLen {
		return key, fmt.Errorf("must be 32 bytes, base64-encoded; got %d bytes", len(decoded))
	}
	copy(key[:], decoded)
	return key, nil
}

// DeriveSharedSecret returns SHA-256(channel_name || master_key) — the channel's
// secretbox key, and the value a channel-auth response carries as shared_secret.
func DeriveSharedSecret(channelName string, masterKey [MasterKeyLen]byte) [32]byte {
	h := sha256.New()
	h.Write([]byte(channelName))
	h.Write(masterKey[:])
	var secret [32]byte
	copy(secret[:], h.Sum(nil))
	return secret
}

// Seal encrypts an event payload for one encrypted channel under a fresh nonce.
func Seal(channelName string, data any, masterKey [MasterKeyLen]byte) (Envelope, error) {
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return Envelope{}, fmt.Errorf("bird: reading random bytes for the event nonce: %w", err)
	}
	return SealWithNonce(channelName, data, nonce, masterKey)
}

// SealWithNonce is Seal with the nonce supplied, so the shared cross-SDK vectors
// can pin the ciphertext their fixed nonce produces.
func SealWithNonce(channelName string, data any, nonce [24]byte, masterKey [MasterKeyLen]byte) (Envelope, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		return Envelope{}, fmt.Errorf("bird: serializing the event data to encrypt: %w", err)
	}
	key := DeriveSharedSecret(channelName, masterKey)
	box := secretbox.Seal(nil, plaintext, &nonce, &key)
	return Envelope{
		Nonce:      base64.StdEncoding.EncodeToString(nonce[:]),
		Ciphertext: base64.StdEncoding.EncodeToString(box),
	}, nil
}

// SignChannel returns hex(HMAC-SHA256(secret, payload)) — the channel-auth
// signature, over "<connection_id>:<channel_name>" and, on a presence channel,
// the member data appended to it.
func SignChannel(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
