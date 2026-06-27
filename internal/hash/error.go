package hash

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// Error represents an error produced by hash helpers.
type Error struct {
	Code knifer.ErrCode
	Msg  string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func (e *Error) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
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
