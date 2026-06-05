package jwt

import "testing"

// 对应 the utility toolkit-jwt JWTUtilTest（简化）。

func TestUtil_CreateAndVerify(t *testing.T) {
	key := []byte("1234567890")
	payload := map[string]any{
		"sub":  "1234567890",
		"name": "looly",
	}
	tok, err := CreateToken(payload, key)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !Verify(tok, key) {
		t.Fatalf("verify failed")
	}
	if Verify(tok, []byte("wrong")) {
		t.Fatalf("verify should fail with wrong key")
	}
}

func TestUtil_CreateWithSigner(t *testing.T) {
	signer := MustHMACSigner(AlgHS512, []byte("secret"))
	tok, err := CreateTokenWithSigner(map[string]any{"a": 1}, signer)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !VerifyWithSigner(tok, signer) {
		t.Fatalf("verify failed")
	}
}

func TestUtil_CreateAndVerifyStrictWithAlgorithm(t *testing.T) {
	key := []byte("secret")
	tok, err := CreateTokenWithAlgorithm(map[string]any{"a": 1}, key, AlgHS512)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	parsed, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Algorithm() != AlgHS512 {
		t.Fatalf("alg = %q, want %q", parsed.Algorithm(), AlgHS512)
	}
	if !VerifyStrict(tok, key) {
		t.Fatal("VerifyStrict failed")
	}
	if _, err := CreateTokenWithAlgorithm(map[string]any{"a": 1}, key, "bad"); err == nil {
		t.Fatal("CreateTokenWithAlgorithm bad alg error = nil")
	}
}

func TestUtil_ParseToken(t *testing.T) {
	tok := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9" +
		".eyJsb2dpblR5cGUiOiJsb2dpbiIsImxvZ2luSWQiOiJhZG1pbiIsImRldmljZSI6ImRlZmF1bHQtZGV2aWNlIiwiZWZmIjoxNjc4Mjg1NzEzOTM1LCJyblN0ciI6IkVuMTczWFhvWUNaaVZUWFNGOTNsN1pabGtOalNTd0pmIn0" +
		".wRe2soTaWYPhwcjxdzesDi1BgEm9D61K-mMT3fPc4YM"
	j, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := j.Payloads()
	if p["loginType"] != "login" {
		t.Fatalf("loginType: %v", p["loginType"])
	}
	// JSON 数字默认解析为 float64
	if v, ok := p["eff"].(float64); !ok || int64(v) != 1678285713935 {
		t.Fatalf("eff: %v (%T)", p["eff"], p["eff"])
	}
}

func TestUtil_CreateTokenWithHeaders(t *testing.T) {
	headers := map[string]any{HeaderKeyID: "kid-1"}
	payload := map[string]any{"a": 1}
	tok, err := CreateTokenWithHeaders(headers, payload, []byte("k"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	j, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if j.Header(HeaderKeyID) != "kid-1" {
		t.Fatalf("kid: %v", j.Header(HeaderKeyID))
	}
}

func TestUtil_ParseInvalid(t *testing.T) {
	if _, err := ParseToken(""); err == nil {
		t.Fatalf("expected error for blank token")
	}
	if _, err := ParseToken("not.a.jwt.too.many"); err == nil {
		t.Fatalf("expected error for malformed token")
	}
	if Verify("bad", []byte("k")) {
		t.Fatalf("expected verify=false for bad token")
	}
}
