// Package vjwt provides public APIs for JWT utilities.
//
// HMAC, RSA-PSS, and ECDSA signers are the production-oriented signing paths.
// Unsigned alg=none tokens are always rejected and no signer is exposed for
// creating them.
//
// This package only acts as a facade. Concrete implementations live in the
// corresponding internal subpackage.
package vjwt
