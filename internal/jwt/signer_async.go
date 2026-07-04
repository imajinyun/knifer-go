package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/asn1"
	"hash"
	"io"
	"math/big"
	"strings"
)

// Asymmetric signing algorithm IDs.
const (
	AlgPS256 = "PS256"
	AlgPS384 = "PS384"
	AlgPS512 = "PS512"
	AlgES256 = "ES256"
	AlgES384 = "ES384"
	AlgES512 = "ES512"
)

// rsaSigner matches the utility toolkit AsymmetricJWTSigner for RSA-PSS only.
type rsaSigner struct {
	alg        string
	pub        *rsa.PublicKey
	priv       *rsa.PrivateKey
	hashID     crypto.Hash
	random     io.Reader
	pssOptions *rsa.PSSOptions
}

type signerConfig struct {
	random     io.Reader
	pssOptions *rsa.PSSOptions
}

// SignerOption customizes asymmetric JWT signers.
type SignerOption func(*signerConfig)

// WithSignerRandomReader sets the random source used by RSA-PSS and ECDSA signing.
func WithSignerRandomReader(reader io.Reader) SignerOption {
	return func(c *signerConfig) {
		if reader != nil {
			c.random = reader
		}
	}
}

// WithRSAPSSOptions sets RSA-PSS options used by RSA-PSS signing and verification.
func WithRSAPSSOptions(opts *rsa.PSSOptions) SignerOption {
	return func(c *signerConfig) {
		if opts == nil {
			c.pssOptions = nil
			return
		}
		clone := *opts
		c.pssOptions = &clone
	}
}

func applySignerOptions(opts []SignerOption) signerConfig {
	cfg := signerConfig{random: rand.Reader, pssOptions: defaultPSSOptions()}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.random == nil {
		cfg.random = rand.Reader
	}
	if cfg.pssOptions == nil {
		cfg.pssOptions = defaultPSSOptions()
	}
	return cfg
}

func defaultPSSOptions() *rsa.PSSOptions {
	return &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}
}

// NewRSAPSSSigner creates an RSA-PSS signer.
// algorithm: PS256 / PS384 / PS512。
func NewRSAPSSSigner(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey) (JWTSigner, error) {
	return NewRSAPSSSignerWithOptions(algorithm, priv, pub)
}

// NewRSAPSSSignerWithOptions creates a configurable RSA-PSS signer.
func NewRSAPSSSignerWithOptions(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) (JWTSigner, error) {
	algorithm = strings.ToUpper(strings.TrimSpace(algorithm))
	hashID, ok := rsaHashOf(algorithm)
	if !ok {
		return nil, unsupportedJWTErrorf("unsupported RSA-PSS algorithm: %s", algorithm)
	}
	if priv != nil && pub == nil {
		pub = &priv.PublicKey
	}
	if priv == nil && pub == nil {
		return nil, NewJWTError("RSA-PSS signer requires private key or public key")
	}
	cfg := applySignerOptions(opts)
	return &rsaSigner{alg: algorithm, priv: priv, pub: pub, hashID: hashID, random: cfg.random, pssOptions: cfg.pssOptions}, nil
}

func rsaHashOf(alg string) (crypto.Hash, bool) {
	switch alg {
	case AlgPS256:
		return crypto.SHA256, true
	case AlgPS384:
		return crypto.SHA384, true
	case AlgPS512:
		return crypto.SHA512, true
	}
	return 0, false
}

func (s *rsaSigner) Algorithm() string { return s.alg }

func (s *rsaSigner) Sign(headerB64, payloadB64 string) string {
	if s.priv == nil {
		return ""
	}
	digest := digestOf(s.hashID, headerB64+"."+payloadB64)
	sig, err := rsa.SignPSS(s.random, s.priv, s.hashID, digest, s.pssOptions)
	if err != nil {
		return ""
	}
	return b64URLEncode(sig)
}

func (s *rsaSigner) Verify(headerB64, payloadB64, signB64 string) bool {
	if s.pub == nil {
		return false
	}
	sig, err := b64URLDecode(signB64)
	if err != nil {
		return false
	}
	digest := digestOf(s.hashID, headerB64+"."+payloadB64)
	return rsa.VerifyPSS(s.pub, s.hashID, digest, sig, s.pssOptions) == nil
}

// ecdsaSigner matches the utility toolkit EllipticCurveJWTSigner.
type ecdsaSigner struct {
	alg    string
	priv   *ecdsa.PrivateKey
	pub    *ecdsa.PublicKey
	hashID crypto.Hash
	rSize  int // r/s fixed byte length in JWT serialization
	random io.Reader
}

// NewECDSASigner creates an ECDSA signer.
// algorithm: ES256(P-256) / ES384(P-384) / ES512(P-521)。
func NewECDSASigner(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) (JWTSigner, error) {
	return NewECDSASignerWithOptions(algorithm, priv, pub)
}

// NewECDSASignerWithOptions creates a configurable ECDSA signer.
func NewECDSASignerWithOptions(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) (JWTSigner, error) {
	algorithm = strings.ToUpper(strings.TrimSpace(algorithm))
	hashID, expectedCurve, rSize, ok := ecdsaParamsOf(algorithm)
	if !ok {
		return nil, unsupportedJWTErrorf("unsupported ECDSA algorithm: %s", algorithm)
	}
	if priv != nil && pub == nil {
		pub = &priv.PublicKey
	}
	if priv == nil && pub == nil {
		return nil, NewJWTError("ECDSA signer requires private key or public key")
	}
	if pub.Curve != expectedCurve {
		return nil, JWTErrorf("curve mismatch: %s requires %s", algorithm, curveName(expectedCurve))
	}
	cfg := applySignerOptions(opts)
	return &ecdsaSigner{alg: algorithm, priv: priv, pub: pub, hashID: hashID, rSize: rSize, random: cfg.random}, nil
}

func ecdsaParamsOf(alg string) (crypto.Hash, elliptic.Curve, int, bool) {
	switch alg {
	case AlgES256:
		return crypto.SHA256, elliptic.P256(), 32, true
	case AlgES384:
		return crypto.SHA384, elliptic.P384(), 48, true
	case AlgES512:
		return crypto.SHA512, elliptic.P521(), 66, true
	}
	return 0, nil, 0, false
}

func curveName(c elliptic.Curve) string {
	switch c {
	case elliptic.P256():
		return "P-256"
	case elliptic.P384():
		return "P-384"
	case elliptic.P521():
		return "P-521"
	}
	return "unknown"
}

func (s *ecdsaSigner) Algorithm() string { return s.alg }

func (s *ecdsaSigner) Sign(headerB64, payloadB64 string) string {
	if s.priv == nil {
		return ""
	}
	digest := digestOf(s.hashID, headerB64+"."+payloadB64)
	r, sVal, err := ecdsa.Sign(s.random, s.priv, digest)
	if err != nil {
		return ""
	}
	rb := r.Bytes()
	sb := sVal.Bytes()
	out := make([]byte, 2*s.rSize)
	copy(out[s.rSize-len(rb):s.rSize], rb)
	copy(out[2*s.rSize-len(sb):], sb)
	return b64URLEncode(out)
}

func (s *ecdsaSigner) Verify(headerB64, payloadB64, signB64 string) bool {
	if s.pub == nil {
		return false
	}
	raw, err := b64URLDecode(signB64)
	if err != nil {
		return false
	}
	digest := digestOf(s.hashID, headerB64+"."+payloadB64)

	var r, sVal *big.Int
	switch len(raw) {
	case 2 * s.rSize:
		// JOSE fixed-length r||s.
		r = new(big.Int).SetBytes(raw[:s.rSize])
		sVal = new(big.Int).SetBytes(raw[s.rSize:])
	default:
		// Accept ASN.1 DER form for compatibility.
		var sig struct{ R, S *big.Int }
		if _, err := asn1.Unmarshal(raw, &sig); err != nil {
			return false
		}
		r, sVal = sig.R, sig.S
	}
	return ecdsa.Verify(s.pub, digest, r, sVal)
}

// digestOf computes the digest for the specified hash.
func digestOf(h crypto.Hash, data string) []byte {
	var hh hash.Hash
	switch h {
	case crypto.SHA256:
		hh = sha256.New()
	case crypto.SHA384:
		hh = sha512.New384()
	case crypto.SHA512:
		hh = sha512.New()
	default:
		hh = sha256.New()
	}
	hh.Write([]byte(data))
	return hh.Sum(nil)
}
