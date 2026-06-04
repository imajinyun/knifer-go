package vjwt

import jwtimpl "github.com/imajinyun/go-knifer/internal/jwt"

// NewJWTError delegates to the internal jwt implementation.
func NewJWTError(msg string) *JWTError {
	return jwtimpl.NewJWTError(msg)
}

// JWTErrorf delegates to the internal jwt implementation.
func JWTErrorf(format string, args ...any) *JWTError {
	return jwtimpl.JWTErrorf(format, args...)
}
