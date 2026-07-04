package vjwt_test

import (
	"encoding/json"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vjwt"
)

func TestSignVerifyWithJSONProviders(t *testing.T) {
	marshalCalled := false
	unmarshalCalled := false
	marshal := func(v any) ([]byte, error) {
		marshalCalled = true
		return json.Marshal(v)
	}
	unmarshal := func(data []byte, v any) error {
		unmarshalCalled = true
		return json.Unmarshal(data, v)
	}

	jwt := vjwt.New().
		SetPayload(vjwt.JWTPayloadSubject, "alice").
		SetKey([]byte("secret"))
	token, err := jwt.SignOptsWithOptions(true, vjwt.WithJSONMarshalFunc(marshal))
	if err != nil {
		t.Fatalf("SignOptsWithOptions: %v", err)
	}
	parsed, err := vjwt.ParseTokenWithOptions(token, vjwt.WithJSONUnmarshalFunc(unmarshal))
	if err != nil {
		t.Fatalf("ParseTokenWithOptions: %v", err)
	}
	if !parsed.SetKey([]byte("secret")).Verify() {
		t.Fatal("parsed token should verify")
	}
	if !marshalCalled || !unmarshalCalled {
		t.Fatalf("JSON providers called marshal=%v unmarshal=%v", marshalCalled, unmarshalCalled)
	}
}

func TestJSONProviderErrorContract(t *testing.T) {
	marshalErr := errors.New("json marshal provider failed")
	_, err := vjwt.New().
		SetPayload(vjwt.JWTPayloadSubject, "alice").
		SetKey([]byte("secret")).
		SignOptsWithOptions(true, vjwt.WithJSONMarshalFunc(func(any) ([]byte, error) {
			return nil, marshalErr
		}))
	if !errors.Is(err, marshalErr) {
		t.Fatalf("SignOptsWithOptions error = %v, want marshal cause", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("SignOptsWithOptions error = %v, want ErrCodeInvalidInput", err)
	}
	var jwtErr *vjwt.JWTError
	if !errors.As(err, &jwtErr) {
		t.Fatalf("errors.As(err, *vjwt.JWTError) = false: %v", err)
	}

	token, err := vjwt.New().SetPayload(vjwt.JWTPayloadSubject, "alice").SetKey([]byte("secret")).Sign()
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	unmarshalErr := errors.New("json unmarshal provider failed")
	_, err = vjwt.ParseTokenWithOptions(token, vjwt.WithJSONUnmarshalFunc(func([]byte, any) error {
		return unmarshalErr
	}))
	if !errors.Is(err, unmarshalErr) {
		t.Fatalf("ParseTokenWithOptions error = %v, want unmarshal cause", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ParseTokenWithOptions error = %v, want ErrCodeInvalidInput", err)
	}
	if !errors.As(err, &jwtErr) {
		t.Fatalf("errors.As(err, *vjwt.JWTError) = false: %v", err)
	}
}
