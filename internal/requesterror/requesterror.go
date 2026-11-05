// Package requesterror keeps request validation independent of response decoding.
package requesterror

// RequestValidationError reports an unset required field before a request is sent.
type RequestValidationError struct {
	Field string
}

func (e *RequestValidationError) Error() string {
	return "bird: required field " + e.Field + " missing"
}
