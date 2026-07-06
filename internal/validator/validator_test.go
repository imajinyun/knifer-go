package validator

import "testing"

func TestValidators(t *testing.T) {
	if !IsEmail("a@b.com") || IsEmail("abc") {
		t.Fatalf("IsEmail failed")
	}
	if !IsMobile("13812345678") || IsMobile("12812345678") {
		t.Fatalf("IsMobile failed")
	}
	if !IsURL("https://example.com") || !IsURL("ftp://x") || IsURL("/relative/path") || IsURL(" https://example.com") {
		t.Fatalf("IsURL failed")
	}
	if !IsIPv4("127.0.0.1") || IsIPv4("256.0.0.1") {
		t.Fatalf("IsIPv4 failed")
	}
	if !IsIPv6("2001:db8::1") || IsIPv6("127.0.0.1") || IsIPv6("bad") {
		t.Fatalf("IsIPv6 failed")
	}
	if !IsIDCard("11010519491231002X") || IsIDCard("110105194912310021") {
		t.Fatalf("IsIDCard failed")
	}
	if !IsChinese("你好") || IsChinese("hello") {
		t.Fatalf("IsChinese failed")
	}
	if !IsNumberStr("-3.14") || IsNumberStr("ab") {
		t.Fatalf("IsNumberStr failed")
	}
}

func TestValidatorsWithCustomMatchers(t *testing.T) {
	cases := []struct {
		name string
		got  bool
	}{
		{
			name: "email",
			got: IsEmailWithOptions("custom-email", WithEmailMatcher(func(s string) bool {
				return s == "custom-email"
			})),
		},
		{
			name: "mobile",
			got: IsMobileWithOptions("custom-mobile", WithMobileMatcher(func(s string) bool {
				return s == "custom-mobile"
			})),
		},
		{
			name: "id-card",
			got: IsIDCardWithOptions("custom-id", WithIDCardMatcher(func(s string) bool {
				return s == "custom-id"
			})),
		},
		{
			name: "chinese",
			got: IsChineseWithOptions("custom-chinese", WithChineseMatcher(func(s string) bool {
				return s == "custom-chinese"
			})),
		},
		{
			name: "number",
			got: IsNumberStrWithOptions("custom-number", WithNumberMatcher(func(s string) bool {
				return s == "custom-number"
			})),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.got {
				t.Fatal("custom matcher was not used")
			}
		})
	}
}

func TestValidatorOptionsFallbackToDefaults(t *testing.T) {
	if !IsEmailWithOptions("a@b.com", nil, WithEmailMatcher(nil)) {
		t.Fatal("nil email matcher should fallback to default")
	}
	if !IsMobileWithOptions("13812345678", nil, WithMobileMatcher(nil)) {
		t.Fatal("nil mobile matcher should fallback to default")
	}
	if !IsIDCardWithOptions("11010519491231002X", nil, WithIDCardMatcher(nil)) {
		t.Fatal("nil id card matcher should fallback to default")
	}
	if !IsChineseWithOptions("你好", nil, WithChineseMatcher(nil)) {
		t.Fatal("nil chinese matcher should fallback to default")
	}
	if !IsNumberStrWithOptions("-3.14", nil, WithNumberMatcher(nil)) {
		t.Fatal("nil number matcher should fallback to default")
	}
}

func TestApplyOptionsFallbacksAfterCustomOptionClearsMatchers(t *testing.T) {
	cfg := applyOptions([]Option{
		func(c *config) {
			c.email = nil
			c.mobile = nil
			c.idCard = nil
			c.chinese = nil
			c.number = nil
		},
	})
	if !cfg.email("a@b.com") {
		t.Fatal("email matcher should fallback to default")
	}
	if !cfg.mobile("13812345678") {
		t.Fatal("mobile matcher should fallback to default")
	}
	if !cfg.idCard("11010519491231002X") {
		t.Fatal("id card matcher should fallback to default")
	}
	if !cfg.chinese("你好") {
		t.Fatal("chinese matcher should fallback to default")
	}
	if !cfg.number("-3.14") {
		t.Fatal("number matcher should fallback to default")
	}
}

func TestNilValidatorOptionsDoNotClearPreviousMatchers(t *testing.T) {
	if !IsEmailWithOptions("custom-email", WithEmailMatcher(func(s string) bool {
		return s == "custom-email"
	}), WithEmailMatcher(nil)) {
		t.Fatal("nil email matcher cleared previous matcher")
	}
	if !IsMobileWithOptions("custom-mobile", WithMobileMatcher(func(s string) bool {
		return s == "custom-mobile"
	}), WithMobileMatcher(nil)) {
		t.Fatal("nil mobile matcher cleared previous matcher")
	}
	if !IsIDCardWithOptions("custom-id", WithIDCardMatcher(func(s string) bool {
		return s == "custom-id"
	}), WithIDCardMatcher(nil)) {
		t.Fatal("nil id card matcher cleared previous matcher")
	}
	if !IsChineseWithOptions("custom-chinese", WithChineseMatcher(func(s string) bool {
		return s == "custom-chinese"
	}), WithChineseMatcher(nil)) {
		t.Fatal("nil chinese matcher cleared previous matcher")
	}
	if !IsNumberStrWithOptions("custom-number", WithNumberMatcher(func(s string) bool {
		return s == "custom-number"
	}), WithNumberMatcher(nil)) {
		t.Fatal("nil number matcher cleared previous matcher")
	}
}

func TestValidatorBoundaryInputs(t *testing.T) {
	if IsChineseWithOptions("", WithChineseMatcher(func(string) bool { return true })) {
		t.Fatal("empty string should not be treated as Chinese even when matcher accepts it")
	}
	validEmails := []string{"first.last+tag@example.co", "a_b-1@example-domain.com"}
	for _, email := range validEmails {
		if !IsEmail(email) {
			t.Fatalf("IsEmail(%q) = false", email)
		}
	}
	invalidEmails := []string{"@example.com", "a@b", "a b@example.com"}
	for _, email := range invalidEmails {
		if IsEmail(email) {
			t.Fatalf("IsEmail(%q) = true", email)
		}
	}
	validNumbers := []string{"0", "-0", "123", "-123.456"}
	for _, number := range validNumbers {
		if !IsNumberStr(number) {
			t.Fatalf("IsNumberStr(%q) = false", number)
		}
	}
	invalidNumbers := []string{"", ".5", "1.", "+1", "1e3"}
	for _, number := range invalidNumbers {
		if IsNumberStr(number) {
			t.Fatalf("IsNumberStr(%q) = true", number)
		}
	}
}
