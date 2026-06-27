package vident

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestCreditCodeFacade(t *testing.T) {
	info, err := ParseCreditCode("91350211M000100Y46")
	if err != nil {
		t.Fatalf("ParseCreditCode error = %v", err)
	}
	if !IsValidCreditCode(info.Raw) || info.RegionCode != "350211" {
		t.Fatalf("credit code facade info = %#v", info)
	}
	_, err = ParseCreditCode("91350211M000100Y44")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ParseCreditCode invalid error = %v", err)
	}
}
