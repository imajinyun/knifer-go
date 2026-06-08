// Package vjwt provides public APIs for JWT utilities.
//
// HMAC, RSA-PSS, and ECDSA signers are the production-oriented signing paths.
// The none signer is retained only for explicit opt-in compatibility and tests;
// verification rejects alg=none unless callers deliberately provide the none
// signer. Do not use unsigned JWTs in production.
//
// This package only acts as a facade. Concrete implementations live in the
// corresponding internal subpackage.
package vjwt
