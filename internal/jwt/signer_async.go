package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/asn1"
	"hash"
	"io"
	"math/big"
	"strings"
)

// 非对称签名相关算法 ID。
const (
	AlgRS256 = "RS256"
	AlgRS384 = "RS384"
	AlgRS512 = "RS512"
	AlgPS256 = "PS256"
	AlgPS384 = "PS384"
	AlgPS512 = "PS512"
	AlgES256 = "ES256"
	AlgES384 = "ES384"
	AlgES512 = "ES512"
)

// rsaSigner 对应 the utility toolkit AsymmetricJWTSigner（仅限 RSA / RSA-PSS）。
type rsaSigner struct {
	alg        string
	pub        *rsa.PublicKey
	priv       *rsa.PrivateKey
	hashID     crypto.Hash
	usePSS     bool
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
	return func(c *signerConfig) { c.random = reader }
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

// NewRSASigner 创建 RSA 签名器（PKCS1v15）。
// algorithm: RS256 / RS384 / RS512。
// privKey、pubKey 至少其一不为 nil；签名需要 priv，验签需要 pub。
func NewRSASigner(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey) (JWTSigner, error) {
	return NewRSASignerWithOptions(algorithm, priv, pub)
}

// NewRSASignerWithOptions 创建可配置 RSA 签名器（PKCS1v15）。
func NewRSASignerWithOptions(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) (JWTSigner, error) {
	algorithm = strings.ToUpper(strings.TrimSpace(algorithm))
	hashID, ok := rsaHashOf(algorithm, false)
	if !ok {
		return nil, unsupportedJWTErrorf("unsupported RSA algorithm: %s", algorithm)
	}
	if priv != nil && pub == nil {
		pub = &priv.PublicKey
	}
	if priv == nil && pub == nil {
		return nil, NewJWTError("RSA signer requires private key or public key")
	}
	cfg := applySignerOptions(opts)
	return &rsaSigner{alg: algorithm, priv: priv, pub: pub, hashID: hashID, usePSS: false, random: cfg.random, pssOptions: cfg.pssOptions}, nil
}

// NewRSAPSSSigner 创建 RSA-PSS 签名器。
// algorithm: PS256 / PS384 / PS512。
func NewRSAPSSSigner(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey) (JWTSigner, error) {
	return NewRSAPSSSignerWithOptions(algorithm, priv, pub)
}

// NewRSAPSSSignerWithOptions 创建可配置 RSA-PSS 签名器。
func NewRSAPSSSignerWithOptions(algorithm string, priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) (JWTSigner, error) {
	algorithm = strings.ToUpper(strings.TrimSpace(algorithm))
	hashID, ok := rsaHashOf(algorithm, true)
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
	return &rsaSigner{alg: algorithm, priv: priv, pub: pub, hashID: hashID, usePSS: true, random: cfg.random, pssOptions: cfg.pssOptions}, nil
}

func rsaHashOf(alg string, pss bool) (crypto.Hash, bool) {
	if pss {
		switch alg {
		case AlgPS256:
			return crypto.SHA256, true
		case AlgPS384:
			return crypto.SHA384, true
		case AlgPS512:
			return crypto.SHA512, true
		}
	} else {
		switch alg {
		case AlgRS256:
			return crypto.SHA256, true
		case AlgRS384:
			return crypto.SHA384, true
		case AlgRS512:
			return crypto.SHA512, true
		}
	}
	return 0, false
}

func (s *rsaSigner) Algorithm() string { return s.alg }

func (s *rsaSigner) Sign(headerB64, payloadB64 string) string {
	if s.priv == nil {
		return ""
	}
	digest := digestOf(s.hashID, headerB64+"."+payloadB64)
	var sig []byte
	var err error
	if s.usePSS {
		sig, err = rsa.SignPSS(s.random, s.priv, s.hashID, digest, s.pssOptions)
	} else {
		sig, err = rsa.SignPKCS1v15(s.random, s.priv, s.hashID, digest)
	}
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
	if s.usePSS {
		return rsa.VerifyPSS(s.pub, s.hashID, digest, sig, s.pssOptions) == nil
	}
	return rsa.VerifyPKCS1v15(s.pub, s.hashID, digest, sig) == nil
}

// ecdsaSigner 对应 the utility toolkit EllipticCurveJWTSigner。
type ecdsaSigner struct {
	alg    string
	priv   *ecdsa.PrivateKey
	pub    *ecdsa.PublicKey
	hashID crypto.Hash
	rSize  int // r/s 在 JWT 序列化中固定字节数
	random io.Reader
}

// NewECDSASigner 创建 ECDSA 签名器。
// algorithm: ES256(P-256) / ES384(P-384) / ES512(P-521)。
func NewECDSASigner(algorithm string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) (JWTSigner, error) {
	return NewECDSASignerWithOptions(algorithm, priv, pub)
}

// NewECDSASignerWithOptions 创建可配置 ECDSA 签名器。
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
		// JOSE 固定长度 r||s
		r = new(big.Int).SetBytes(raw[:s.rSize])
		sVal = new(big.Int).SetBytes(raw[s.rSize:])
	default:
		// 兼容 ASN.1 DER 形式
		var sig struct{ R, S *big.Int }
		if _, err := asn1.Unmarshal(raw, &sig); err != nil {
			return false
		}
		r, sVal = sig.R, sig.S
	}
	return ecdsa.Verify(s.pub, digest, r, sVal)
}

// digestOf 计算指定 hash 的摘要。
func digestOf(h crypto.Hash, data string) []byte {
	var hh hash.Hash
	switch h {
	case crypto.SHA1:
		hh = sha1.New()
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
