package jwt

import (
	"fmt"

	knifer "github.com/imajinyun/go-knifer"
)

// JWTError JWT 相关错误。
type JWTError struct {
	Code knifer.ErrCode
	Msg  string
	Err  error
}

// Error 实现 error 接口。
func (e *JWTError) Error() string {
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

// ErrorCode returns the go-knifer error code.
func (e *JWTError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap 返回内部错误。
func (e *JWTError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is 支持 errors.Is(err, knifer.ErrCodeXxx) 按错误码匹配。
func (e *JWTError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	code, ok := target.(knifer.ErrCode)
	return ok && e.Code == code
}

// NewJWTError 构造错误。
func NewJWTError(msg string) *JWTError {
	return &JWTError{Code: knifer.ErrCodeInvalidInput, Msg: msg}
}

// JWTErrorf 格式化构造错误。
func JWTErrorf(format string, args ...any) *JWTError {
	return &JWTError{Code: knifer.ErrCodeInvalidInput, Msg: fmt.Sprintf(format, args...)}
}

func wrapJWTError(cause error, msg string) *JWTError {
	return &JWTError{Code: knifer.ErrCodeInvalidInput, Msg: msg, Err: cause}
}

func unsupportedJWTErrorf(format string, args ...any) *JWTError {
	return &JWTError{Code: knifer.ErrCodeUnsupported, Msg: fmt.Sprintf(format, args...)}
}
