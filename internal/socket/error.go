package socket

import (
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
)

// SocketRuntimeError represents a runtime error during socket communication.
type SocketRuntimeError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the error message.
func (e *SocketRuntimeError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil && e.Msg == "" {
		return e.Cause.Error()
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// Unwrap supports errors.Is and errors.As.
func (e *SocketRuntimeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *SocketRuntimeError) Is(target error) bool {
	code, ok := target.(knifer.ErrCode)
	return ok && e.Code == code
}

// NewSocketError creates a SocketRuntimeError from any error.
func NewSocketError(err error) *SocketRuntimeError {
	if err == nil {
		return nil
	}
	return &SocketRuntimeError{Code: knifer.ErrCodeInternal, Msg: err.Error(), Cause: err}
}

// NewSocketErrorMsg creates a SocketRuntimeError from a message.
func NewSocketErrorMsg(msg string) *SocketRuntimeError {
	return &SocketRuntimeError{Code: knifer.ErrCodeInternal, Msg: msg}
}

// NewSocketErrorf creates a formatted SocketRuntimeError.
func NewSocketErrorf(format string, args ...any) *SocketRuntimeError {
	return &SocketRuntimeError{Code: knifer.ErrCodeInternal, Msg: fmt.Sprintf(format, args...)}
}

// WrapSocketError wraps an underlying error with an additional message.
func WrapSocketError(err error, msg string) *SocketRuntimeError {
	if err == nil {
		return nil
	}
	return &SocketRuntimeError{Code: knifer.ErrCodeInternal, Msg: msg, Cause: err}
}
