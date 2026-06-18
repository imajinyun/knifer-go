package jwt

import (
	"testing"
	"time"
)

func TestWithJSONMarshalFunc(t *testing.T) {
	cfg := jsonConfig{}
	fn := func(any) ([]byte, error) { return nil, nil }
	opt := WithJSONMarshalFunc(fn)
	opt(&cfg)
	if cfg.marshal == nil {
		t.Fatal("WithJSONMarshalFunc did not set marshal")
	}
}

func TestWithJSONUnmarshalFunc(t *testing.T) {
	cfg := jsonConfig{}
	fn := func([]byte, any) error { return nil }
	opt := WithJSONUnmarshalFunc(fn)
	opt(&cfg)
	if cfg.unmarshal == nil {
		t.Fatal("WithJSONUnmarshalFunc did not set unmarshal")
	}
}

func TestApplyJSONOptions(t *testing.T) {
	cfg := applyJSONOptions(nil)
	if cfg.marshal == nil || cfg.unmarshal == nil {
		t.Fatal("applyJSONOptions with nil should use defaults")
	}

	cfg2 := applyJSONOptions([]JSONOption{nil})
	if cfg2.marshal == nil || cfg2.unmarshal == nil {
		t.Fatal("applyJSONOptions with nil option should use defaults")
	}
}

func TestWithTokenJSONOptions(t *testing.T) {
	opt := WithTokenJSONOptions(WithJSONMarshalFunc(func(any) ([]byte, error) { return nil, nil }))
	cfg := tokenConfig{}
	opt(&cfg)
	if len(cfg.json) != 1 {
		t.Fatalf("WithTokenJSONOptions did not append json options, got %d", len(cfg.json))
	}
}

func TestApplyTokenOptionsNil(t *testing.T) {
	cfg := applyTokenOptions(nil)
	if cfg.alg != "" {
		t.Fatalf("applyTokenOptions nil = %+v, want zero", cfg)
	}
}

func TestApplyTokenOptionsNilGuard(t *testing.T) {
	cfg := applyTokenOptions([]TokenOption{nil})
	if cfg.alg != "" {
		t.Fatalf("applyTokenOptions with nil option should not panic")
	}
}

func TestWithTokenSigner(t *testing.T) {
	opt := WithTokenSigner(nil)
	cfg := tokenConfig{}
	opt(&cfg)
	if cfg.signer != nil {
		t.Fatal("WithTokenSigner(nil) should set nil signer")
	}
}

func TestWithValidateTime(t *testing.T) {
	opt := WithValidateTime(time.Now())
	cfg := validateConfig{}
	opt(&cfg)
	if cfg.now == nil {
		t.Fatal("WithValidateTime did not set now")
	}
}

func TestWithValidateLeeway(t *testing.T) {
	opt := WithValidateLeeway(10)
	cfg := validateConfig{}
	opt(&cfg)
	if cfg.leeway != 10 {
		t.Fatalf("WithValidateLeeway = %d, want 10", cfg.leeway)
	}
}