package bird

import (
	"context"
	"net/http"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/internal/requestconfig"
	"github.com/messagebird/bird-sdk-go/option"
)

// BroadcastsService sends one piece of template content to a stored audience:
// draft a broadcast, update it, send or schedule it, cancel it, and read how it
// landed. Reach it via Client.Broadcasts.
//
// List, Get, Delete, Cancel, ListEvents, ListRecipients, Counts,
// ListClickedLinks and SendQuota are generated; Create, Update and Send are
// hand-written because their bodies carry a From address union and a
// ScheduledAt instant the facade generator has no wire conversion for.
type BroadcastsService struct{ resource }

// BroadcastsCreateParams is a broadcast to create. Every field is optional, so
// the zero value creates an empty draft; a broadcast cannot send until it has a
// From on a verified domain, an AudienceID and a Template.
//
// From and ReplyTo accept either a bare email address or RFC 5322 mailbox
// syntax with a display name: "Newsletter <news@example.com>".
type BroadcastsCreateParams struct {
	// The address the broadcast sends from. The domain has to be one this workspace has verified.
	From string
	// The audience this broadcast sends to. Its contacts are taken as they stand when the send starts, minus any suppressed addresses.
	AudienceID string
	// The template the broadcast sends, as its ID (`emt_…`). A draft can leave it out, but a broadcast cannot send without one. Its published version is fixed when the broadcast is prepared for sending.
	Template string
	// Where replies to this broadcast should go. Up to 25 addresses.
	ReplyTo []string
	// Custom email headers to set on the broadcast. Up to 25 of them, each value up to 998 characters. `List-Unsubscribe` and `List-Unsubscribe-Post` are ours and are dropped if you send them.
	Headers map[string]string
	// Labels on this broadcast, up to 20 of them. Filter the broadcast list by a tag, break stats down by one, and read them back off webhook payloads.
	Tags []EmailTag
	// Any JSON you want to keep on the broadcast, up to 2 KB once serialized. Handed back on reads and in webhook payloads.
	Metadata map[string]any
	// TrackOpens and TrackClicks are pointers because the server default is
	// true — a nil leaves the default, false explicitly disables tracking.
	TrackOpens  *bool
	TrackClicks *bool
	// The IP pool to send this broadcast from. Pass a pool ID, or `ipp_shared` to send through the shared pool on purpose. Empty uses your organization's default pool.
	IPPoolID string
	// What kind of email this is: it decides which suppressions apply and whether the one-click unsubscribe headers are added. Empty sends as marketing.
	Category Category
	// Whether to send the broadcast as soon as it is created. False leaves a draft you can update and send later.
	Send bool
	// When to send the broadcast: at least 30 seconds and at most 365 days from
	// now. Only meaningful alongside Send; on its own the request is refused.
	ScheduledAt time.Time
}

func (p BroadcastsCreateParams) toWire() oapi.EmailBroadcastCreateRequest {
	body := oapi.EmailBroadcastCreateRequest{
		AudienceId:  optStr(p.AudienceID),
		Headers:     optMap(p.Headers),
		Metadata:    optMap(p.Metadata),
		IpPoolId:    optStr(p.IPPoolID),
		TrackOpens:  p.TrackOpens,
		TrackClicks: p.TrackClicks,
		ScheduledAt: optTime(p.ScheduledAt),
	}
	if p.From != "" {
		from := addressInput(p.From)
		body.From = &from
	}
	if p.Template != "" {
		body.Template = &oapi.EmailBroadcastTemplate{Id: p.Template}
	}
	if len(p.ReplyTo) > 0 {
		replyTo := addressInputs(p.ReplyTo)
		body.ReplyTo = &replyTo
	}
	if len(p.Tags) > 0 {
		tags := p.Tags
		body.Tags = &tags
	}
	if p.Category != "" {
		category := oapi.EmailBroadcastCreateRequestCategory(p.Category)
		body.Category = &category
	}
	// Omitted rather than sent as false: the server defaults it to false, and a
	// literal false alongside no scheduled_at is the same draft either way.
	if p.Send {
		body.Send = optBool(true)
	}
	return body
}

// withDefaults fills any field the caller left unset from the configured
// channel defaults, so a broadcast sends under the same policy as the client's
// other email. The per-call value always wins.
//
// Applied to the params rather than the wire body, because the wire encoder
// folds an explicit empty map into an omitted key (optMap nils a len-0 map): a
// caller passing Headers: map[string]string{} means "no headers", and merging
// after that conversion would silently replace it with the configured ones,
// where TypeScript, Python and PHP all keep the empty collection.
//
// Create only. An update leaves an unset field at whatever the draft already
// holds, so filling one from a default there would overwrite a stored value the
// caller never named.
func (p BroadcastsCreateParams) withDefaults(d requestconfig.EmailDefaults) BroadcastsCreateParams {
	if p.From == "" {
		p.From = d.From
	}
	if p.ReplyTo == nil {
		p.ReplyTo = d.ReplyTo
	}
	if p.Category == "" {
		p.Category = d.Category
	}
	if p.TrackOpens == nil {
		p.TrackOpens = d.TrackOpens
	}
	if p.TrackClicks == nil {
		p.TrackClicks = d.TrackClicks
	}
	if p.Headers == nil {
		p.Headers = d.Headers
	}
	if p.Tags == nil {
		p.Tags = d.Tags
	}
	if p.Metadata == nil {
		p.Metadata = d.Metadata
	}
	if p.IPPoolID == "" {
		p.IPPoolID = d.IpPoolID
	}
	return p
}

// Create saves a broadcast. It is a draft unless Send is set, in which case it
// goes out immediately, or at ScheduledAt when one is given. A send needs a
// verified From, an AudienceID and a Template; without them the call is refused
// with a 422 rather than saved. Retried safely with a reused idempotency key.
func (s *BroadcastsService) Create(ctx context.Context, params BroadcastsCreateParams, opts ...option.RequestOption) (*EmailBroadcast, error) {
	// Resolved here rather than inside post, because the channel defaults have to
	// reach the body before it is built. Same shape as EmailService.Send.
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	editors, err := s.client.credentialEditors(cfg, nil)
	if err != nil {
		return nil, err
	}
	wire := params.withDefaults(cfg.EmailDefaults).toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		op := &oapi.CreateEmailBroadcastParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateEmailBroadcast(ctx, op, wire, editors...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailBroadcast
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BroadcastsUpdateParams changes a broadcast that is still a draft or is
// scheduled. Whatever it sets is applied; anything left at its zero value keeps
// the value the broadcast already had. Template, ReplyTo and IPPoolID are
// clearable: build them with Value to set and Null to clear.
type BroadcastsUpdateParams struct {
	// The address the broadcast sends from. The domain has to be one this workspace has verified.
	From string
	// The audience this broadcast sends to.
	AudienceID string
	// The template the broadcast sends, as its ID (`emt_…`). Null takes the template off a draft.
	Template Nullable[string]
	// Where replies to this broadcast should go. Null removes the addresses already set.
	ReplyTo Nullable[[]string]
	// Headers, Tags and Metadata are pointers because an empty collection is how
	// the wire clears them: a nil leaves what the draft already has alone, while
	// a pointer to an empty map or slice removes it.
	//
	// Custom email headers to set on the broadcast. What you send replaces the headers the draft already had rather than adding to them.
	Headers *map[string]string
	// Labels on this broadcast. What you send replaces the tags the draft already had rather than adding to them.
	Tags *[]EmailTag
	// Any JSON you want to keep on the broadcast. What you send replaces the metadata the draft already had rather than merging into it.
	Metadata *map[string]any
	// TrackOpens and TrackClicks are pointers so a nil leaves the current
	// setting alone, distinct from an explicit false that turns tracking off.
	TrackOpens  *bool
	TrackClicks *bool
	// The IP pool to send this broadcast from. Null falls back to your organization's default pool.
	IPPoolID Nullable[string]
	// What kind of email this is: it decides which suppressions apply and whether the one-click unsubscribe headers are added.
	Category Category
}

// emptyIfNil turns a pointer to a nil map into a pointer to an empty one, so a
// clear built by accumulating into a `var m map[string]string` reaches the wire
// as `{}` rather than the `null` encoding/json writes for a nil map, which the
// request schema has no null branch for and rejects with a 400.
func emptyIfNil[V any](m *map[string]V) *map[string]V {
	if m == nil || *m != nil {
		return m
	}
	empty := map[string]V{}
	return &empty
}

// emptySliceIfNil is emptyIfNil for a list-valued clear, and also copies, so a
// later append by the caller cannot reach the body already sent.
func emptySliceIfNil[T any](s *[]T) *[]T {
	if s == nil {
		return nil
	}
	out := make([]T, len(*s))
	copy(out, *s)
	return &out
}

func (p BroadcastsUpdateParams) toWire() oapi.EmailBroadcastUpdateRequest {
	body := oapi.EmailBroadcastUpdateRequest{
		AudienceId:  optStr(p.AudienceID),
		Headers:     emptyIfNil(p.Headers),
		Metadata:    emptyIfNil(p.Metadata),
		IpPoolId:    p.IPPoolID,
		TrackOpens:  p.TrackOpens,
		TrackClicks: p.TrackClicks,
	}
	if p.From != "" {
		from := addressInput(p.From)
		body.From = &from
	}
	// The wire nests the template reference in an object, so a set and a clear
	// have to be rebuilt rather than passed through as the string Nullable.
	if p.Template.IsSpecified() {
		if p.Template.IsNull() {
			body.Template = Null[oapi.EmailBroadcastTemplate]()
		} else {
			body.Template = Value(oapi.EmailBroadcastTemplate{Id: p.Template.MustGet()})
		}
	}
	if p.ReplyTo.IsSpecified() {
		if p.ReplyTo.IsNull() {
			body.ReplyTo = Null[[]oapi.EmailAddressInput]()
		} else {
			body.ReplyTo = Value(addressInputs(p.ReplyTo.MustGet()))
		}
	}
	body.Tags = emptySliceIfNil(p.Tags)
	if p.Category != "" {
		category := oapi.EmailBroadcastUpdateRequestCategory(p.Category)
		body.Category = &category
	}
	return body
}

// Update changes a draft or scheduled broadcast and returns it as it now
// stands. A broadcast that has started sending can no longer be edited and is
// refused with a 409. Retried safely with a reused idempotency key.
func (s *BroadcastsService) Update(ctx context.Context, broadcastId string, params BroadcastsUpdateParams, opts ...option.RequestOption) (*EmailBroadcast, error) {
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.UpdateEmailBroadcastParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.UpdateEmailBroadcast(ctx, broadcastId, op, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailBroadcast
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BroadcastsSendParams schedules the send. Its zero value sends immediately.
type BroadcastsSendParams struct {
	// When to send the broadcast: at least 30 seconds and at most 365 days from
	// now. Zero sends straight away.
	ScheduledAt time.Time
}

func (p BroadcastsSendParams) toWire() oapi.EmailBroadcastSendNowRequest {
	return oapi.EmailBroadcastSendNowRequest{ScheduledAt: optTime(p.ScheduledAt)}
}

// Send dispatches a draft broadcast, immediately or at ScheduledAt. The
// broadcast needs a verified From, an AudienceID and a Template with a
// published version; one that has already started sending or has reached a
// final state is refused with a 409. Retried safely with a reused idempotency
// key.
func (s *BroadcastsService) Send(ctx context.Context, broadcastId string, params BroadcastsSendParams, opts ...option.RequestOption) (*EmailBroadcast, error) {
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.SendEmailBroadcastParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.SendEmailBroadcast(ctx, broadcastId, op, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailBroadcast
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
