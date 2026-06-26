package bean

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// BeanError represents an error produced by bean mapping helpers.
type BeanError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the bean error message.
func (e *BeanError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *BeanError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *BeanError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *BeanError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*BeanError); ok {
		return e.Code == other.Code
	}
	return false
}

func beanErrorf(code knifer.ErrCode, format string, args ...any) *BeanError {
	return &BeanError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func wrapBeanError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &BeanError{Code: code, Msg: msg, Cause: cause}
}

func invalidBeanInputf(format string, args ...any) *BeanError {
	return beanErrorf(knifer.ErrCodeInvalidInput, format, args...)
}

func wrapBeanInput(msg string, cause error) error {
	return wrapBeanError(knifer.ErrCodeInvalidInput, msg, cause)
}
