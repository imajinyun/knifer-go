package vjwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"io"

	jwtimpl "github.com/imajinyun/knifer-go/internal/jwt"
)

// JWTSignerHMAC creates an HMAC signer.
func JWTSignerHMAC(algorithm string, key []byte) (JWTSigner, error) {
	return jwtimpl.NewHMACSigner(algorithm, key)
}

// JWTSignerHS256 creates an HS256 signer.
func JWTSignerHS256(key []byte) JWTSigner { return jwtimpl.HS256(key) }

// JWTSignerECDSA creates an ECDSA signer.
func JWTSignerECDSA(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewECDSASigner(algorithm, priv, pub)
}

// JWTSignerES256 creates an ES256 signer.
func JWTSignerES256(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES256(priv, pub)
}

// NewHMACSigner creates an HMAC signer for HS256, HS384, or HS512.
func NewHMACSigner(algorithm string, key []byte) (JWTSigner, error) {
	return jwtimpl.NewHMACSigner(algorithm, key)
}

// NewHMACSignerStrict creates an HMAC signer and enforces the recommended minimum key length.
func NewHMACSignerStrict(algorithm string, key []byte) (JWTSigner, error) {
	return jwtimpl.NewHMACSignerStrict(algorithm, key)
}

// MustHMACSigner creates an HMAC signer and panics on invalid algorithms.
func MustHMACSigner(algorithm string, key []byte) JWTSigner {
	return jwtimpl.MustHMACSigner(algorithm, key)
}

// CreateSigner creates an HMAC signer from algorithm ID. The none algorithm is always rejected.
func CreateSigner(algorithmID string, key []byte) (JWTSigner, error) {
	return jwtimpl.CreateSigner(algorithmID, key)
}

// CreateSignerStrict creates an HMAC signer and enforces the recommended minimum key length.
func CreateSignerStrict(algorithmID string, key []byte) (JWTSigner, error) {
	return jwtimpl.CreateSignerStrict(algorithmID, key)
}

// MinHMACKeyBytes returns the minimum recommended HMAC key length for an HS* JWT algorithm.
func MinHMACKeyBytes(algorithm string) (int, error) {
	return jwtimpl.MinHMACKeyBytes(algorithm)
}

// AlgorithmName returns the standard cryptographic algorithm name for a JWT algorithm ID.
func AlgorithmName(idOrAlgorithm string) string {
	return jwtimpl.AlgorithmName(idOrAlgorithm)
}

// NewRSAPSSSigner creates an RSA-PSS signer for PS256, PS384, or PS512.
func NewRSAPSSSigner(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewRSAPSSSigner(algorithm, priv, pub)
}

// NewRSAPSSSignerWithOptions creates an RSA-PSS signer for PS256, PS384, or PS512 with options.
func NewRSAPSSSignerWithOptions(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) (JWTSigner, error) {
	return jwtimpl.NewRSAPSSSignerWithOptions(algorithm, priv, pub, opts...)
}

// NewECDSASigner creates an ECDSA signer for ES256, ES384, or ES512.
func NewECDSASigner(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) (JWTSigner, error) {
	return jwtimpl.NewECDSASigner(algorithm, priv, pub)
}

// NewECDSASignerWithOptions creates an ECDSA signer for ES256, ES384, or ES512 with options.
func NewECDSASignerWithOptions(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) (JWTSigner, error) {
	return jwtimpl.NewECDSASignerWithOptions(algorithm, priv, pub, opts...)
}

// WithSignerRandomReader sets the random source used by RSA-PSS and ECDSA signing.
func WithSignerRandomReader(reader io.Reader) SignerOption {
	return jwtimpl.WithSignerRandomReader(reader)
}

// WithRSAPSSOptions sets RSA-PSS options used by RSA-PSS signing and verification.
func WithRSAPSSOptions(opts *rsa.PSSOptions) SignerOption {
	return jwtimpl.WithRSAPSSOptions(opts)
}

// HS256 creates an HS256 signer.
func HS256(key []byte) JWTSigner {
	return jwtimpl.HS256(key)
}

// HS384 creates an HS384 signer.
func HS384(key []byte) JWTSigner {
	return jwtimpl.HS384(key)
}

// HS512 creates an HS512 signer.
func HS512(key []byte) JWTSigner {
	return jwtimpl.HS512(key)
}

// PS256 creates a PS256 signer.
func PS256(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.PS256(priv, pub)
}

// PS256WithOptions creates a PS256 signer with options.
func PS256WithOptions(priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) JWTSigner {
	return jwtimpl.PS256WithOptions(priv, pub, opts...)
}

// PS384 creates a PS384 signer.
func PS384(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.PS384(priv, pub)
}

// PS384WithOptions creates a PS384 signer with options.
func PS384WithOptions(priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) JWTSigner {
	return jwtimpl.PS384WithOptions(priv, pub, opts...)
}

// PS512 creates a PS512 signer.
func PS512(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return jwtimpl.PS512(priv, pub)
}

// PS512WithOptions creates a PS512 signer with options.
func PS512WithOptions(priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) JWTSigner {
	return jwtimpl.PS512WithOptions(priv, pub, opts...)
}

// ES256 creates an ES256 signer.
func ES256(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES256(priv, pub)
}

// ES256WithOptions creates an ES256 signer with options.
func ES256WithOptions(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) JWTSigner {
	return jwtimpl.ES256WithOptions(priv, pub, opts...)
}

// ES384 creates an ES384 signer.
func ES384(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES384(priv, pub)
}

// ES384WithOptions creates an ES384 signer with options.
func ES384WithOptions(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) JWTSigner {
	return jwtimpl.ES384WithOptions(priv, pub, opts...)
}

// ES512 creates an ES512 signer.
func ES512(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return jwtimpl.ES512(priv, pub)
}

// ES512WithOptions creates an ES512 signer with options.
func ES512WithOptions(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) JWTSigner {
	return jwtimpl.ES512WithOptions(priv, pub, opts...)
}
