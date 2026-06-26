package zip

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// ZipError represents an error produced by ZIP helpers.
type ZipError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the ZIP error message.
func (e *ZipError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *ZipError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *ZipError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *ZipError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*ZipError); ok {
		return e.Code == other.Code
	}
	return false
}

func zipErrorf(code knifer.ErrCode, format string, args ...any) *ZipError {
	return &ZipError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func wrapZipError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &ZipError{Code: code, Msg: msg, Cause: cause}
}

func invalidInputf(format string, args ...any) *ZipError {
	return zipErrorf(knifer.ErrCodeInvalidInput, format, args...)
}

func notFound(msg string, cause error) error {
	return wrapZipError(knifer.ErrCodeNotFound, msg, cause)
}
