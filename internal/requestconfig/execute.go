package requestconfig

import (
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	rand "math/rand/v2"
	"net/http"
	"time"

	"github.com/messagebird/bird-sdk-go/internal/apierror"
)

const (
	backoffBase   = 500 * time.Millisecond
	backoffCap    = 8 * time.Second
	retryAfterCap = 60 * time.Second
)

// Attempt performs one network round-trip with the per-attempt context and the
// idempotency key to set as a header.
type Attempt func(ctx context.Context, idempotencyKey string) (*http.Response, error)

// Execute runs the request lifecycle: for a mutation, a single idempotency key
// generated once and reused across attempts; a per-attempt timeout; and retries
// of transient failures with jittered backoff honoring Retry-After. It returns
// the drained 2xx body, or a typed *apierror error, and writes transport
// metadata to cfg.ResponseInto when set.
func (cfg Config) Execute(ctx context.Context, mutation bool, call Attempt) ([]byte, error) {
	idempotencyKey := cfg.IdempotencyKey
	if idempotencyKey == "" && mutation {
		key, err := newIdempotencyKey()
		if err != nil {
			return nil, err
		}
		idempotencyKey = key
	}

	for attemptNum := 0; ; attemptNum++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		attemptCtx := ctx
		var cancel context.CancelFunc
		if cfg.Timeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		}

		resp, err := call(attemptCtx, idempotencyKey)
		if err != nil {
			if cancel != nil {
				cancel()
			}
			if ctx.Err() != nil { // caller cancellation wins, surfaced verbatim
				return nil, ctx.Err()
			}
			if attemptNum >= cfg.MaxRetries {
				return nil, classifyTransport(err, cfg.Timeout)
			}
			if !sleep(ctx, backoff(attemptNum)) {
				return nil, ctx.Err()
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if cancel != nil {
			cancel()
		}
		if readErr != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if attemptNum >= cfg.MaxRetries {
				return nil, &apierror.ConnectionError{Err: readErr}
			}
			if !sleep(ctx, backoff(attemptNum)) {
				return nil, ctx.Err()
			}
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			cfg.fillResponse(resp)
			return body, nil
		}
		if !isRetryable(resp.StatusCode) || attemptNum >= cfg.MaxRetries {
			return nil, apierror.FromResponse(resp.StatusCode, body, resp.Header)
		}
		if !sleep(ctx, retryDelay(attemptNum, resp.Header)) {
			return nil, ctx.Err()
		}
	}
}

func (cfg Config) fillResponse(resp *http.Response) {
	if cfg.ResponseInto == nil || resp == nil {
		return
	}
	*cfg.ResponseInto = Response{
		Status:    resp.StatusCode,
		Header:    resp.Header,
		RequestID: resp.Header.Get("X-Request-Id"),
	}
}

// isRetryable reports whether a status is worth retrying. 409 is a semantic
// conflict a retry cannot resolve; other 4xx are deterministic.
func isRetryable(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func classifyTransport(err error, timeout time.Duration) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &apierror.TimeoutError{Timeout: timeout}
	}
	return &apierror.ConnectionError{Err: err}
}

// backoff is full-jitter exponential backoff: random in [0, min(cap, base·2^n)].
func backoff(attemptNum int) time.Duration {
	ceiling := backoffCap
	if attemptNum < 30 {
		if scaled := backoffBase << uint(attemptNum); scaled > 0 && scaled < backoffCap {
			ceiling = scaled
		}
	}
	return time.Duration(rand.Int64N(int64(ceiling) + 1))
}

// retryDelay honors Retry-After on a retryable response, else falls back to backoff.
func retryDelay(attemptNum int, header http.Header) time.Duration {
	if d, ok := apierror.ParseRetryAfter(header); ok {
		return min(d, retryAfterCap)
	}
	return backoff(attemptNum)
}

// sleep waits for d, returning false if the context is cancelled first.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// newIdempotencyKey returns a random UUIDv4.
func newIdempotencyKey() (string, error) {
	var b [16]byte
	if _, err := crand.Read(b[:]); err != nil {
		return "", fmt.Errorf("bird: generating idempotency key: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
