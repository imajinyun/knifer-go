package vform

import "testing"

func TestValidatorFacade(t *testing.T) {
	if !IsEmail("a@b.com") || IsEmail("bad") {
		t.Fatal("IsEmail failed")
	}
	if !IsMobile("13812345678") || IsMobile("12812345678") {
		t.Fatal("IsMobile failed")
	}
	if !IsURL("https://example.com") || !IsURL("ftp://example.com") || IsURL("/relative/path") {
		t.Fatal("IsURL failed")
	}
	if !IsIPv4("127.0.0.1") || IsIPv4("256.0.0.1") {
		t.Fatal("IsIPv4 failed")
	}
	if !IsIPv6("2001:db8::1") || IsIPv6("127.0.0.1") || IsIPv6("bad") {
		t.Fatal("IsIPv6 failed")
	}
	if !IsIDCard("11010519491231002X") || IsIDCard("110105194912310021") {
		t.Fatal("IsIDCard failed")
	}
	if !IsChinese("你好") || IsChinese("hello") {
		t.Fatal("IsChinese failed")
	}
	if !IsNumberStr("-3.14") || IsNumberStr("x") {
		t.Fatal("IsNumberStr failed")
	}
}

func TestValidatorFacadeWithOptions(t *testing.T) {
	const accepted = "accepted"

	if !IsEmailWithOptions(accepted, WithEmailMatcher(func(s string) bool { return s == accepted })) {
		t.Fatal("IsEmailWithOptions did not use custom matcher")
	}
	if !IsMobileWithOptions(accepted, WithMobileMatcher(func(s string) bool { return s == accepted })) {
		t.Fatal("IsMobileWithOptions did not use custom matcher")
	}
	if !IsIDCardWithOptions(accepted, WithIDCardMatcher(func(s string) bool { return s == accepted })) {
		t.Fatal("IsIDCardWithOptions did not use custom matcher")
	}
	if !IsChineseWithOptions(accepted, WithChineseMatcher(func(s string) bool { return s == accepted })) {
		t.Fatal("IsChineseWithOptions did not use custom matcher")
	}
	if !IsNumberStrWithOptions(accepted, WithNumberMatcher(func(s string) bool { return s == accepted })) {
		t.Fatal("IsNumberStrWithOptions did not use custom matcher")
	}
}

func TestValidatorFacadeNilOptionsUseDefaults(t *testing.T) {
	if !IsEmailWithOptions("a@b.com", nil, WithEmailMatcher(nil)) {
		t.Fatal("nil email options should keep default matcher")
	}
	if !IsMobileWithOptions("13812345678", nil, WithMobileMatcher(nil)) {
		t.Fatal("nil mobile options should keep default matcher")
	}
	if !IsIDCardWithOptions("11010519491231002X", nil, WithIDCardMatcher(nil)) {
		t.Fatal("nil ID card options should keep default matcher")
	}
	if !IsChineseWithOptions("你好", nil, WithChineseMatcher(nil)) {
		t.Fatal("nil Chinese options should keep default matcher")
	}
	if !IsNumberStrWithOptions("-3.14", nil, WithNumberMatcher(nil)) {
		t.Fatal("nil number options should keep default matcher")
	}
}

func TestValidatorFacadeBoundaries(t *testing.T) {
	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "email rejects missing domain", got: IsEmail("user@"), want: false},
		{name: "email accepts plus tag", got: IsEmail("user+tag@example.com"), want: true},
		{name: "mobile rejects too short", got: IsMobile("1381234567"), want: false},
		{name: "url rejects hostless scheme", got: IsURL("https:///path"), want: false},
		{name: "ipv4 rejects host port", got: IsIPv4("127.0.0.1:80"), want: false},
		{name: "ipv6 accepts compressed loopback", got: IsIPv6("::1"), want: true},
		{name: "id card rejects empty", got: IsIDCard(""), want: false},
		{name: "chinese rejects empty", got: IsChinese(""), want: false},
		{name: "number rejects plus sign", got: IsNumberStr("+3.14"), want: false},
		{name: "number accepts negative integer", got: IsNumberStr("-42"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}
