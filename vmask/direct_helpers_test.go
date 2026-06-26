package vmask_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vmask"
)

func TestFacadeDirectMaskHelpers(t *testing.T) {
	if got := vmask.FirstMask("中文ab"); got != "中***" {
		t.Fatalf("FirstMask = %q", got)
	}
	if got := vmask.IDCardNum("1234567890", 2, 3); got != "12*****890" {
		t.Fatalf("IDCardNum = %q", got)
	}
	if got := vmask.IDCardNum("123", 2, 2); got != "" {
		t.Fatalf("IDCardNum invalid = %q", got)
	}
	if got := vmask.FixedPhone("01012345678"); got != "0101*****78" {
		t.Fatalf("FixedPhone = %q", got)
	}
	if got := vmask.MobilePhone("18049531999"); got != "180****1999" {
		t.Fatalf("MobilePhone = %q", got)
	}
	if got := vmask.Address("北京市朝阳区望京街道", 4); got != "北京市朝阳区****" {
		t.Fatalf("Address = %q", got)
	}
	if got := vmask.Email("a@example.com"); got != "a@example.com" {
		t.Fatalf("Email short local = %q", got)
	}
	if got := vmask.Password("密码ab"); got != "****" {
		t.Fatalf("Password = %q", got)
	}
	if got := vmask.CarLicense("粤B123456"); got != "粤B1****6" {
		t.Fatalf("CarLicense new energy = %q", got)
	}
	if got := vmask.CarLicense("too-long-plate"); got != "too-long-plate" {
		t.Fatalf("CarLicense invalid len = %q", got)
	}
	if got := vmask.BankCard("1234 5678 9012 3456"); got != "1234 **** **** 3456" {
		t.Fatalf("BankCard spaced = %q", got)
	}
	if got := vmask.BankCard("12345678"); got != "12345678" {
		t.Fatalf("BankCard short = %q", got)
	}
	if got := vmask.IPv4("localhost"); got != "localhost.*.*.*" {
		t.Fatalf("IPv4 no dot = %q", got)
	}
	if got := vmask.IPv6("localhost"); got != "localhost:*:*:*:*:*:*:*:*" {
		t.Fatalf("IPv6 no colon = %q", got)
	}
	if got := vmask.Passport("AB"); got != "**" {
		t.Fatalf("Passport short = %q", got)
	}
	if got := vmask.CreditCode("1234"); got != "****" {
		t.Fatalf("CreditCode short = %q", got)
	}
	if got := vmask.Hide("abcdef", -1, 99); got != "******" {
		t.Fatalf("Hide clamped = %q", got)
	}
	if got := vmask.Hide("abcdef", 4, 2); got != "abcdef" {
		t.Fatalf("Hide reversed = %q", got)
	}
}
