package bird

import (
	"context"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// resource is the base every service embeds: it holds the client and the
// get/post request-core helpers the generated methods call, so the request
// lifecycle lives in one place instead of being copied per resource.
type resource struct{ client *Client }

// requestConfig is the per-request editor slice a call hands to the generated
// low-level client.
type requestConfig = []oapi.RequestEditorFn

// get runs a read through the shared request core: resolve per-call options,
// then execute (never a mutation, so no idempotency key) and return the decoded
// response body.
func (r resource) get(ctx context.Context, opts []option.RequestOption, call func(context.Context, requestConfig) (*http.Response, error)) ([]byte, error) {
	return r.getAuth(ctx, opts, nil, call)
}

// getAuth is get for an operation that also authenticates with the extra
// credentials named by schemes.
func (r resource) getAuth(ctx context.Context, opts []option.RequestOption, schemes []string, call func(context.Context, requestConfig) (*http.Response, error)) ([]byte, error) {
	cfg, err := r.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	editors, err := r.client.credentialEditors(cfg, schemes)
	if err != nil {
		return nil, err
	}
	return cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return call(ctx, editors)
	})
}

// getRedirect is get for an operation whose success is a redirect: status counts
// as success alongside 2xx, and the transport metadata comes back so the caller
// can read Location and fetch it with a client that carries no credentials.
func (r resource) getRedirect(ctx context.Context, opts []option.RequestOption, status int, call func(context.Context, requestConfig) (*http.Response, error)) ([]byte, Response, error) {
	var meta Response
	cfg, err := r.client.resolve(opts)
	if err != nil {
		return nil, meta, err
	}
	if cfg.APIKey == "" {
		return nil, meta, ErrMissingAPIKey
	}
	editors, err := r.client.credentialEditors(cfg, nil)
	if err != nil {
		return nil, meta, err
	}
	// ResponseInto is the caller's own handle when they passed
	// option.WithResponseInto, so borrow the field and hand their copy back
	// afterwards rather than dropping what they asked for.
	caller := cfg.ResponseInto
	cfg.ResponseInto = &meta
	cfg.SuccessStatus = status
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return call(ctx, editors)
	})
	if caller != nil {
		*caller = meta
	}
	return body, meta, err
}

// post runs a write through the shared request core: resolve options, then
// execute with one idempotency key reused across every retry attempt (the hard
// lifecycle invariant), and return the decoded response body.
func (r resource) post(ctx context.Context, opts []option.RequestOption, call func(context.Context, string, requestConfig) (*http.Response, error)) ([]byte, error) {
	return r.postAuth(ctx, opts, nil, call)
}

// postAuth is post for an operation that also authenticates with the extra
// credentials named by schemes.
func (r resource) postAuth(ctx context.Context, opts []option.RequestOption, schemes []string, call func(context.Context, string, requestConfig) (*http.Response, error)) ([]byte, error) {
	cfg, err := r.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	editors, err := r.client.credentialEditors(cfg, schemes)
	if err != nil {
		return nil, err
	}
	return cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		return call(ctx, idempotencyKey, editors)
	})
}
