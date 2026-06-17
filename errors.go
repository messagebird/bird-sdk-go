package bird

import "github.com/messagebird/bird-sdk-go/internal/apierror"

// The SDK error model, re-exported from internal/apierror so these names are
// the semver-locked public surface. Catch *APIError (via errors.As) to handle
// any server error; the variants carry extra data. Transport failures with no
// HTTP response are *ConnectionError / *TimeoutError; a bad webhook signature is
// *WebhookVerificationError.
type (
	APIError                 = apierror.APIError
	RateLimitError           = apierror.RateLimitError
	ValidationError          = apierror.ValidationError
	ConnectionError          = apierror.ConnectionError
	TimeoutError             = apierror.TimeoutError
	WebhookVerificationError = apierror.WebhookVerificationError
	ErrorDetail              = apierror.ErrorDetail
	ErrorType                = apierror.ErrorType
)

// ErrorType values — the coarse categories clients branch on (ADR-0016).
const (
	ErrorTypeBadRequest         = apierror.ErrorTypeBadRequest
	ErrorTypeAuth               = apierror.ErrorTypeAuth
	ErrorTypeBilling            = apierror.ErrorTypeBilling
	ErrorTypePermission         = apierror.ErrorTypePermission
	ErrorTypeNotFound           = apierror.ErrorTypeNotFound
	ErrorTypeConflict           = apierror.ErrorTypeConflict
	ErrorTypePrecondition       = apierror.ErrorTypePrecondition
	ErrorTypePayloadTooLarge    = apierror.ErrorTypePayloadTooLarge
	ErrorTypeMisdirected        = apierror.ErrorTypeMisdirected
	ErrorTypeValidation         = apierror.ErrorTypeValidation
	ErrorTypeRateLimit          = apierror.ErrorTypeRateLimit
	ErrorTypeInternal           = apierror.ErrorTypeInternal
	ErrorTypeNotImplemented     = apierror.ErrorTypeNotImplemented
	ErrorTypeServiceUnavailable = apierror.ErrorTypeServiceUnavailable
)
