package jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
)

// 对应 the utility toolkit-jwt JWTSignerUtil 的便捷工厂函数。
// 在 Go 风格中以包级函数提供。

// HS256 创建 HS256 签名器。
func HS256(key []byte) JWTSigner { return MustHMACSigner(AlgHS256, key) }

// HS384 创建 HS384 签名器。
func HS384(key []byte) JWTSigner { return MustHMACSigner(AlgHS384, key) }

// HS512 创建 HS512 签名器。
func HS512(key []byte) JWTSigner { return MustHMACSigner(AlgHS512, key) }

// PS256 创建 RSA-PSS 签名器。
func PS256(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return PS256WithOptions(priv, pub)
}

// PS256WithOptions 创建可配置 RSA-PSS 签名器。
func PS256WithOptions(priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) JWTSigner {
	return mustRSAPSSWithOptions(AlgPS256, priv, pub, opts...)
}

// PS384 同上。
func PS384(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return PS384WithOptions(priv, pub)
}

// PS384WithOptions 创建可配置 RSA-PSS 签名器。
func PS384WithOptions(priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) JWTSigner {
	return mustRSAPSSWithOptions(AlgPS384, priv, pub, opts...)
}

// PS512 同上。
func PS512(priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTSigner {
	return PS512WithOptions(priv, pub)
}

// PS512WithOptions 创建可配置 RSA-PSS 签名器。
func PS512WithOptions(priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) JWTSigner {
	return mustRSAPSSWithOptions(AlgPS512, priv, pub, opts...)
}

// ES256 创建 ECDSA(P-256) 签名器。
func ES256(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return ES256WithOptions(priv, pub)
}

// ES256WithOptions 创建可配置 ECDSA(P-256) 签名器。
func ES256WithOptions(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) JWTSigner {
	return mustECDSAWithOptions(AlgES256, priv, pub, opts...)
}

// ES384 创建 ECDSA(P-384) 签名器。
func ES384(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return ES384WithOptions(priv, pub)
}

// ES384WithOptions 创建可配置 ECDSA(P-384) 签名器。
func ES384WithOptions(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) JWTSigner {
	return mustECDSAWithOptions(AlgES384, priv, pub, opts...)
}

// ES512 创建 ECDSA(P-521) 签名器。
func ES512(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) JWTSigner {
	return ES512WithOptions(priv, pub)
}

// ES512WithOptions 创建可配置 ECDSA(P-521) 签名器。
func ES512WithOptions(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) JWTSigner {
	return mustECDSAWithOptions(AlgES512, priv, pub, opts...)
}

func mustRSAPSSWithOptions(alg string, priv *rsa.PrivateKey, pub *rsa.PublicKey, opts ...SignerOption) JWTSigner {
	s, err := NewRSAPSSSignerWithOptions(alg, priv, pub, opts...)
	if err != nil {
		panic(err)
	}
	return s
}

func mustECDSAWithOptions(alg string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, opts ...SignerOption) JWTSigner {
	s, err := NewECDSASignerWithOptions(alg, priv, pub, opts...)
	if err != nil {
		panic(err)
	}
	return s
}
