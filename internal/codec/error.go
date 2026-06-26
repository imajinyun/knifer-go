package codec

import knifer "github.com/imajinyun/knifer-go"

// CodecError represents an error produced by codec helpers.
type CodecError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the codec error message.
func (e *CodecError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *CodecError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *CodecError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *CodecError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*CodecError); ok {
		return e.Code == other.Code
	}
	return false
}

func wrapCodecError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &CodecError{Code: code, Msg: msg, Cause: cause}
}

func invalidCodecInput(msg string, cause error) error {
	return wrapCodecError(knifer.ErrCodeInvalidInput, msg, cause)
}
