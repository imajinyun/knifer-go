package jwt

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// JWTError is a JWT-related error.
type JWTError struct {
	Code knifer.ErrCode
	Msg  string
	Err  error
}

// Error implements the error interface.
func (e *JWTError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *JWTError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the wrapped error.
func (e *JWTError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *JWTError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*JWTError); ok {
		return e.Code == other.Code
	}
	return false
}

// NewJWTError creates an error.
func NewJWTError(msg string) *JWTError {
	return &JWTError{Code: knifer.ErrCodeInvalidInput, Msg: msg}
}

// JWTErrorf creates a formatted error.
func JWTErrorf(format string, args ...any) *JWTError {
	return &JWTError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...)}
}

func wrapJWTError(cause error, msg string) *JWTError {
	return &JWTError{Code: knifer.ErrCodeInvalidInput, Msg: msg, Err: cause}
}

func unsupportedJWTErrorf(format string, args ...any) *JWTError {
	return &JWTError{Code: knifer.ErrCodeUnsupported, Msg: fmt.Sprintf(format, args...)}
}
