package bird

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/realtimecrypto"
	"github.com/messagebird/bird-sdk-go/option"
)

// RealtimeService publishes events to a Realtime app and inspects its live
// state. Reach it via Client.Realtime.
//
// Every call needs the app's own credentials on top of the workspace API key:
// configure them with option.WithRealtimeCredentials, at construction for a
// single app or per call when one client serves several apps. Without them a
// method fails before any request is sent.
//
// The app id is a positional argument rather than client config, so one client
// can address any app the workspace owns.
type RealtimeService struct {
	resource

	// Channels reads the app's occupied channels and their members.
	Channels *RealtimeChannelsService
	// Members acts on an app-defined member across all of its connections.
	Members *RealtimeMembersService
}

// RealtimeChannelsService reads channel state. Reach it via
// Client.Realtime.Channels.
type RealtimeChannelsService struct{ resource }

// RealtimeMembersService acts on the connections of one member. Reach it via
// Client.Realtime.Members.
type RealtimeMembersService struct{ resource }

// The app credentials every Realtime call authenticates with, on top of the
// workspace API key.
var realtimeCredentialSchemes = []string{"RealtimeKey", "RealtimeSecret"}

// RealtimeBatchEventParams is one events item.
type RealtimeBatchEventParams struct {
	// The event name clients bind to. Application event names are free-form; the `bird:` and `bird_internal:` prefixes are reserved for the protocol and rejected.
	Event string
	// A Realtime channel name. Only letters, digits, and _ - = @ , . ; Prefix with `private-` or `presence-` for authenticated channels, or `private-encrypted-` for channels whose payloads are end-to-end encrypted with a key only you hold.
	Channel string
	// Arbitrary JSON payload delivered as the event data — an object, array, or scalar. Cap: 10 KB serialized.
	Data any
	// Exclude this connection from delivery, to avoid echoing a change back to the client that triggered it. The value is the client's connection id, assigned when its connection is established.
	ExcludeConnectionID string
	// Attributes of this event's channel to return alongside the publish (same semantics and validation errors as on the channel endpoints). Requesting attributes counts as one additional message toward usage.
	Include []RealtimeChannelInclude
}

func (p RealtimeBatchEventParams) toWire() oapi.RealtimeBatchEvent {
	body := oapi.RealtimeBatchEvent{}
	body.Event = p.Event
	body.Channel = p.Channel
	if p.Data != nil {
		v := p.Data
		body.Data = &v
	}
	if p.ExcludeConnectionID != "" {
		body.ExcludeConnectionId = Ptr(p.ExcludeConnectionID)
	}
	include := make([]oapi.RealtimeChannelInclude, len(p.Include))
	for i, v := range p.Include {
		include[i] = oapi.RealtimeChannelInclude(v)
	}
	if len(include) > 0 {
		body.Include = &include
	}
	return body
}

// RealtimePublishParams is the request body for publish.
type RealtimePublishParams struct {
	// The event name clients bind to. Application event names are free-form; the `bird:` and `bird_internal:` prefixes are reserved for the protocol and rejected.
	Event string
	// The channels to deliver the event to (up to 100 per call). Prefix with `private-` or `presence-` for authenticated channels. A `private-encrypted-` channel must be the only channel in its publish: each encrypted channel has its own key, so a fan-out would hand the other channels unreadable ciphertext.
	Channels []string
	// Arbitrary JSON payload delivered as the event data — an object, array, or scalar. Cap: 10 KB serialized.
	Data any
	// Exclude this connection from delivery, to avoid echoing a change back to the client that triggered it. The value is the client's connection id, assigned when its connection is established.
	ExcludeConnectionID string
	// Per-channel attributes to return alongside the publish, reflecting each channel's state at publish time (same semantics and validation errors as on the channel endpoints: `member_count` is presence-channels only, `connection_count` requires the app's connection-counting flag). Requesting attributes counts as one additional message toward usage.
	Include []RealtimeChannelInclude
}

func (p RealtimePublishParams) toWire() oapi.RealtimePublish {
	body := oapi.RealtimePublish{}
	body.Event = p.Event
	channels := make([]oapi.RealtimeChannelName, len(p.Channels))
	for i, v := range p.Channels {
		channels[i] = oapi.RealtimeChannelName(v)
	}
	body.Channels = channels
	if p.Data != nil {
		v := p.Data
		body.Data = &v
	}
	if p.ExcludeConnectionID != "" {
		body.ExcludeConnectionId = Ptr(p.ExcludeConnectionID)
	}
	include := make([]oapi.RealtimeChannelInclude, len(p.Include))
	for i, v := range p.Include {
		include[i] = oapi.RealtimeChannelInclude(v)
	}
	if len(include) > 0 {
		body.Include = &include
	}
	return body
}

// RealtimePublishBatchParams is the request body for publish_batch.
type RealtimePublishBatchParams struct {
	// Up to 10 events per batch.
	Events []RealtimeBatchEventParams
}

func (p RealtimePublishBatchParams) toWire() oapi.RealtimeBatchPublish {
	body := oapi.RealtimeBatchPublish{}
	events := make([]oapi.RealtimeBatchEvent, len(p.Events))
	for i, e := range p.Events {
		events[i] = e.toWire()
	}
	body.Events = events
	return body
}

// sealEncrypted replaces Data with its encrypted envelope when the publish names
// a private-encrypted- channel, leaving a plain publish untouched.
//
// Exactly one channel per encrypted publish: each channel derives its own key, so
// a fan-out would hand every other channel ciphertext its subscribers cannot open.
func (p *RealtimePublishParams) sealEncrypted(masterKey string) error {
	channel := ""
	for _, c := range p.Channels {
		if realtimecrypto.IsEncryptedChannel(c) {
			channel = c
			break
		}
	}
	if channel == "" {
		return nil
	}
	if len(p.Channels) > 1 {
		return fmt.Errorf("bird: a publish to a private-encrypted- channel must name exactly that one channel, but this one names %d: every channel derives its own key, so a multi-channel publish would hand the other channels undecryptable ciphertext — publish per channel instead", len(p.Channels))
	}
	key, err := masterKeyFor(masterKey, channel)
	if err != nil {
		return err
	}
	envelope, err := realtimecrypto.Seal(channel, p.Data, key)
	if err != nil {
		return err
	}
	p.Data = envelope
	return nil
}

// sealEncrypted seals each encrypted item under its own channel's key. A batch
// event names one channel, so the items encrypt independently and plain ones pass
// through.
func (p *RealtimePublishBatchParams) sealEncrypted(masterKey string) error {
	encrypted := false
	for _, e := range p.Events {
		if realtimecrypto.IsEncryptedChannel(e.Channel) {
			encrypted = true
			break
		}
	}
	if !encrypted {
		return nil
	}
	// The caller's slice shares its backing array with ours, so seal into a copy
	// rather than replacing the Data they passed in.
	events := make([]RealtimeBatchEventParams, len(p.Events))
	copy(events, p.Events)
	for i, e := range events {
		if !realtimecrypto.IsEncryptedChannel(e.Channel) {
			continue
		}
		key, err := masterKeyFor(masterKey, e.Channel)
		if err != nil {
			return err
		}
		envelope, err := realtimecrypto.Seal(e.Channel, e.Data, key)
		if err != nil {
			return err
		}
		events[i].Data = envelope
	}
	p.Events = events
	return nil
}

// Publish delivers one event to the named channels and reports how it fanned out.
//
// A private-encrypted- channel's payload is sealed here, before the request
// leaves the process: Bird sees only {"nonce", "ciphertext"}, and the master key
// configured with option.WithRealtimeEncryptionMasterKey never travels. Such a
// publish names exactly one channel — each channel derives its own key, so a
// fan-out would deliver ciphertext the other channels' subscribers cannot open.
//
// The payload is sealed once and reused across retries, so a retried publish
// delivers the same envelope rather than re-encrypting under a fresh nonce.
func (s *RealtimeService) Publish(ctx context.Context, realtimeAppId string, params RealtimePublishParams, opts ...option.RequestOption) (*RealtimePublishResult, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	editors, err := s.client.credentialEditors(cfg, realtimeCredentialSchemes)
	if err != nil {
		return nil, err
	}
	if err := params.sealEncrypted(cfg.RealtimeEncryptionMasterKey); err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		op := &oapi.PublishRealtimeAppEventParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.PublishRealtimeAppEvent(ctx, oapi.RealtimeAppID(realtimeAppId), op, wire, editors...)
	})
	if err != nil {
		return nil, err
	}
	var out RealtimePublishResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PublishBatch delivers several events in one request, each to a single channel.
// Every event addressed to a private-encrypted- channel is sealed here under that
// channel's own key, so encrypted and plain events can share a batch.
func (s *RealtimeService) PublishBatch(ctx context.Context, realtimeAppId string, params RealtimePublishBatchParams, opts ...option.RequestOption) (*RealtimeBatchPublishResult, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	editors, err := s.client.credentialEditors(cfg, realtimeCredentialSchemes)
	if err != nil {
		return nil, err
	}
	if err := params.sealEncrypted(cfg.RealtimeEncryptionMasterKey); err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		op := &oapi.PublishRealtimeAppBatchParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.PublishRealtimeAppBatch(ctx, oapi.RealtimeAppID(realtimeAppId), op, wire, editors...)
	})
	if err != nil {
		return nil, err
	}
	var out RealtimeBatchPublishResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RealtimeChannelAuthorizationParams identifies the subscription to sign. Both
// ids come from the body the browser client POSTs to your auth endpoint.
type RealtimeChannelAuthorizationParams struct {
	// The subscribing connection's id.
	ConnectionID string
	// The channel being subscribed.
	ChannelName string
	// Presence channels: the member-identity JSON to sign and echo, carrying
	// member_id and optionally member_info. Signed and returned byte-identical,
	// so hand over the exact string the client will see.
	MemberData string
}

// RealtimeChannelAuthorization is the body your auth endpoint returns to the
// browser client. The JSON tags are the wire spelling the client expects, so it
// can be marshaled as-is.
type RealtimeChannelAuthorization struct {
	// The <key>:<signature> pair the Realtime edge verifies.
	Auth string `json:"auth"`
	// Echo of the signed member data, on a presence channel.
	MemberData string `json:"member_data,omitempty"`
	// The channel's decryption key, base64, on a private-encrypted- channel.
	SharedSecret string `json:"shared_secret,omitempty"`
}

// AuthorizeChannel signs a channel subscription for a browser client and returns
// the body your auth endpoint responds with. It is pure crypto — no request is
// sent — so authorize as many subscriptions as you like: the signature is
// HMAC-SHA256(secret, "<connection_id>:<channel_name>[:<member_data>]") keyed by
// the app secret and prefixed with the app key. A private-encrypted- channel's
// response also carries the channel's shared_secret, derived from the configured
// encryption master key; hand it only to a client you have just authorized to
// read that channel, since it decrypts every event on it.
//
// Authorize only after your own check that this user may join this channel — the
// signature is the edge's only evidence that they may.
//
//	auth, err := client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
//		ConnectionID: req.ConnectionID,
//		ChannelName:  req.ChannelName,
//	})
//
// See ExampleRealtimeService_AuthorizeChannel for the whole auth endpoint.
func (s *RealtimeService) AuthorizeChannel(params RealtimeChannelAuthorizationParams, opts ...option.RequestOption) (*RealtimeChannelAuthorization, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	if cfg.RealtimeKey == "" || cfg.RealtimeSecret == "" {
		return nil, fmt.Errorf("bird: AuthorizeChannel signs with the Realtime app's own credentials; pass option.WithRealtimeCredentials")
	}

	signed := params.ConnectionID + ":" + params.ChannelName
	if params.MemberData != "" {
		signed += ":" + params.MemberData
	}
	out := RealtimeChannelAuthorization{
		Auth:       cfg.RealtimeKey + ":" + realtimecrypto.SignChannel(cfg.RealtimeSecret, signed),
		MemberData: params.MemberData,
	}
	if realtimecrypto.IsEncryptedChannel(params.ChannelName) {
		masterKey, err := masterKeyFor(cfg.RealtimeEncryptionMasterKey, params.ChannelName)
		if err != nil {
			return nil, err
		}
		secret := realtimecrypto.DeriveSharedSecret(params.ChannelName, masterKey)
		out.SharedSecret = base64.StdEncoding.EncodeToString(secret[:])
	}
	return &out, nil
}

// masterKeyFor decodes the configured master key for an operation on channel,
// naming the option to set when there is none — a caller who reaches an
// encrypted channel without a key has a configuration gap, not a crypto one.
func masterKeyFor(configured, channel string) ([realtimecrypto.MasterKeyLen]byte, error) {
	key, err := realtimecrypto.DecodeMasterKey(configured)
	if errors.Is(err, realtimecrypto.ErrNoMasterKey) {
		return key, fmt.Errorf("bird: %s is end-to-end encrypted and needs the encryption master key; pass option.WithRealtimeEncryptionMasterKey", channel)
	}
	if err != nil {
		return key, fmt.Errorf("bird: the configured Realtime encryption master key %w", err)
	}
	return key, nil
}
