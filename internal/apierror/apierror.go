// Package apierror holds the SDK's error model: the wire-error mapping and the
// typed error hierarchy returned to callers. It is re-exported by the bird
// package (bird.APIError, bird.RateLimitError, …), so these names are the
// semver-locked public surface.
package apierror

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ErrorType is the coarse error category clients branch on (ADR-0016). Branch
// on APIError.Type for the flat categories, or use errors.As for the variants
// that carry extra data (RateLimitError, ValidationError).
type ErrorType string

const (
	ErrorTypeBadRequest         ErrorType = "bad_request_error"
	ErrorTypeAuth               ErrorType = "auth_error"
	ErrorTypeBilling            ErrorType = "billing_error"
	ErrorTypePermission         ErrorType = "permission_error"
	ErrorTypeNotFound           ErrorType = "not_found_error"
	ErrorTypeConflict           ErrorType = "conflict_error"
	ErrorTypePrecondition       ErrorType = "precondition_error"
	ErrorTypePayloadTooLarge    ErrorType = "payload_too_large_error"
	ErrorTypeMisdirected        ErrorType = "misdirected_error"
	ErrorTypeValidation         ErrorType = "validation_error"
	ErrorTypeRateLimit          ErrorType = "rate_limit_error"
	ErrorTypeInternal           ErrorType = "internal_error"
	ErrorTypeNotImplemented     ErrorType = "not_implemented_error"
	ErrorTypeServiceUnavailable ErrorType = "service_unavailable_error"
)

// APIError is the error the server returned. It is the base for every API
// failure; catch it with errors.As to handle any server error uniformly.
type APIError struct {
	StatusCode int
	Type       ErrorType
	// Code is the stable, opaque error code (E#####).
	Code string
	// Name is a human-readable slug for logs; paired with Code, never replaces it.
	Name string
	// Message is the human-readable description.
	Message string
	// DocURL links to the documentation for this Code.
	DocURL string
	// RequestID correlates with the X-Request-Id response header.
	RequestID string
	// Param is the offending field, when applicable.
	Param string
	// VendorCode is a verbatim code from a downstream system (SMTP reply, decline).
	VendorCode string
	// Remediation is a human-readable next step to resolve this error, when a
	// recovery is known (ADR-0073).
	Remediation string
	// Next lists the operations that resolve this error, in the order to try them.
	Next []ErrorNextAction
	// UnmetGates lists the verification requirements blocking this action, each
	// with the flow that resolves it. Present only when an action is blocked
	// pending verification.
	UnmetGates []UnmetGate
}

func (e *APIError) Error() string {
	return fmt.Sprintf("bird: %s (status %d, type %s, request_id %s)", e.Message, e.StatusCode, e.Type, e.RequestID)
}

// RateLimitError is a 429. RetryAfter is the server-advised wait, or zero when
// none was advised.
type RateLimitError struct {
	*APIError
	RetryAfter time.Duration
}

func (e *RateLimitError) Unwrap() error { return e.APIError }

// ErrorDetail is one per-field validation failure.
type ErrorDetail struct {
	// Param is the dotted field path, e.g. "to[0]", "subject".
	Param string `json:"param"`
	// Message is what is wrong with this field.
	Message string `json:"message"`
}

// ErrorNextAction is one recovery operation the server suggests (ADR-0073): call
// it to resolve the error, then retry the original request.
type ErrorNextAction struct {
	// Operation is the operationId of the follow-up operation that resolves this error.
	Operation string `json:"operation"`
	// Description is a short human-readable label for the recovery step.
	Description string `json:"description,omitempty"`
	// Scope is the permission scope the recovery operation requires, when it is scoped.
	Scope string `json:"scope,omitempty"`
}

// UnmetGate is one verification requirement blocking the action, with the flow
// that resolves it. Present on APIError.UnmetGates when an action is blocked
// pending verification.
type UnmetGate struct {
	// Slug is the stable identifier for the verification requirement.
	Slug string `json:"slug"`
	// Name is the human-readable name of the verification requirement.
	Name string `json:"name"`
	// Status is the requirement's current state.
	Status string `json:"status"`
	// RemediationKind is how to resolve this requirement.
	RemediationKind string `json:"remediation_kind"`
}

// ValidationError is a 422; Details carries the per-field failures.
type ValidationError struct {
	*APIError
	Details []ErrorDetail
}

func (e *ValidationError) Unwrap() error { return e.APIError }

// ConnectionError is a network-level failure with no HTTP response (DNS,
// connection refused, socket hangup).
type ConnectionError struct{ Err error }

func (e *ConnectionError) Error() string { return "bird: connection error: " + e.Err.Error() }
func (e *ConnectionError) Unwrap() error { return e.Err }

// TimeoutError means a single attempt exceeded its per-attempt timeout.
type TimeoutError struct{ Timeout time.Duration }

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("bird: request timed out after %s", e.Timeout)
}

// WebhookVerificationError means a webhook payload failed signature
// verification (bad signature, stale timestamp, malformed headers).
type WebhookVerificationError struct{ Reason string }

func (e *WebhookVerificationError) Error() string {
	return "bird: webhook verification failed: " + e.Reason
}

// wireError is the on-the-wire error body (snake_case as sent).
type wireError struct {
	Type        string            `json:"type"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Message     string            `json:"message"`
	DocURL      string            `json:"doc_url"`
	RequestID   string            `json:"request_id"`
	Param       string            `json:"param"`
	VendorCode  string            `json:"vendor_code"`
	Details     []ErrorDetail     `json:"details"`
	Remediation string            `json:"remediation"`
	Next        []ErrorNextAction `json:"next"`
	UnmetGates  []UnmetGate       `json:"unmet_gates"`
}

// errorEnvelope is the wire wrapper: the API sends every error as
// {"error": {...}} (the OpenAPI Error schema).
type errorEnvelope struct {
	Error *wireError `json:"error"`
}

// FromResponse turns a terminal non-2xx response into a typed error. It is the
// single place a wire error becomes a Go error. The API wraps errors as
// {"error": {...}}; a bare top-level body and a non-JSON body are both
// tolerated.
func FromResponse(status int, body []byte, header http.Header) error {
	var w wireError
	if env := (errorEnvelope{}); json.Unmarshal(body, &env) == nil && env.Error != nil {
		w = *env.Error // the wrapped {"error": {...}} envelope
	} else {
		_ = json.Unmarshal(body, &w) // bare top-level body, or a non-JSON body (proxy 502s, etc.)
	}

	typ := ErrorType(w.Type)
	if typ == "" {
		typ = inferType(status)
	}
	requestID := w.RequestID
	if requestID == "" {
		requestID = header.Get("X-Request-Id")
	}
	message := w.Message
	if message == "" {
		message = fmt.Sprintf("request failed with status %d", status)
	}

	base := &APIError{
		StatusCode:  status,
		Type:        typ,
		Code:        w.Code,
		Name:        w.Name,
		Message:     message,
		DocURL:      w.DocURL,
		RequestID:   requestID,
		Param:       w.Param,
		VendorCode:  w.VendorCode,
		Remediation: w.Remediation,
		Next:        w.Next,
		UnmetGates:  w.UnmetGates,
	}

	switch typ {
	case ErrorTypeRateLimit:
		retryAfter, _ := ParseRetryAfter(header)
		return &RateLimitError{APIError: base, RetryAfter: retryAfter}
	case ErrorTypeValidation:
		return &ValidationError{APIError: base, Details: w.Details}
	default:
		return base
	}
}

// inferType maps a status to a type for error bodies that carry none.
func inferType(status int) ErrorType {
	switch status {
	case http.StatusBadRequest:
		return ErrorTypeBadRequest
	case http.StatusUnauthorized:
		return ErrorTypeAuth
	case http.StatusPaymentRequired:
		return ErrorTypeBilling
	case http.StatusForbidden:
		return ErrorTypePermission
	case http.StatusNotFound:
		return ErrorTypeNotFound
	case http.StatusConflict:
		return ErrorTypeConflict
	case http.StatusPreconditionFailed, http.StatusPreconditionRequired:
		return ErrorTypePrecondition
	case http.StatusRequestEntityTooLarge:
		return ErrorTypePayloadTooLarge
	case http.StatusMisdirectedRequest:
		return ErrorTypeMisdirected
	case http.StatusUnprocessableEntity:
		return ErrorTypeValidation
	case http.StatusTooManyRequests:
		return ErrorTypeRateLimit
	case http.StatusNotImplemented:
		return ErrorTypeNotImplemented
	case http.StatusServiceUnavailable:
		return ErrorTypeServiceUnavailable
	default:
		if status >= 500 {
			return ErrorTypeInternal
		}
		return ErrorTypeBadRequest
	}
}

// ParseRetryAfter reads a Retry-After header (delta-seconds or HTTP-date) as a
// duration. A negative or unparseable value reports ok=false — a negative wait
// is meaningless.
func ParseRetryAfter(header http.Header) (time.Duration, bool) {
	v := header.Get("Retry-After")
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d >= 0 {
			return d, true
		}
	}
	return 0, false
}
