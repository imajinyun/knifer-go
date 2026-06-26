package vjwt_test

import (
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vjwt"
)

func TestFacadeDateValidationOptions(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	j := vjwt.New().
		SetPayload(vjwt.JWTPayloadNotBefore, now.Add(-time.Minute).Unix()).
		SetPayload(vjwt.JWTPayloadExpiresAt, now.Add(time.Minute).Unix()).
		SetPayload(vjwt.JWTPayloadIssuedAt, now.Add(-time.Second).Unix()).
		SetKey([]byte("0123456789abcdef0123456789abcdef"))
	token, err := j.Sign()
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parsed, err := vjwt.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	parsed.SetKey([]byte("0123456789abcdef0123456789abcdef"))
	if err := vjwt.ValidateJWTDate(parsed, now, 0); err != nil {
		t.Fatalf("ValidateJWTDate: %v", err)
	}
	if err := vjwt.ValidateDate(parsed, now, 0); err != nil {
		t.Fatalf("ValidateDate: %v", err)
	}
	if !parsed.ValidateWithOptions(vjwt.WithValidateTime(now), vjwt.WithValidateClock(func() time.Time { return now }), vjwt.WithValidateLeeway(0)) {
		t.Fatal("ValidateWithOptions = false")
	}

	expired := vjwt.New().SetPayload(vjwt.JWTPayloadExpiresAt, now.Add(-2*time.Second).Unix()).SetKey([]byte("0123456789abcdef0123456789abcdef"))
	if err := vjwt.ValidateDate(expired, now, 1); err == nil {
		t.Fatal("ValidateDate(expired) error = nil")
	}
}
