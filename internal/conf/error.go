package conf

import (
	"fmt"
	"os"

	knifer "github.com/imajinyun/knifer-go"
)

// ConfError represents an error produced by configuration helpers.
type ConfError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the configuration error message.
func (e *ConfError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *ConfError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *ConfError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *ConfError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*ConfError); ok {
		return e.Code == other.Code
	}
	return false
}

func confErrorf(code knifer.ErrCode, format string, args ...any) *ConfError {
	return &ConfError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func wrapConfError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &ConfError{Code: code, Msg: msg, Cause: cause}
}

func invalidInputf(format string, args ...any) *ConfError {
	return confErrorf(knifer.ErrCodeInvalidInput, format, args...)
}

func notFoundf(format string, args ...any) *ConfError {
	return confErrorf(knifer.ErrCodeNotFound, format, args...)
}

func wrapConfigIO(msg string, cause error) error {
	code := knifer.ErrCodeInternal
	if os.IsNotExist(cause) {
		code = knifer.ErrCodeNotFound
	} else if causeCode, ok := knifer.CodeOf(cause); ok && causeCode != "" {
		code = causeCode
	}
	return wrapConfError(code, msg, cause)
}

func wrapConfigParse(msg string, cause error) error {
	return wrapConfError(knifer.ErrCodeInvalidInput, msg, cause)
}
