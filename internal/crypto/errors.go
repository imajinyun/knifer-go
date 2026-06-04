package crypto

import knifer "github.com/imajinyun/go-knifer"

type sentinel struct {
	code knifer.ErrCode
	msg  string
}

func (e *sentinel) Error() string { return e.msg }

func (e *sentinel) ErrorCode() knifer.ErrCode { return e.code }

func (e *sentinel) Is(target error) bool {
	if e == target {
		return true
	}
	code, ok := target.(knifer.ErrCode)
	return ok && e.code == code
}

var (
	// ErrInvalidKey indicates an invalid cryptographic key.
	ErrInvalidKey error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid key"}
	// ErrInvalidIV indicates an invalid initialization vector.
	ErrInvalidIV error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid iv"}
	// ErrInvalidCipherText indicates invalid encrypted data.
	ErrInvalidCipherText error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid cipher text"}
)
