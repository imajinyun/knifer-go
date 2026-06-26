package db

import (
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// DBError represents an error produced by database helpers.
type DBError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the database error message.
func (e *DBError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *DBError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *DBError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *DBError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*DBError); ok {
		return e.Code == other.Code
	}
	return false
}

func dbErrorf(code knifer.ErrCode, format string, args ...any) *DBError {
	return &DBError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func wrapDBError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &DBError{Code: code, Msg: msg, Cause: cause}
}

func invalidInputf(format string, args ...any) *DBError {
	return dbErrorf(knifer.ErrCodeInvalidInput, format, args...)
}

func unsupportedf(format string, args ...any) *DBError {
	return dbErrorf(knifer.ErrCodeUnsupported, format, args...)
}

func wrapInternal(msg string, cause error) error {
	return wrapDBError(knifer.ErrCodeInternal, msg, cause)
}
