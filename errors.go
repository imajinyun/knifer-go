package knifer

import (
	"errors"
	"fmt"
)

// ErrCode is a stable, cross-subpackage error classifier.
//
// It is itself an error so that it can be used directly with errors.Is:
//
//	if errors.Is(err, knifer.ErrCodeInvalidInput) { ... }
type ErrCode string

// Error implements the error interface so an ErrCode constant can be used
// as the target of errors.Is without wrapping.
func (c ErrCode) Error() string { return string(c) }

// Predefined cross-cutting error codes.
//
// These cover the most common failure categories shared across subpackages.
// Domain-specific codes can be defined within each subpackage when needed.
const (
	// ErrCodeInvalidInput indicates the caller provided invalid arguments.
	ErrCodeInvalidInput ErrCode = "GK_INVALID_INPUT"
	// ErrCodeNotFound indicates the requested resource does not exist.
	ErrCodeNotFound ErrCode = "GK_NOT_FOUND"
	// ErrCodeUnsupported indicates the requested operation is not supported.
	ErrCodeUnsupported ErrCode = "GK_UNSUPPORTED"
	// ErrCodeTimeout indicates the operation exceeded its time budget.
	ErrCodeTimeout ErrCode = "GK_TIMEOUT"
	// ErrCodeInternal indicates an unexpected internal failure.
	ErrCodeInternal ErrCode = "GK_INTERNAL"
)

// Error is the unified error type for go-knifer subpackages.
//
// Subpackages that want to participate in the cross-cutting error contract
// should return *Error and wrap any underlying cause through Cause so that
// callers can rely on errors.Is(err, knifer.ErrCodeXxx) and on the standard
// error-chain helpers.
type Error struct {
	// Code classifies the error. It is matched by errors.Is.
	Code ErrCode
	// Message is a human-readable description.
	Message string
	// Cause is the underlying error, if any. It is exposed through Unwrap.
	Cause error
}

// CodeCarrier is implemented by errors that can expose a go-knifer error code.
//
// Custom subpackage errors can implement this interface while preserving their
// own concrete type and sentinel semantics.
type CodeCarrier interface {
	ErrorCode() ErrCode
}

// Error returns "CODE: Message" or "CODE: Message: cause" when Cause is set.
func (e *Error) Error() string {
	switch {
	case e == nil:
		return ""
	case e.Cause != nil && e.Message != "":
		return fmt.Sprintf("%s: %s: %s", e.Code, e.Message, e.Cause.Error())
	case e.Cause != nil:
		return fmt.Sprintf("%s: %s", e.Code, e.Cause.Error())
	default:
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
}

// Unwrap returns the underlying cause, enabling errors.Is and errors.As to
// traverse the error chain.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// ErrorCode returns the error code carried by e.
func (e *Error) ErrorCode() ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Is reports whether target matches this error.
//
// It returns true when target is the same ErrCode value or another *Error with
// the same Code. The standard library walks the chain via Unwrap, so callers
// can write errors.Is(err, knifer.ErrCodeInvalidInput).
func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*Error); ok {
		return e.Code == other.Code
	}
	return false
}

// NewError builds an *Error with the given code and message.
func NewError(code ErrCode, message string) *Error {
	return &Error{Code: code, Message: message}
}

// WrapError builds an *Error that wraps cause; cause is preserved on the
// chain and remains discoverable via errors.Is / errors.As.
func WrapError(code ErrCode, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// Errorf builds an *Error whose message is formatted with fmt.Sprintf.
func Errorf(code ErrCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// CodeOf extracts a go-knifer error code from err.
//
// It first looks for an error implementing CodeCarrier, then falls back to the
// predefined base ErrCode values through errors.Is. The fallback keeps sentinel
// errors that only implement Is(target ErrCode) discoverable.
func CodeOf(err error) (ErrCode, bool) {
	if err == nil {
		return "", false
	}
	var carrier CodeCarrier
	if errors.As(err, &carrier) {
		if code := carrier.ErrorCode(); code != "" {
			return code, true
		}
	}
	for _, code := range []ErrCode{
		ErrCodeInvalidInput,
		ErrCodeNotFound,
		ErrCodeUnsupported,
		ErrCodeTimeout,
		ErrCodeInternal,
	} {
		if errors.Is(err, code) {
			return code, true
		}
	}
	return "", false
}
