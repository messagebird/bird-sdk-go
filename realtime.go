package bird

import (
	"context"
	"errors"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
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
	client *Client

	// Channels reads the app's occupied channels and their members.
	Channels *RealtimeChannelsService
	// Members acts on an app-defined member across all of its connections.
	Members *RealtimeMembersService
}

// RealtimeChannelsService reads channel state. Reach it via
// Client.Realtime.Channels.
type RealtimeChannelsService struct{ client *Client }

// RealtimeMembersService acts on the connections of one member. Reach it via
// Client.Realtime.Members.
type RealtimeMembersService struct{ client *Client }

// RealtimePublishParams is a publish: one event delivered to one or more
// channels of the app.
type RealtimePublishParams struct {
	// Event is the name clients bind to. The bird: and bird_internal: prefixes
	// are reserved for the protocol and rejected.
	Event string
	// Channels are the target channels (up to 100 per call); listing several
	// fans the event out to all of them. Prefix a name with private- or
	// presence- for an authenticated channel.
	Channels []string
	// Data is the event payload — any JSON value (object, array, or scalar),
	// capped at 10 KB serialized. Nil sends no data field at all.
	Data any
	// ExcludeConnectionID suppresses delivery to one connection, so a change
	// isn't echoed back to the client that triggered it.
	ExcludeConnectionID string
	// Include asks for per-channel counts at publish time, and counts as one
	// extra message toward usage.
	Include []RealtimeChannelInclude
}

func (p RealtimePublishParams) toWire() oapi.RealtimePublish {
	return oapi.RealtimePublish{
		Event:               p.Event,
		Channels:            p.Channels,
		Data:                realtimeData(p.Data),
		ExcludeConnectionId: realtimeStr(p.ExcludeConnectionID),
		Include:             realtimeInclude(p.Include),
	}
}

// Publish delivers one event to one or more of the app's channels. It returns
// once the event is accepted; delivery to connected clients is asynchronous.
// Retried safely: a single idempotency key is reused across attempts. Provide
// your own key with option.WithIdempotencyKey.
func (s *RealtimeService) Publish(ctx context.Context, appID string, params RealtimePublishParams, opts ...option.RequestOption) (*RealtimePublishResult, error) {
	cfg, key, secret, err := s.client.realtimeConfig(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.PublishRealtimeAppEventParams{XRealtimeKey: key, XRealtimeSecret: secret}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.PublishRealtimeAppEvent(ctx, appID, p, wire, s.client.callEditors(cfg)...)
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

// RealtimeBatchEventParams is one event of a batch publish. Unlike a single
// publish, a batch item targets exactly one channel.
type RealtimeBatchEventParams struct {
	Event               string
	Channel             string
	Data                any
	ExcludeConnectionID string
	Include             []RealtimeChannelInclude
}

func (p RealtimeBatchEventParams) toWire() oapi.RealtimeBatchEvent {
	return oapi.RealtimeBatchEvent{
		Event:               p.Event,
		Channel:             p.Channel,
		Data:                realtimeData(p.Data),
		ExcludeConnectionId: realtimeStr(p.ExcludeConnectionID),
		Include:             realtimeInclude(p.Include),
	}
}

// RealtimePublishBatchParams is a batch of events sent in one request.
type RealtimePublishBatchParams struct {
	Events []RealtimeBatchEventParams
}

func (p RealtimePublishBatchParams) toWire() oapi.RealtimeBatchPublish {
	w := oapi.RealtimeBatchPublish{Events: make([]oapi.RealtimeBatchEvent, len(p.Events))}
	for i, e := range p.Events {
		w.Events[i] = e.toWire()
	}
	return w
}

// PublishBatch delivers several events, each to a single channel, in one
// request. Retried safely with a reused idempotency key.
func (s *RealtimeService) PublishBatch(ctx context.Context, appID string, params RealtimePublishBatchParams, opts ...option.RequestOption) (*RealtimeBatchPublishResult, error) {
	cfg, key, secret, err := s.client.realtimeConfig(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.PublishRealtimeAppBatchParams{XRealtimeKey: key, XRealtimeSecret: secret}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.PublishRealtimeAppBatch(ctx, appID, p, wire, s.client.callEditors(cfg)...)
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

// RealtimeChannelListParams filters the channel listing. Both fields are
// optional; zero values are omitted from the query.
type RealtimeChannelListParams struct {
	// Prefix keeps only channels whose name starts with it, e.g. "presence-".
	Prefix string
	// Include asks for per-channel counts. RealtimeIncludeMemberCount needs a
	// presence-channel Prefix; RealtimeIncludeConnectionCount needs the app's
	// connection-counting flag. Either mismatch is a 400.
	Include []RealtimeChannelInclude
}

func (p RealtimeChannelListParams) toWire(key, secret string) *oapi.ListRealtimeAppChannelsParams {
	return &oapi.ListRealtimeAppChannelsParams{
		Prefix:          realtimeStr(p.Prefix),
		Include:         realtimeInclude(p.Include),
		XRealtimeKey:    key,
		XRealtimeSecret: secret,
	}
}

// List returns the app's currently occupied channels, sorted by name. The
// Realtime service does not paginate this listing, so the single response holds
// every occupied channel.
func (s *RealtimeChannelsService) List(ctx context.Context, appID string, params RealtimeChannelListParams, opts ...option.RequestOption) (*RealtimeChannelsList, error) {
	cfg, key, secret, err := s.client.realtimeConfig(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListRealtimeAppChannels(ctx, appID, params.toWire(key, secret), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out RealtimeChannelsList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RealtimeChannelGetParams selects the attributes a channel read returns.
type RealtimeChannelGetParams struct {
	// Include asks for per-channel counts. RealtimeIncludeMemberCount is
	// presence-channels only; RealtimeIncludeConnectionCount needs the app's
	// connection-counting flag. Either mismatch is a 400.
	Include []RealtimeChannelInclude
}

func (p RealtimeChannelGetParams) toWire(key, secret string) *oapi.GetRealtimeAppChannelParams {
	return &oapi.GetRealtimeAppChannelParams{
		Include:         realtimeInclude(p.Include),
		XRealtimeKey:    key,
		XRealtimeSecret: secret,
	}
}

// Get returns one channel's occupancy, plus the counts named in params.Include.
func (s *RealtimeChannelsService) Get(ctx context.Context, appID, channelName string, params RealtimeChannelGetParams, opts ...option.RequestOption) (*RealtimeChannelInfo, error) {
	cfg, key, secret, err := s.client.realtimeConfig(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetRealtimeAppChannel(ctx, appID, channelName, params.toWire(key, secret), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out RealtimeChannelInfo
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Members returns the members present on a presence channel. It is a presence
// channel operation — a channel without the presence- prefix has no members.
// The listing is not paginated.
func (s *RealtimeChannelsService) Members(ctx context.Context, appID, channelName string, opts ...option.RequestOption) (*RealtimeChannelMembers, error) {
	cfg, key, secret, err := s.client.realtimeConfig(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		p := &oapi.ListRealtimeAppChannelMembersParams{XRealtimeKey: key, XRealtimeSecret: secret}
		return s.client.oapi.ListRealtimeAppChannelMembers(ctx, appID, channelName, p, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out RealtimeChannelMembers
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RealtimeMemberSendParams is an event addressed to one member rather than to
// channels.
type RealtimeMemberSendParams struct {
	// Event is the name clients bind to. The bird: and bird_internal: prefixes
	// are reserved for the protocol and rejected.
	Event string
	// Data is the event payload — any JSON value (object, array, or scalar),
	// capped at 10 KB serialized. Nil sends no data field at all.
	Data any
}

func (p RealtimeMemberSendParams) toWire() oapi.RealtimeMemberPublish {
	return oapi.RealtimeMemberPublish{
		Event: p.Event,
		Data:  realtimeData(p.Data),
	}
}

// Send delivers an event to one member instead of to a channel. Every
// connection that member currently holds receives it, across tabs and devices,
// so there is no need to track their connections or give them a channel of
// their own. The member must have signed in on the connection to be
// addressable.
//
// Delivery is best-effort: a member holding no connections right now simply
// does not receive the event, and that is not reported back.
func (s *RealtimeMembersService) Send(ctx context.Context, appID, memberID string, params RealtimeMemberSendParams, opts ...option.RequestOption) error {
	cfg, key, secret, err := s.client.realtimeConfig(opts)
	if err != nil {
		return err
	}
	wire := params.toWire()
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.SendRealtimeAppMemberEventParams{XRealtimeKey: key, XRealtimeSecret: secret}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.SendRealtimeAppMemberEvent(ctx, appID, memberID, p, wire, s.client.callEditors(cfg)...)
	})
	return err
}

// Disconnect closes every active connection of one member, across all channels
// — the sign-out or ban path. Retried safely with a reused idempotency key.
func (s *RealtimeMembersService) Disconnect(ctx context.Context, appID, memberID string, opts ...option.RequestOption) error {
	cfg, key, secret, err := s.client.realtimeConfig(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DisconnectRealtimeAppMemberParams{XRealtimeKey: key, XRealtimeSecret: secret}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DisconnectRealtimeAppMember(ctx, appID, memberID, p, s.client.callEditors(cfg)...)
	})
	return err
}

// realtimeConfig resolves the per-call options and pulls out the Realtime app
// credentials every operation needs. Missing credentials fail here, before any
// network call. The key and secret travel as X-Realtime-Key/X-Realtime-Secret,
// stamped by the generated request builder from the wire params — the request
// editors that follow leave them alone, because isReservedHeader lists both.
func (c *Client) realtimeConfig(opts []option.RequestOption) (requestconfig.Config, string, string, error) {
	cfg, err := c.resolve(opts)
	if err != nil {
		return cfg, "", "", err
	}
	if cfg.RealtimeKey == "" || cfg.RealtimeSecret == "" {
		return cfg, "", "", errors.New("bird: Realtime app credentials are required; pass option.WithRealtimeCredentials")
	}
	return cfg, cfg.RealtimeKey, cfg.RealtimeSecret, nil
}

// realtimeStr renders an optional string field; empty is omitted.
func realtimeStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// realtimeData wraps the caller's payload; a nil payload leaves data off the
// wire entirely rather than sending an explicit null.
func realtimeData(v any) *oapi.RealtimeEventData {
	if v == nil {
		return nil
	}
	d := oapi.RealtimeEventData(v)
	return &d
}

// realtimeInclude renders the repeatable include filter; an empty list is
// omitted.
func realtimeInclude(in []RealtimeChannelInclude) *[]RealtimeChannelInclude {
	if len(in) == 0 {
		return nil
	}
	return &in
}
