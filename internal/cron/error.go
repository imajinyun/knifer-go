package cron

import (
	"errors"
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
)

// ErrSchedulerStarted is returned when immutable scheduler configuration is changed after Start.
var ErrSchedulerStarted = errors.New("scheduler already started")

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

func newSchedulerStartedError() *CronError {
	return &CronError{Code: knifer.ErrCodeUnsupported, Msg: "cannot change scheduler config", Cause: ErrSchedulerStarted}
}

func (e *CronError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
	}
	return e.Msg
}

// ErrorCode returns the go-knifer error code.
func (e *CronError) ErrorCode() knifer.ErrCode { return e.Code }

// Unwrap supports errors.Is and errors.As.
func (e *CronError) Unwrap() error { return e.Cause }

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *CronError) Is(target error) bool {
	code, ok := target.(knifer.ErrCode)
	return ok && e.Code == code
}
