package vjwt

import jwtimpl "github.com/imajinyun/go-knifer/internal/jwt"

// JWT is a JSON Web Token object.
type JWT = jwtimpl.JWT

// JWTSigner signs and verifies JWT tokens.
type JWTSigner = jwtimpl.JWTSigner

// JWTValidator validates JWT claims.
type JWTValidator = jwtimpl.JWTValidator

// TokenOption customizes CreateTokenWithOptions.
type TokenOption = jwtimpl.TokenOption

// ValidateOption customizes JWT ValidateWithOptions.
type ValidateOption = jwtimpl.ValidateOption

// SignerOption customizes asymmetric JWT signers.
type SignerOption = jwtimpl.SignerOption

// JWTError is the JWT module error type.
type JWTError = jwtimpl.JWTError

const (
	// JWTAlgNone is the none algorithm identifier.
	JWTAlgNone = jwtimpl.AlgNone
	// JWTAlgHS256 is the HS256 algorithm identifier.
	JWTAlgHS256 = jwtimpl.AlgHS256
	// JWTAlgHS384 is the HS384 algorithm identifier.
	JWTAlgHS384 = jwtimpl.AlgHS384
	// JWTAlgHS512 is the HS512 algorithm identifier.
	JWTAlgHS512 = jwtimpl.AlgHS512
	// JWTAlgRS256 is the RS256 algorithm identifier.
	JWTAlgRS256 = jwtimpl.AlgRS256
	// JWTAlgRS384 is the RS384 algorithm identifier.
	JWTAlgRS384 = jwtimpl.AlgRS384
	// JWTAlgRS512 is the RS512 algorithm identifier.
	JWTAlgRS512 = jwtimpl.AlgRS512
	// JWTAlgPS256 is the PS256 algorithm identifier.
	JWTAlgPS256 = jwtimpl.AlgPS256
	// JWTAlgPS384 is the PS384 algorithm identifier.
	JWTAlgPS384 = jwtimpl.AlgPS384
	// JWTAlgPS512 is the PS512 algorithm identifier.
	JWTAlgPS512 = jwtimpl.AlgPS512
	// JWTAlgES256 is the ES256 algorithm identifier.
	JWTAlgES256 = jwtimpl.AlgES256
	// JWTAlgES384 is the ES384 algorithm identifier.
	JWTAlgES384 = jwtimpl.AlgES384
	// JWTAlgES512 is the ES512 algorithm identifier.
	JWTAlgES512 = jwtimpl.AlgES512
	// JWTHeaderAlgorithm is the alg header key.
	JWTHeaderAlgorithm = jwtimpl.HeaderAlgorithm
	// JWTHeaderType is the typ header key.
	JWTHeaderType = jwtimpl.HeaderType
	// JWTHeaderContentType is the cty header key.
	JWTHeaderContentType = jwtimpl.HeaderContentType
	// JWTHeaderKeyID is the kid header key.
	JWTHeaderKeyID = jwtimpl.HeaderKeyID
	// JWTPayloadIssuer is the iss payload key.
	JWTPayloadIssuer = jwtimpl.PayloadIssuer
	// JWTPayloadSubject is the sub payload key.
	JWTPayloadSubject = jwtimpl.PayloadSubject
	// JWTPayloadAudience is the aud payload key.
	JWTPayloadAudience = jwtimpl.PayloadAudience
	// JWTPayloadExpiresAt is the exp payload key.
	JWTPayloadExpiresAt = jwtimpl.PayloadExpiresAt
	// JWTPayloadNotBefore is the nbf payload key.
	JWTPayloadNotBefore = jwtimpl.PayloadNotBefore
	// JWTPayloadIssuedAt is the iat payload key.
	JWTPayloadIssuedAt = jwtimpl.PayloadIssuedAt
	// JWTPayloadJWTID is the jti payload key.
	JWTPayloadJWTID = jwtimpl.PayloadJWTID
)
