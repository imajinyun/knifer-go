package mail

import (
	"errors"
	"fmt"

	knifer "github.com/imajinyun/knifer-go"
)

// sentinel is a package-level error value that carries a knifer-go error code
// while preserving sentinel identity for errors.Is comparisons.
type sentinel struct {
	code knifer.ErrCode
	msg  string
}

func (e *sentinel) Error() string { return e.msg }

// ErrorCode implements the knifer.CodeCarrier interface so knifer.CodeOf can
// classify mail errors.
func (e *sentinel) ErrorCode() knifer.ErrCode { return e.code }

// Is matches the same sentinel pointer or a bare knifer.ErrCode target.
func (e *sentinel) Is(target error) bool {
	if e == target {
		return true
	}
	code, ok := target.(knifer.ErrCode)
	return ok && e.code == code
}

var (
	// ErrInvalidAddress is returned when an email address cannot be parsed or validated.
	ErrInvalidAddress error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "mail: invalid address"}
	// ErrInvalidHeader is returned when a header name or value is not safe for SMTP/MIME output.
	ErrInvalidHeader error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "mail: invalid header"}
	// ErrMissingFrom is returned when a message has no From address.
	ErrMissingFrom error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "mail: missing from address"}
	// ErrMissingRecipient is returned when a message has no To, Cc, or Bcc recipient.
	ErrMissingRecipient error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "mail: missing recipient"}
	// ErrMissingBody is returned when a message has no body content.
	ErrMissingBody error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "mail: missing body"}
	// ErrTLSRequired is returned when the configured security policy requires TLS but TLS is unavailable.
	ErrTLSRequired error = &sentinel{code: knifer.ErrCodeUnsupported, msg: "mail: tls required"}
	// ErrPlainAuth is returned when SMTP AUTH would be sent over a plaintext connection.
	ErrPlainAuth error = &sentinel{code: knifer.ErrCodeUnsupported, msg: "mail: plaintext auth disabled"}
	// ErrAttachmentTooLarge is returned when an attachment exceeds the configured size limit.
	ErrAttachmentTooLarge error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "mail: attachment too large"}
)

// MailError carries a knifer-go error code while preserving an underlying cause.
type MailError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

func (e *MailError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.Cause != nil && e.Msg != "":
		return fmt.Sprintf("%s: %s", e.Msg, e.Cause.Error())
	case e.Cause != nil:
		return e.Cause.Error()
	default:
		return e.Msg
	}
}

func (e *MailError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

func (e *MailError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *MailError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*MailError); ok {
		return e.Code == other.Code
	}
	return false
}

func wrapMailError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return &MailError{Code: code, Msg: msg}
	}
	return &MailError{Code: code, Msg: msg, Cause: cause}
}

func wrapProviderError(msg string, cause error) error {
	if cause == nil {
		return wrapMailError(knifer.ErrCodeProviderFailure, msg, nil)
	}
	for _, code := range []knifer.ErrCode{
		knifer.ErrCodeInvalidInput,
		knifer.ErrCodeNotFound,
		knifer.ErrCodeUnsupported,
		knifer.ErrCodeUnsafeResource,
		knifer.ErrCodeTimeout,
		knifer.ErrCodeProviderFailure,
		knifer.ErrCodeInternal,
	} {
		if errors.Is(cause, code) {
			return cause
		}
	}
	return wrapMailError(knifer.ErrCodeProviderFailure, msg, cause)
}
