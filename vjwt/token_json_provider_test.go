package vjwt_test

import (
	"encoding/json"
	"testing"

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
