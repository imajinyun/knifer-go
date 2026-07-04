package vjwt_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
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

func TestFacadeValidateDateErrorContract(t *testing.T) {
	j := vjwt.New().SetPayload(vjwt.JWTPayloadExpiresAt, "not-a-time")
	err := vjwt.ValidateDate(j, time.Unix(1_700_000_000, 0), 0)
	if err == nil {
		t.Fatal("ValidateDate malformed claim error = nil")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ValidateDate error = %v, want ErrCodeInvalidInput", err)
	}
	var jwtErr *vjwt.JWTError
	if !errors.As(err, &jwtErr) {
		t.Fatalf("errors.As(err, *vjwt.JWTError) = false: %v", err)
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Fatalf("errors.As(err, *strconv.NumError) = false: %v", err)
	}
	if errors.Is(err, knifer.ErrCodeProviderFailure) {
		t.Fatalf("ValidateDate malformed claim should not be provider failure: %v", err)
	}
}
