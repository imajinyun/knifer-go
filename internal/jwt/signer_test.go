package jwt

import (
	"testing"
)

// 对应 the utility toolkit-jwt JWTSignerUtilTest 的简化版本。

func TestHMACSigner_HS256(t *testing.T) {
	s, err := NewHMACSigner(AlgHS256, []byte("1234567890"))
	if err != nil {
		t.Fatalf("new HS256: %v", err)
	}
	// 来自 the utility toolkit 固定 token 的 header / payload
	header := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
	payload := "eyJzdWIiOiIxMjM0NTY3ODkwIiwiYWRtaW4iOnRydWUsIm5hbWUiOiJsb29seSJ9"
	want := "U2aQkC2THYV9L0fTN-yBBI7gmo5xhmvMhATtu8v0zEA"
	got := s.Sign(header, payload)
	if got != want {
		t.Fatalf("hs256 sign mismatch: got=%s want=%s", got, want)
	}
	if !s.Verify(header, payload, want) {
		t.Fatalf("hs256 verify failed")
	}
}

func TestHMACSigner_AllAlgs(t *testing.T) {
	for _, alg := range []string{AlgHS256, AlgHS384, AlgHS512} {
		s, err := NewHMACSigner(alg, []byte("k"))
		if err != nil {
			t.Fatalf("%s: %v", alg, err)
		}
		sig := s.Sign("h", "p")
		if sig == "" {
			t.Fatalf("%s: empty sig", alg)
		}
		if !s.Verify("h", "p", sig) {
			t.Fatalf("%s: verify failed", alg)
		}
		if s.Verify("h", "p", "wrong") {
			t.Fatalf("%s: verify should fail with wrong sig", alg)
		}
	}
}

func TestHMACSigner_UnsupportedAlg(t *testing.T) {
	if _, err := NewHMACSigner("RS256", []byte("k")); err == nil {
		t.Fatalf("expected error for unsupported alg")
	}
}

func TestNoneSigner(t *testing.T) {
	s := NoneSigner()
	if s.Algorithm() != AlgNone {
		t.Fatalf("alg: %s", s.Algorithm())
	}
	if s.Sign("h", "p") != "" {
		t.Fatalf("none sign should be empty")
	}
	if !s.Verify("h", "p", "") {
		t.Fatalf("none verify '' failed")
	}
	if s.Verify("h", "p", "x") {
		t.Fatalf("none verify non-empty should fail")
	}
}

func TestCreateSigner(t *testing.T) {
	if _, ok := must(CreateSigner("none", nil))(t).(noneSigner); !ok {
		t.Fatalf("expected noneSigner")
	}
	if IsNoneAlg("") {
		t.Fatalf("empty alg must not be treated as none")
	}
	if _, err := CreateSigner("", []byte("k")); err == nil {
		t.Fatalf("expected error for empty alg")
	}
	s := must(CreateSigner(AlgHS256, []byte("k")))(t)
	if s.Algorithm() != AlgHS256 {
		t.Fatalf("alg: %s", s.Algorithm())
	}
}

func must(s JWTSigner, err error) func(*testing.T) JWTSigner {
	return func(t *testing.T) JWTSigner {
		t.Helper()
		if err != nil {
			t.Fatalf("create signer: %v", err)
		}
		return s
	}
}
