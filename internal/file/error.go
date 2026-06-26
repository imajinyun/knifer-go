package file

import (
	"fmt"
	"os"

	knifer "github.com/imajinyun/knifer-go"
)

// FileError represents an error produced by file helpers.
type FileError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the file error message.
func (e *FileError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *FileError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *FileError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *FileError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*FileError); ok {
		return e.Code == other.Code
	}
	return false
}

func fileErrorf(code knifer.ErrCode, format string, args ...any) *FileError {
	return &FileError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func wrapFileError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &FileError{Code: code, Msg: msg, Cause: cause}
}

func invalidInputf(format string, args ...any) *FileError {
	return fileErrorf(knifer.ErrCodeInvalidInput, format, args...)
}

func wrapFileIO(msg string, cause error) error {
	code := knifer.ErrCodeInternal
	if os.IsNotExist(cause) {
		code = knifer.ErrCodeNotFound
	}
	return wrapFileError(code, msg, cause)
}
