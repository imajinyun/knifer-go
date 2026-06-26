package shared

import (
	"errors"
	"fmt"
	"net"
	"os"

	knifer "github.com/imajinyun/knifer-go"
)

// HTTPError represents an error during HTTP operations.
type HTTPError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the error message.
func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *HTTPError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying error.
func (e *HTTPError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *HTTPError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*HTTPError); ok {
		return e.Code == other.Code
	}
	return false
}

// NewHTTPError creates an HTTP error.
func NewHTTPError(msg string, cause error) *HTTPError {
	return NewHTTPErrorWithCode(ClassifyHTTPErrorCode(cause, knifer.ErrCodeInternal), msg, cause)
}

// HTTPErrorf creates an HTTP error with a formatted message.
func HTTPErrorf(format string, args ...any) *HTTPError {
	return NewHTTPErrorWithCode(knifer.ErrCodeInternal, fmt.Sprintf(format, args...), nil)
}

// HTTPErrorfWithCode creates an HTTP error with an explicit code and formatted message.
func HTTPErrorfWithCode(code knifer.ErrCode, format string, args ...any) *HTTPError {
	return NewHTTPErrorWithCode(code, fmt.Sprintf(format, args...), nil)
}

// NewHTTPErrorWithCode creates an HTTP error with an explicit code.
func NewHTTPErrorWithCode(code knifer.ErrCode, msg string, cause error) *HTTPError {
	if code == "" {
		code = knifer.ErrCodeInternal
	}
	return &HTTPError{Code: code, Msg: msg, Cause: cause}
}

// ClassifyHTTPErrorCode maps common transport errors to knifer-go error codes.
func ClassifyHTTPErrorCode(err error, fallback knifer.ErrCode) knifer.ErrCode {
	if err == nil {
		if fallback == "" {
			return knifer.ErrCodeInternal
		}
		return fallback
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return knifer.ErrCodeTimeout
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return knifer.ErrCodeTimeout
	}
	if code, ok := knifer.CodeOf(err); ok {
		return code
	}
	if fallback == "" {
		return knifer.ErrCodeInternal
	}
	return fallback
}
