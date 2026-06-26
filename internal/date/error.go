package date

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// DateError represents an error produced by date helpers.
type DateError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the date error message.
func (e *DateError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *DateError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *DateError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *DateError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*DateError); ok {
		return e.Code == other.Code
	}
	return false
}

func dateErrorf(code knifer.ErrCode, format string, args ...any) *DateError {
	return &DateError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func wrapDateError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &DateError{Code: code, Msg: msg, Cause: cause}
}

func invalidDateInputf(format string, args ...any) *DateError {
	return dateErrorf(knifer.ErrCodeInvalidInput, format, args...)
}

func wrapDateParse(msg string, cause error) error {
	return wrapDateError(knifer.ErrCodeInvalidInput, msg, cause)
}
