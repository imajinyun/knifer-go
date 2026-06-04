package cron

import (
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
)

// CronError is aligned with the utility toolkit CronException and represents cron-related errors.
type CronError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// NewCronError creates an error with a formatted message.
func NewCronError(format string, args ...any) *CronError {
	return &CronError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...)}
}

// WrapCronError wraps an underlying error with a formatted message.
func WrapCronError(cause error, format string, args ...any) *CronError {
	return &CronError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...), Cause: cause}
}

func (e *CronError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
	}
	return e.Msg
}

// Unwrap supports errors.Is and errors.As.
func (e *CronError) Unwrap() error { return e.Cause }

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *CronError) Is(target error) bool {
	code, ok := target.(knifer.ErrCode)
	return ok && e.Code == code
}
