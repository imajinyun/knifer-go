package json

import (
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
)

// JSONError 对应 the utility JSONException。
type JSONError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// NewJSONError 使用消息构造错误。
func NewJSONError(format string, args ...any) *JSONError {
	return &JSONError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...)}
}

// WrapJSONError 包装一个底层错误。
func WrapJSONError(cause error, format string, args ...any) *JSONError {
	return &JSONError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...), Cause: cause}
}

func (e *JSONError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
	}
	return e.Msg
}

// ErrorCode returns the go-knifer error code.
func (e *JSONError) ErrorCode() knifer.ErrCode { return e.Code }

// Unwrap 支持 errors.Is/As。
func (e *JSONError) Unwrap() error { return e.Cause }

// Is 支持 errors.Is(err, knifer.ErrCodeXxx) 按错误码匹配。
func (e *JSONError) Is(target error) bool {
	code, ok := target.(knifer.ErrCode)
	return ok && e.Code == code
}
