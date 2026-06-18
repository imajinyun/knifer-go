package vjwt_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vjwt"
)

func TestAlgorithmName(t *testing.T) {
	if got := vjwt.AlgorithmName(vjwt.JWTAlgPS256); got != "SHA256withRSA_PSS" {
		t.Fatalf("AlgorithmName(PS256) = %q", got)
	}
}

func TestFacadeJWTErrorf(t *testing.T) {
	err := vjwt.JWTErrorf("code %d: %s", 400, "bad request")
	if err == nil || err.Error() != "code 400: bad request" {
		t.Fatalf("JWTErrorf = %v", err)
	}
}
