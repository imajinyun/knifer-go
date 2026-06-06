package identity

import (
	"testing"
	"time"
)

func TestConvert15To18(t *testing.T) {
	got, ok := Convert15To18("130503670401001")
	if !ok || got != "130503196704010016" {
		t.Fatalf("Convert15To18() = %q, %v", got, ok)
	}

	got15, ok := Convert18To15(got)
	if !ok || got15 != "130503670401001" {
		t.Fatalf("Convert18To15() = %q, %v", got15, ok)
	}
}

func TestIsValidIDCard18(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"11010519491231002X", true},
		{"11010519491231002x", true},
		{"81000019980902013X", true},
		{"820000200009100032", true},
		{"83000019810715006X", true},
		{"11010519490231002X", false},
		{"99010519491231002X", false},
		{"110105194912310021", false},
	}
	for _, tt := range tests {
		if got := IsValidIDCard18(tt.id); got != tt.want {
			t.Fatalf("IsValidIDCard18(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
	if IsValidIDCard18WithIgnoreCase("11010519491231002x", false) {
		t.Fatal("IsValidIDCard18WithIgnoreCase should reject lowercase x when ignoreCase=false")
	}
}

func TestIsValidIDCard15(t *testing.T) {
	if !IsValidIDCard15("130503670401001") {
		t.Fatal("expected valid 15-digit ID card")
	}
	if IsValidIDCard15("130503990230001") {
		t.Fatal("expected invalid birthday to be rejected")
	}
}

func TestRegionCards(t *testing.T) {
	if !IsValidTWIDCard("A123456789") {
		t.Fatal("expected valid Taiwan card")
	}
	if !IsValidHKIDCard("A123456(3)") {
		t.Fatal("expected valid Hong Kong card")
	}
	info, ok := ParseRegionCard("A123456789")
	if !ok || info.Region != "台湾" || info.Gender != "M" || !info.Valid {
		t.Fatalf("ParseRegionCard Taiwan = %+v, %v", info, ok)
	}
	info, ok = ParseRegionCard("1571234(5)")
	if !ok || info.Region != "澳门" || !info.Valid {
		t.Fatalf("ParseRegionCard Macau = %+v, %v", info, ok)
	}
}

func TestIDCardFields(t *testing.T) {
	const id = "11010519491231002X"
	birth, ok := BirthString(id)
	if !ok || birth != "19491231" {
		t.Fatalf("BirthString() = %q, %v", birth, ok)
	}
	year, ok := Year(id)
	if !ok || year != 1949 {
		t.Fatalf("Year() = %d, %v", year, ok)
	}
	month, ok := Month(id)
	if !ok || month != 12 {
		t.Fatalf("Month() = %d, %v", month, ok)
	}
	day, ok := Day(id)
	if !ok || day != 31 {
		t.Fatalf("Day() = %d, %v", day, ok)
	}
	age, ok := AgeAt(id, time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local))
	if !ok || age != 74 {
		t.Fatalf("AgeAt(before birthday) = %d, %v", age, ok)
	}
	age, ok = AgeAt(id, time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local))
	if !ok || age != 75 {
		t.Fatalf("AgeAt(on birthday) = %d, %v", age, ok)
	}
	age, ok = AgeWithOptions(id, WithAgeTime(time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)))
	if !ok || age != 75 {
		t.Fatalf("AgeWithOptions(WithAgeTime) = %d, %v", age, ok)
	}
	age, ok = AgeWithOptions(id, WithAgeClock(func() time.Time {
		return time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local)
	}))
	if !ok || age != 74 {
		t.Fatalf("AgeWithOptions(WithAgeClock) = %d, %v", age, ok)
	}
	gender, ok := GenderOf(id)
	if !ok || gender != GenderFemale {
		t.Fatalf("GenderOf() = %d, %v", gender, ok)
	}
	province, ok := Province(id)
	if !ok || province != "北京" {
		t.Fatalf("Province() = %q, %v", province, ok)
	}
	district, ok := DistrictCode(id)
	if !ok || district != "110105" {
		t.Fatalf("DistrictCode() = %q, %v", district, ok)
	}
}

func TestParseIDCardAndHide(t *testing.T) {
	info, ok := ParseIDCard("11010519491231002X")
	if !ok || info.ProvinceCode != "11" || info.CityCode != "1101" || info.DistrictCode != "110105" {
		t.Fatalf("ParseIDCard() = %+v, %v", info, ok)
	}
	if got := Hide("11010519491231002X", 6, 14); got != "110105********002X" {
		t.Fatalf("Hide() = %q", got)
	}
}
