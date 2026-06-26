package xml

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// XMLError represents an error produced by XML helpers.
type XMLError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the XML error message.
func (e *XMLError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *XMLError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *XMLError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *XMLError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*XMLError); ok {
		return e.Code == other.Code
	}
	return false
}

func xmlErrorf(code knifer.ErrCode, format string, args ...any) *XMLError {
	return &XMLError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func wrapXMLError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &XMLError{Code: code, Msg: msg, Cause: cause}
}

func invalidInputf(format string, args ...any) *XMLError {
	return xmlErrorf(knifer.ErrCodeInvalidInput, format, args...)
}

func wrapInvalidInput(msg string, cause error) error {
	return wrapXMLError(knifer.ErrCodeInvalidInput, msg, cause)
}

func wrapInternal(msg string, cause error) error {
	return wrapXMLError(knifer.ErrCodeInternal, msg, cause)
}
