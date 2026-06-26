package bloomfilter

import (
	"fmt"
	"os"

	knifer "github.com/imajinyun/knifer-go"
)

// BloomFilterError represents an error produced by bloom filter helpers.
type BloomFilterError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the bloom filter error message.
func (e *BloomFilterError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *BloomFilterError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *BloomFilterError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *BloomFilterError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*BloomFilterError); ok {
		return e.Code == other.Code
	}
	return false
}

func wrapBloomFilterError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &BloomFilterError{Code: code, Msg: msg, Cause: cause}
}

func wrapBloomFilterIO(msg string, cause error) error {
	code := knifer.ErrCodeInternal
	if os.IsNotExist(cause) {
		code = knifer.ErrCodeNotFound
	}
	return wrapBloomFilterError(code, msg, cause)
}

func invalidInputf(format string, args ...any) error {
	return &BloomFilterError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...)}
}
