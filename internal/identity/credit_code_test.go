package identity

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestParseCreditCode(t *testing.T) {
	info, err := ParseCreditCode("91350211M000100Y46")
	if err != nil {
		t.Fatalf("ParseCreditCode error = %v", err)
	}
	if info.Raw != "91350211M000100Y46" ||
		info.AdminDept != "9" ||
		info.OrgCategory != "1" ||
		info.RegionCode != "350211" ||
		info.OrgCode != "M000100Y4" ||
		info.CheckDigit != "6" {
		t.Fatalf("ParseCreditCode info = %#v", info)
	}
}

func TestParseCreditCodeInvalid(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{name: "length", code: "91350211M000100Y4"},
		{name: "char", code: "91350211M000100Y4I"},
		{name: "check", code: "91350211M000100Y44"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsValidCreditCode(tt.code) {
				t.Fatalf("IsValidCreditCode(%q) = true", tt.code)
			}
			_, err := ParseCreditCode(tt.code)
			if !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("ParseCreditCode error = %v, want invalid input", err)
			}
		})
	}
}
