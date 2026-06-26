package vmask_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vmask"
)

func TestFacadeBuiltInRules(t *testing.T) {
	if got := vmask.Masked("18049531999", vmask.MobilePhoneType); got != "180****1999" {
		t.Fatalf("mobile: %q", got)
	}
	if got := vmask.ChineseName("段正淳"); got != "段**" {
		t.Fatalf("name: %q", got)
	}
	if got := vmask.Email("duandazhi-jack@gmail.com.cn"); got != "d*************@gmail.com.cn" {
		t.Fatalf("email: %q", got)
	}
	if got := vmask.BankCard("11011111222233333256"); got != "1101 **** **** **** 3256" {
		t.Fatalf("bank: %q", got)
	}
	if got := vmask.Passport("PJ1234567"); got != "PJ*****67" {
		t.Fatalf("passport: %q", got)
	}
	if vmask.MaskedPtr("x", vmask.ClearToNullType) != nil {
		t.Fatal("ClearToNullType should return nil")
	}
}
