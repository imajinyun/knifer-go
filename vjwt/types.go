package vjwt

import jwtimpl "github.com/imajinyun/go-knifer/internal/jwt"

// JWT is a JSON Web Token object.
type JWT = jwtimpl.JWT

// JWTSigner signs and verifies JWT tokens.
type JWTSigner = jwtimpl.JWTSigner

// JWTValidator validates JWT claims.
type JWTValidator = jwtimpl.JWTValidator

// JWTError is the JWT module error type.
type JWTError = jwtimpl.JWTError

const (
	// JWTAlgNone is the none algorithm identifier.
	JWTAlgNone = jwtimpl.AlgNone
	// JWTAlgHS256 is the HS256 algorithm identifier.
	JWTAlgHS256 = jwtimpl.AlgHS256
	// JWTAlgRS256 is the RS256 algorithm identifier.
	JWTAlgRS256 = jwtimpl.AlgRS256
	// JWTAlgES256 is the ES256 algorithm identifier.
	JWTAlgES256 = jwtimpl.AlgES256
	// JWTHeaderAlgorithm is the alg header key.
	JWTHeaderAlgorithm = jwtimpl.HeaderAlgorithm
	// JWTPayloadIssuer is the iss payload key.
	JWTPayloadIssuer = jwtimpl.PayloadIssuer
	// JWTPayloadSubject is the sub payload key.
	JWTPayloadSubject = jwtimpl.PayloadSubject
	// JWTPayloadExpiresAt is the exp payload key.
	JWTPayloadExpiresAt = jwtimpl.PayloadExpiresAt
)
