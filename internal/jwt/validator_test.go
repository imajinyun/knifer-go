package jwt

import (
	"testing"
	"time"
)

// 对应 the utility toolkit-jwt JWTValidatorTest。

// TestExpiredAt 已过期的 token 应返回校验错误。
func TestExpiredAt(t *testing.T) {
	// 与 the utility toolkit 测试同一 token，exp=1477592 已过期
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0Nzc1OTJ9.isvT0Pqx0yjnZk53mUFSeYFJLDs-Ls9IsNAm86gIdZo"
	j, err := Of(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := ValidateDate(j, time.Now(), 0); err == nil {
		t.Fatalf("expected expired error")
	}
}

// TestIssueAt 签发时间晚于参考时间应失败。
func TestIssueAt(t *testing.T) {
	now := time.Now()
	tok, err := New().SetIssuedAt(now).SetKey([]byte("123456")).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	j, _ := Of(tok)
	yesterday := now.AddDate(0, 0, -1)
	if err := ValidateDate(j, yesterday, 0); err == nil {
		t.Fatalf("expected error: iat in future of yesterday")
	}
}

// TestIssueAtPass 签发时间不晚于参考时间应通过。
func TestIssueAtPass(t *testing.T) {
	now := time.Now()
	tok, err := New().SetIssuedAt(now).SetKey([]byte("123456")).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	j, _ := Of(tok)
	if err := ValidateDate(j, now, 0); err != nil {
		t.Fatalf("should pass: %v", err)
	}
}

// TestNotBefore nbf 晚于参考时间应失败。
func TestNotBefore(t *testing.T) {
	now := time.Now()
	j := New().SetNotBefore(now)
	yesterday := now.AddDate(0, 0, -1)
	if err := ValidateDate(j, yesterday, 0); err == nil {
		t.Fatalf("expected error: nbf later than now")
	}
}

// TestNotBeforePass nbf 不晚于参考时间应通过。
func TestNotBeforePass(t *testing.T) {
	now := time.Now()
	j := New().SetNotBefore(now)
	if err := ValidateDate(j, now, 0); err != nil {
		t.Fatalf("should pass: %v", err)
	}
}

// TestValidateAlgorithm 算法一致时校验通过。
func TestValidateAlgorithm(t *testing.T) {
	tok, err := New().SetNotBefore(time.Now()).SetKey([]byte("123456")).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	signer := MustHMACSigner(AlgHS256, []byte("123456"))
	if err := ValidateAlgorithm(tok, signer); err != nil {
		t.Fatalf("should pass: %v", err)
	}
}

// TestValidateAlgorithmMismatch 算法不一致应报错。
func TestValidateAlgorithmMismatch(t *testing.T) {
	tok, err := New().SetKey([]byte("123456")).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	signer := MustHMACSigner(AlgHS512, []byte("123456"))
	if err := ValidateAlgorithm(tok, signer); err == nil {
		t.Fatalf("expected algorithm mismatch error")
	}
}

// TestValidateExpired 校验整体合法性时过期 token 返回 false（leeway=0）。
func TestValidateExpired(t *testing.T) {
	// 与 the utility toolkit 测试 validateTest 中相同
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9." +
		"eyJpc3MiOiJNb0xpIiwiZXhwIjoxNjI0OTU4MDk0NTI4LCJpYXQiOjE2MjQ5NTgwMzQ1MjAsInVzZXIiOiJ1c2VyIn0." +
		"L0uB38p9sZrivbmP0VlDe--j_11YUXTu3TfHhfQhRKc"
	j, err := Of(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// 注意 the utility toolkit 该 token 的 exp 字段单位实际上看似毫秒（1624958094528）。
	// validate(0) 在 the utility toolkit 中返回 false，这里同样应不通过。
	if j.SetKey([]byte("1234567890")).Validate(0) {
		t.Fatalf("expected validate=false")
	}
}

// TestValidateDateExpired 直接构造已过期 JWT 应被校验拒绝。
func TestValidateDateExpired(t *testing.T) {
	exp, _ := time.Parse("2006-01-02 15:04:05", "2021-10-13 09:59:00")
	j := New().
		SetPayload("id", 123).
		SetPayload("username", "the utility toolkit").
		SetExpiresAt(exp)
	if err := ValidateDate(j, time.Now(), 0); err == nil {
		t.Fatalf("expected expired error")
	}
}

// TestValidateLeeway leeway 容忍区间内通过。
func TestValidateLeeway(t *testing.T) {
	now := time.Now()
	expired := now.Add(3 * time.Second)
	tok, err := New().
		SetPayload("sub", "blue-light").
		SetIssuedAt(now).
		SetNotBefore(expired).
		SetExpiresAt(expired).
		SetKey([]byte("123456")).
		Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	j, _ := Of(tok)
	// 4 秒前——超过 nbf 边界 1 秒，但 leeway=10 容许
	before := now.Add(-4 * time.Second)
	if err := ValidateDate(j, before, 10); err != nil {
		t.Fatalf("should pass with leeway: %v", err)
	}
	// 4 秒后——超过 exp 1 秒，但 leeway=10 仍容许
	after := now.Add(4 * time.Second)
	if err := ValidateDate(j, after, 10); err != nil {
		t.Fatalf("should pass with leeway: %v", err)
	}
}

func TestValidateWithOptions(t *testing.T) {
	now := time.Now()
	tok, err := New().
		SetPayload("sub", "options").
		SetIssuedAt(now).
		SetNotBefore(now.Add(3 * time.Second)).
		SetExpiresAt(now.Add(3 * time.Second)).
		SetKey([]byte("123456")).
		Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	j, err := Of(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	j.SetKey([]byte("123456"))
	if j.ValidateWithOptions(WithValidateTime(now.Add(-4 * time.Second))) {
		t.Fatal("ValidateWithOptions should reject token before nbf without leeway")
	}
	if !j.ValidateWithOptions(
		WithValidateClock(func() time.Time { return now.Add(-4 * time.Second) }),
		WithValidateLeeway(10),
	) {
		t.Fatal("ValidateWithOptions should accept token within configured leeway")
	}
}
