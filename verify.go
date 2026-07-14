package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// VerifyService is the Verify product namespace. Reach it via Client.Verify.
type VerifyService struct {
	// Verifications starts verifications and checks the passcodes recipients submit.
	Verifications *VerificationsService
}

// VerificationsService starts a verification — sending a one-time passcode — and
// checks the passcode a recipient submits.
type VerificationsService struct{ client *Client }

// VerificationCreateParams starts a verification. Provide Email, Phone, or both;
// zero-value fields are omitted from the request.
type VerificationCreateParams struct {
	Email      string         // recipient email address, verified over email
	Phone      string         // recipient phone number in E.164, verified over SMS
	CodeLength int            // passcode length 4–8, overriding the configured length
	Channels   []string       // delivery channels to try, in order; empty uses the configured order
	Metadata   map[string]any // arbitrary key/value pairs stored on the verification
}

func (p VerificationCreateParams) toWire() oapi.VerificationCreateRequest {
	var body oapi.VerificationCreateRequest
	if p.Email != "" {
		email := openapi_types.Email(p.Email)
		body.To.EmailAddress = &email
	}
	if p.Phone != "" {
		phone := p.Phone
		body.To.PhoneNumber = &phone
	}
	if p.CodeLength > 0 || len(p.Channels) > 0 {
		var opts oapi.VerificationOptions
		if p.CodeLength > 0 {
			codeLength := p.CodeLength
			opts.CodeLength = &codeLength
		}
		if len(p.Channels) > 0 {
			channels := make([]oapi.VerificationChannel, len(p.Channels))
			copy(channels, p.Channels)
			opts.Channels = &channels
		}
		body.Options = &opts
	}
	if len(p.Metadata) > 0 {
		metadata := p.Metadata
		body.Metadata = &metadata
	}
	return body
}

// Create starts a verification and sends a one-time passcode to the recipient.
// It is also the resend: starting again for the same recipient re-sends the code
// after the cooldown rather than opening a second verification. Retried safely — a
// single idempotency key is reused across attempts.
func (s *VerificationsService) Create(ctx context.Context, params VerificationCreateParams, opts ...option.RequestOption) (*Verification, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateVerificationParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateVerification(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Verification
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerificationCheckParams checks a submitted passcode. Identify the verification by
// the same recipient it was started for.
type VerificationCheckParams struct {
	Email string // recipient email the verification was started for
	Phone string // recipient phone the verification was started for
	Code  string // the passcode the recipient submitted
}

func (p VerificationCheckParams) toWire() oapi.VerificationCheckRequest {
	body := oapi.VerificationCheckRequest{Code: p.Code}
	if p.Email != "" {
		email := openapi_types.Email(p.Email)
		body.To.EmailAddress = &email
	}
	if p.Phone != "" {
		phone := p.Phone
		body.To.PhoneNumber = &phone
	}
	return body
}

// Check checks a passcode a recipient submitted. A wrong, expired, or already-used
// code returns a result with Success false and a Reason — not an error. Retried
// safely with a reused idempotency key.
func (s *VerificationsService) Check(ctx context.Context, params VerificationCheckParams, opts ...option.RequestOption) (*VerificationCheckResult, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateVerificationCheckParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateVerificationCheck(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out VerificationCheckResult
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
