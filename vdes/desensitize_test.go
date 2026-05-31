package vdes_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vdes"
)

func TestFacadeBuiltInRules(t *testing.T) {
	if got := vdes.Desensitized("18049531999", vdes.MobilePhoneType); got != "180****1999" {
		t.Fatalf("mobile: %q", got)
	}
	if got := vdes.ChineseName("段正淳"); got != "段**" {
		t.Fatalf("name: %q", got)
	}
	if got := vdes.Email("duandazhi-jack@gmail.com.cn"); got != "d*************@gmail.com.cn" {
		t.Fatalf("email: %q", got)
	}
	if got := vdes.BankCard("11011111222233333256"); got != "1101 **** **** **** 3256" {
		t.Fatalf("bank: %q", got)
	}
	if got := vdes.Passport("PJ1234567"); got != "PJ*****67" {
		t.Fatalf("passport: %q", got)
	}
	if vdes.DesensitizedPtr("x", vdes.ClearToNullType) != nil {
		t.Fatal("ClearToNullType should return nil")
	}
}
