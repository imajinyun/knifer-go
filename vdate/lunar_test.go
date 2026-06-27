package vdate

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestLunarFacade(t *testing.T) {
	lunar, err := SolarToLunar(2024, 2, 10)
	if err != nil {
		t.Fatalf("SolarToLunar error = %v", err)
	}
	if lunar.Year != 2024 || lunar.Month != 1 || lunar.Day != 1 {
		t.Fatalf("SolarToLunar = %+v, want lunar 2024-01-01", lunar)
	}
	if lunar.YearGanZhi != "甲辰" || lunar.Zodiac != "龙" {
		t.Fatalf("lunar metadata = %+v, want 甲辰/龙", lunar)
	}

	solar, err := LunarToSolar(2024, 1, 1, false)
	if err != nil {
		t.Fatalf("LunarToSolar error = %v", err)
	}
	if solar != (SolarDate{Year: 2024, Month: 2, Day: 10}) {
		t.Fatalf("LunarToSolar = %+v, want 2024-02-10", solar)
	}
}

func TestLunarFacadeHelpers(t *testing.T) {
	if got := LeapMonth(2020); got != 4 {
		t.Fatalf("LeapMonth(2020) = %d, want 4", got)
	}
	if !IsLeapMonth(2020, 4) {
		t.Fatalf("IsLeapMonth(2020, 4) = false, want true")
	}
	if got := LunarMonthDays(2020, 4, true); got != 29 {
		t.Fatalf("LunarMonthDays(2020, 4, true) = %d, want 29", got)
	}
	if got := LunarYearDays(2020); got != 384 {
		t.Fatalf("LunarYearDays(2020) = %d, want 384", got)
	}
	if got := Zodiac(2024); got != "龙" {
		t.Fatalf("Zodiac(2024) = %q, want 龙", got)
	}
	if got := YearGanZhi(2024); got != "甲辰" {
		t.Fatalf("YearGanZhi(2024) = %q, want 甲辰", got)
	}
	if got := SolarTerm(2024, 4, 4); got != "清明" {
		t.Fatalf("SolarTerm(2024, 4, 4) = %q, want 清明", got)
	}
}

func TestLunarFacadeErrorContract(t *testing.T) {
	_, err := SolarToLunar(1899, 12, 31)
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("SolarToLunar error = %v, want ErrCodeInvalidInput", err)
	}

	_, err = LunarToSolar(2024, 1, 31, false)
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("LunarToSolar error = %v, want ErrCodeInvalidInput", err)
	}
}
