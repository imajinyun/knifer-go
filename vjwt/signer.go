package vjwt

import (
	"crypto/ecdsa"
	"crypto/rsa"

	jwtimpl "github.com/imajinyun/go-knifer/internal/jwt"
)

// JWTSignerHMAC creates an HMAC signer.
func JWTSignerHMAC(algorithm string, key []byte) (JWTSigner, error) {
	return jwtimpl.NewHMACSigner(algorithm, key)
}

// JWTSignerHS256 creates an HS256 signer.
func JWTSignerHS256(key []byte) JWTSigner { return jwtimpl.HS256(key) }

// JWTSignerRSA creates an RSA signer.
func JWTSignerRSA(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewRSASigner(algorithm, priv, pub)
}

// JWTSignerRS256 creates an RS256 signer.
func JWTSignerRS256(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.RS256(priv, pub)
}

// JWTSignerECDSA creates an ECDSA signer.
func JWTSignerECDSA(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewECDSASigner(algorithm, priv, pub)
}

// JWTSignerES256 creates an ES256 signer.
func JWTSignerES256(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES256(priv, pub)
}

// JWTSignerNone creates a none signer.
func JWTSignerNone() JWTSigner { return jwtimpl.None() }

// NoneSigner delegates to the internal jwt implementation.
func NoneSigner() JWTSigner {
	return jwtimpl.NoneSigner()
}

// IsNoneAlg delegates to the internal jwt implementation.
func IsNoneAlg(alg string) bool {
	return jwtimpl.IsNoneAlg(alg)
}

// NewHMACSigner delegates to the internal jwt implementation.
func NewHMACSigner(algorithm string, key []byte) (JWTSigner, error) {
	return jwtimpl.NewHMACSigner(algorithm, key)
}

// MustHMACSigner delegates to the internal jwt implementation.
func MustHMACSigner(algorithm string, key []byte) JWTSigner {
	return jwtimpl.MustHMACSigner(algorithm, key)
}

// CreateSigner delegates to the internal jwt implementation.
func CreateSigner(algorithmID string, key []byte) (JWTSigner, error) {
	return jwtimpl.CreateSigner(algorithmID, key)
}

// AlgorithmName delegates to the internal jwt implementation.
func AlgorithmName(idOrAlgorithm string) string {
	return jwtimpl.AlgorithmName(idOrAlgorithm)
}

// NewRSASigner delegates to the internal jwt implementation.
func NewRSASigner(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewRSASigner(algorithm, priv, pub)
}

// NewRSAPSSSigner delegates to the internal jwt implementation.
func NewRSAPSSSigner(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewRSAPSSSigner(algorithm, priv, pub)
}

// NewECDSASigner delegates to the internal jwt implementation.
func NewECDSASigner(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewECDSASigner(algorithm, priv, pub)
}

// HS256 delegates to the internal jwt implementation.
func HS256(key []byte) JWTSigner {
	return jwtimpl.HS256(key)
}

// HS384 delegates to the internal jwt implementation.
func HS384(key []byte) JWTSigner {
	return jwtimpl.HS384(key)
}

// HS512 delegates to the internal jwt implementation.
func HS512(key []byte) JWTSigner {
	return jwtimpl.HS512(key)
}

// RS256 delegates to the internal jwt implementation.
func RS256(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.RS256(priv, pub)
}

// RS384 delegates to the internal jwt implementation.
func RS384(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.RS384(priv, pub)
}

// RS512 delegates to the internal jwt implementation.
func RS512(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.RS512(priv, pub)
}

// PS256 delegates to the internal jwt implementation.
func PS256(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.PS256(priv, pub)
}

// PS384 delegates to the internal jwt implementation.
func PS384(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.PS384(priv, pub)
}

// PS512 delegates to the internal jwt implementation.
func PS512(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.PS512(priv, pub)
}

// ES256 delegates to the internal jwt implementation.
func ES256(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES256(priv, pub)
}

// ES384 delegates to the internal jwt implementation.
func ES384(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES384(priv, pub)
}

// ES512 delegates to the internal jwt implementation.
func ES512(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES512(priv, pub)
}

// None delegates to the internal jwt implementation.
func None() JWTSigner {
	return jwtimpl.None()
}
