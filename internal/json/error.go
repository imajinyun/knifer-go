package json

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// JSONError matches the utility JSONException.
type JSONError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// NewJSONError creates an error with a message.
func NewJSONError(format string, args ...any) *JSONError {
	return &JSONError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...)}
}

// WrapJSONError wraps an underlying error.
func WrapJSONError(cause error, format string, args ...any) *JSONError {
	return &JSONError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...), Cause: cause}
}

func (e *JSONError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *JSONError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap supports errors.Is and errors.As.
func (e *JSONError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *JSONError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*JSONError); ok {
		return e.Code == other.Code
	}
	return false
}
