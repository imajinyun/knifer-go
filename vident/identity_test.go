package vident_test

import (
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vident"
)

func TestIDCardFacade(t *testing.T) {
	converted, ok := vident.Convert15To18("130503670401001")
	if !ok || converted != "130503196704010016" {
		t.Fatalf("Convert15To18() = %q, %v", converted, ok)
	}
	if !vident.IsValidIDCard("11010519491231002X") {
		t.Fatal("expected valid ID card")
	}
	birth, ok := vident.BirthDate("11010519491231002X")
	if !ok || birth.Format("2006-01-02") != "1949-12-31" {
		t.Fatalf("BirthDate() = %v, %v", birth, ok)
	}
	age, ok := vident.AgeAt("11010519491231002X", time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local))
	if !ok || age != 75 {
		t.Fatalf("AgeAt() = %d, %v", age, ok)
	}
	age, ok = vident.AgeWithOptions(
		"11010519491231002X",
		vident.WithAgeClock(func() time.Time { return time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local) }),
	)
	if !ok || age != 74 {
		t.Fatalf("AgeWithOptions() = %d, %v", age, ok)
	}
	info, ok := vident.ParseIDCard("11010519491231002X")
	if !ok || info.Province != "北京" || info.Gender != vident.GenderFemale {
		t.Fatalf("ParseIDCard() = %+v, %v", info, ok)
	}
}

func TestRegionCardFacade(t *testing.T) {
	info, ok := vident.ParseRegionCard("A123456(3)")
	if !ok || info.Region != "香港" || !info.Valid {
		t.Fatalf("ParseRegionCard() = %+v, %v", info, ok)
	}
	if !vident.IsValidTWIDCard("A123456789") {
		t.Fatal("expected valid Taiwan card")
	}
}
