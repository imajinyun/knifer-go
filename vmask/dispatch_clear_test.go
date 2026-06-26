package vmask_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vmask"
)

func TestFacadeMaskDispatchAndClearHelpers(t *testing.T) {
	tests := []struct {
		name string
		in   string
		typ  vmask.Type
		want string
	}{
		{name: "user id", in: "12345", typ: vmask.UserID, want: "0"},
		{name: "id card", in: "11010519491231002X", typ: vmask.IDCard, want: "1***************2X"},
		{name: "fixed phone", in: "01012345678", typ: vmask.FixedPhoneType, want: "0101*****78"},
		{name: "address", in: "北京市朝阳区望京街道", typ: vmask.AddressType, want: "北京********"},
		{name: "password", in: "secret", typ: vmask.PasswordType, want: "******"},
		{name: "car license", in: "京A12345", typ: vmask.CarLicenseType, want: "京A1***5"},
		{name: "ipv4", in: "192.168.1.10", typ: vmask.IPv4Type, want: "192.*.*.*"},
		{name: "ipv6", in: "2001:db8::1", typ: vmask.IPv6Type, want: "2001:*:*:*:*:*:*:*:*"},
		{name: "credit code", in: "91350211M000100Y43", typ: vmask.CreditCodeType, want: "9135**********0Y43"},
		{name: "first mask", in: "hello", typ: vmask.FirstMaskType, want: "h****"},
		{name: "clear empty", in: "hello", typ: vmask.ClearToEmptyType, want: ""},
		{name: "unknown type", in: "hello", typ: vmask.Type(999), want: "hello"},
		{name: "blank input", in: "  ", typ: vmask.PasswordType, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vmask.Masked(tt.in, tt.typ); got != tt.want {
				t.Fatalf("Masked(%q, %v) = %q, want %q", tt.in, tt.typ, got, tt.want)
			}
		})
	}

	if got := vmask.Clear(); got != "" {
		t.Fatalf("Clear = %q", got)
	}
	if got := vmask.ClearToNil(); got != nil {
		t.Fatalf("ClearToNil = %v", got)
	}
	if got := vmask.UserIDValue(); got != 0 {
		t.Fatalf("UserIDValue = %d", got)
	}
	ptr := vmask.MaskedPtr("hello", vmask.FirstMaskType)
	if ptr == nil || *ptr != "h****" {
		t.Fatalf("MaskedPtr FirstMask = %v", ptr)
	}
}
