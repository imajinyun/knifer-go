package geo

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// Error represents an error produced by geo helpers.
type Error struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

func (e *Error) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*Error); ok {
		return e.Code == other.Code
	}
	return false
}

func invalidInputf(format string, args ...any) *Error {
	return &Error{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...)}
}
