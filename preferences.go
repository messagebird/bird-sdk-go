package bird

import (
	"context"
	"net/http"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// PreferencesService reads and writes the workspace's stated messaging
// preferences: consent grants and opt-outs keyed by channel and handle, and
// optionally scoped to one sender. Reach it via Client.Preferences.
type PreferencesService struct{ resource }

// PreferencesCreateParams records one preference statement. Channel, Handle,
// and Status are required; the rest are optional and omitted from the request
// at their zero value.
type PreferencesCreateParams struct {
	// Channel the statement applies to.
	Channel PreferenceChannel
	// Handle is who the statement is about: an email address on the email
	// channel, an E.164 phone number on SMS and WhatsApp.
	Handle string
	// Status is what the statement says: granted or revoked.
	Status PreferenceStatus
	// Coverage is how much traffic the statement covers. Defaults to
	// non_transactional server-side when left unset.
	Coverage PreferenceCoverage
	// SenderScope limits the statement to one sender instead of the whole
	// channel. Not supported on email.
	SenderScope string
	// Source is a free-form note on where the statement came from: a form
	// name, an import batch, a campaign.
	Source string
	// ConsentedAt is when the person consented, required evidence when
	// granting over a stored opt-out: the grant applies only if this is later
	// than the opt-out it reverses. Omitted from the request when zero.
	ConsentedAt time.Time
}

func (p PreferencesCreateParams) toWire() oapi.PreferenceCreate {
	return oapi.PreferenceCreate{
		Channel:     p.Channel,
		Handle:      p.Handle,
		Status:      p.Status,
		Coverage:    optZero(p.Coverage),
		SenderScope: optStr(p.SenderScope),
		Source:      optStr(p.Source),
		ConsentedAt: optTime(p.ConsentedAt),
	}
}

// Create records a preference statement: a consent grant or opt-out for a
// handle on a channel, optionally scoped to one sender. The write is an
// ordered upsert keyed by channel, handle, and sender scope: a statement
// older than the one already on record is not an error, it comes back with
// Applied false and Preference set to the statement that survived. The HTTP
// status (200 for an existing key, 201 for a fresh one) does not change what
// is returned, so callers never need to branch on it. Retried safely with a
// reused idempotency key.
func (s *PreferencesService) Create(ctx context.Context, params PreferencesCreateParams, opts ...option.RequestOption) (*PreferenceWriteResult, error) {
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreatePreferenceParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreatePreference(ctx, op, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out PreferenceWriteResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a preference statement by ID. The delete is conditional: if
// a newer statement has since been recorded on the same key, the delete is
// refused rather than applied, and the returned write result carries Applied
// false with Preference set to the statement that survived — a person's own
// opt-out is refused the same way, since it cannot be overridden or deleted
// through this API. Applied true with a nil Preference means the key now has
// no record. Retried safely with a reused idempotency key.
func (s *PreferencesService) Delete(ctx context.Context, preferenceId string, opts ...option.RequestOption) (*PreferenceWriteResult, error) {
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.DeletePreferenceParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DeletePreference(ctx, oapi.PreferenceID(preferenceId), op, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out PreferenceWriteResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
