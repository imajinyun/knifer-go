package shared

import (
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
)

// HTTPError represents an error during HTTP operations.
type HTTPError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the error message.
func (e *HTTPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
	}
	return e.Msg
}

// Unwrap returns the underlying error.
func (e *HTTPError) Unwrap() error { return e.Cause }

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *HTTPError) Is(target error) bool {
	code, ok := target.(knifer.ErrCode)
	return ok && e.Code == code
}

// NewHTTPError creates an HTTP error.
func NewHTTPError(msg string, cause error) *HTTPError {
	return &HTTPError{Code: knifer.ErrCodeInternal, Msg: msg, Cause: cause}
}

// HTTPErrorf creates an HTTP error with a formatted message.
func HTTPErrorf(format string, args ...any) *HTTPError {
	return &HTTPError{Code: knifer.ErrCodeInternal, Msg: fmt.Sprintf(format, args...)}
}
