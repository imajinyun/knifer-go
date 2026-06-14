package jwt

import "testing"

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
	// JSON numbers parse as float64 by default.
	if v, ok := p["eff"].(float64); !ok || int64(v) != 1678285713935 {
		t.Fatalf("eff: %v (%T)", p["eff"], p["eff"])
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
