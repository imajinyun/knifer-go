package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"hash"
	"strings"
)

// JWTSigner JWT 签名器接口（对应 the utility toolkit-jwt JWTSigner）。
type JWTSigner interface {
	// Algorithm 返回算法 ID（如 HS256）。
	Algorithm() string
	// Sign 对 headerBase64.payloadBase64 计算签名，返回 base64url（无填充）字符串。
	Sign(headerB64, payloadB64 string) string
	// Verify 校验签名是否匹配。
	Verify(headerB64, payloadB64, signB64 string) bool
}

// 算法 ID 常量。
const (
	AlgNone  = "none"
	AlgHS256 = "HS256"
	AlgHS384 = "HS384"
	AlgHS512 = "HS512"
)

// b64URLEncode 标准 JWT 中使用的 base64url（不带 padding）编码。
func b64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// b64URLDecode base64url 解码，兼容带 padding 的输入。
func b64URLDecode(s string) ([]byte, error) {
	// 优先使用无 padding；若失败则尝试带 padding。
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.URLEncoding.DecodeString(s)
}

// noneSigner 对应 NoneJWTSigner：算法为 none，不签名也不校验内容。
type noneSigner struct{}

// NoneSigner 返回 None 签名器单例。
func NoneSigner() JWTSigner { return noneSigner{} }

func (noneSigner) Algorithm() string                                 { return AlgNone }
func (noneSigner) Sign(headerB64, payloadB64 string) string          { return "" }
func (noneSigner) Verify(headerB64, payloadB64, signB64 string) bool { return signB64 == "" }

// IsNoneAlg 判断算法 ID 是否为 none。
func IsNoneAlg(alg string) bool { return strings.EqualFold(strings.TrimSpace(alg), AlgNone) }

// hmacSigner HMAC 系列签名器。
type hmacSigner struct {
	alg    string
	key    []byte
	hashFn func() hash.Hash
}

// NewHMACSigner 创建 HMAC 签名器。algorithm 仅支持 HS256/HS384/HS512。
func NewHMACSigner(algorithm string, key []byte) (JWTSigner, error) {
	algorithm = strings.ToUpper(strings.TrimSpace(algorithm))
	switch algorithm {
	case AlgHS256:
		return &hmacSigner{alg: AlgHS256, key: append([]byte{}, key...), hashFn: sha256.New}, nil
	case AlgHS384:
		return &hmacSigner{alg: AlgHS384, key: append([]byte{}, key...), hashFn: sha512.New384}, nil
	case AlgHS512:
		return &hmacSigner{alg: AlgHS512, key: append([]byte{}, key...), hashFn: sha512.New}, nil
	}
	// 兼容传 SHA384 直接 hash 别名
	return nil, unsupportedJWTErrorf("unsupported HMAC algorithm: %s", algorithm)
}

// MustHMACSigner 创建 HMAC 签名器，失败 panic。
func MustHMACSigner(algorithm string, key []byte) JWTSigner {
	s, err := NewHMACSigner(algorithm, key)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *hmacSigner) Algorithm() string { return s.alg }

func (s *hmacSigner) Sign(headerB64, payloadB64 string) string {
	mac := hmac.New(s.hashFn, s.key)
	mac.Write([]byte(headerB64))
	mac.Write([]byte{'.'})
	mac.Write([]byte(payloadB64))
	return b64URLEncode(mac.Sum(nil))
}

func (s *hmacSigner) Verify(headerB64, payloadB64, signB64 string) bool {
	expected := s.Sign(headerB64, payloadB64)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signB64)) == 1
}

// CreateSigner 根据算法 ID 与 HMAC key 自动选择签名器（仅支持 HS* 与 none）。
//
// 非对称算法请使用 NewRSAPSSSigner / NewECDSASigner，
// 或 JWTSignerUtil 提供的 PS256/ES256 等便捷工厂。
func CreateSigner(algorithmID string, key []byte) (JWTSigner, error) {
	if IsNoneAlg(algorithmID) {
		return NoneSigner(), nil
	}
	return NewHMACSigner(algorithmID, key)
}

// AlgorithmName 返回 JWT 算法 ID 对应的标准算法名（the utility toolkit AlgorithmUtil.getAlgorithm）。
// 若传入未知 ID，则原样返回。
func AlgorithmName(idOrAlgorithm string) string {
	id := strings.ToUpper(strings.TrimSpace(idOrAlgorithm))
	switch id {
	case AlgHS256:
		return "HmacSHA256"
	case AlgHS384:
		return "HmacSHA384"
	case AlgHS512:
		return "HmacSHA512"
	case AlgPS256:
		return "SHA256withRSA_PSS"
	case AlgPS384:
		return "SHA384withRSA_PSS"
	case AlgPS512:
		return "SHA512withRSA_PSS"
	case AlgES256:
		return "SHA256withECDSA"
	case AlgES384:
		return "SHA384withECDSA"
	case AlgES512:
		return "SHA512withECDSA"
	case "NONE", "":
		return "None"
	}
	return idOrAlgorithm
}
