package date

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestSolarToLunarKnownSpringFestival(t *testing.T) {
	tests := []struct {
		name   string
		year   int
		month  int
		day    int
		lunar  LunarDate
		zodiac string
	}{
		{
			name:  "2024 spring festival",
			year:  2024,
			month: 2,
			day:   10,
			lunar: LunarDate{
				Year:       2024,
				Month:      1,
				Day:        1,
				YearGanZhi: "甲辰",
			},
			zodiac: "龙",
		},
		{
			name:  "2023 spring festival",
			year:  2023,
			month: 1,
			day:   22,
			lunar: LunarDate{
				Year:       2023,
				Month:      1,
				Day:        1,
				YearGanZhi: "癸卯",
			},
			zodiac: "兔",
		},
		{
			name:  "2020 spring festival",
			year:  2020,
			month: 1,
			day:   25,
			lunar: LunarDate{
				Year:       2020,
				Month:      1,
				Day:        1,
				YearGanZhi: "庚子",
			},
			zodiac: "鼠",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SolarToLunar(tt.year, tt.month, tt.day)
			if err != nil {
				t.Fatalf("SolarToLunar error = %v", err)
			}
			if got.Year != tt.lunar.Year || got.Month != tt.lunar.Month || got.Day != tt.lunar.Day {
				t.Fatalf("SolarToLunar = %+v, want lunar %04d-%02d-%02d", got, tt.lunar.Year, tt.lunar.Month, tt.lunar.Day)
			}
			if got.YearGanZhi != tt.lunar.YearGanZhi {
				t.Fatalf("YearGanZhi = %q, want %q", got.YearGanZhi, tt.lunar.YearGanZhi)
			}
			if got.Zodiac != tt.zodiac {
				t.Fatalf("Zodiac = %q, want %q", got.Zodiac, tt.zodiac)
			}
		})
	}
}

func TestLunarToSolarRoundTrip(t *testing.T) {
	solar, err := LunarToSolar(2024, 1, 1, false)
	if err != nil {
		t.Fatalf("LunarToSolar error = %v", err)
	}
	if solar != (SolarDate{Year: 2024, Month: 2, Day: 10}) {
		t.Fatalf("LunarToSolar = %+v, want 2024-02-10", solar)
	}

	lunar, err := SolarToLunar(solar.Year, solar.Month, solar.Day)
	if err != nil {
		t.Fatalf("SolarToLunar round trip error = %v", err)
	}
	if lunar.Year != 2024 || lunar.Month != 1 || lunar.Day != 1 || lunar.IsLeapMonth {
		t.Fatalf("round trip lunar = %+v, want 2024-01-01 non-leap", lunar)
	}
}

func TestLunarLeapMonth(t *testing.T) {
	if got := LeapMonth(2020); got != 4 {
		t.Fatalf("LeapMonth(2020) = %d, want 4", got)
	}
	if !IsLeapMonth(2020, 4) {
		t.Fatalf("IsLeapMonth(2020, 4) = false, want true")
	}
	if got := LunarMonthDays(2020, 4, true); got != 29 {
		t.Fatalf("LunarMonthDays(2020, leap 4) = %d, want 29", got)
	}

	solar, err := LunarToSolar(2020, 4, 1, true)
	if err != nil {
		t.Fatalf("LunarToSolar leap month error = %v", err)
	}
	if solar != (SolarDate{Year: 2020, Month: 5, Day: 23}) {
		t.Fatalf("LunarToSolar leap month = %+v, want 2020-05-23", solar)
	}
}

func TestLunarYearDays(t *testing.T) {
	if got := LunarYearDays(2024); got != 354 {
		t.Fatalf("LunarYearDays(2024) = %d, want 354", got)
	}
	if got := LunarYearDays(2020); got != 384 {
		t.Fatalf("LunarYearDays(2020) = %d, want 384", got)
	}
}

func TestGanZhiAndZodiac(t *testing.T) {
	if got := YearGanZhi(2024); got != "甲辰" {
		t.Fatalf("YearGanZhi(2024) = %q, want 甲辰", got)
	}
	if got := Zodiac(2024); got != "龙" {
		t.Fatalf("Zodiac(2024) = %q, want 龙", got)
	}
	if got := DayGanZhi(2024, 2, 10); got == "" {
		t.Fatalf("DayGanZhi returned empty")
	}
	if got := MonthGanZhi(2024, 2); got == "" {
		t.Fatalf("MonthGanZhi returned empty")
	}
}

func TestSolarTerm(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		day   int
		want  string
	}{
		{name: "qingming 2024", year: 2024, month: 4, day: 4, want: "清明"},
		{name: "winter solstice 2024", year: 2024, month: 12, day: 21, want: "冬至"},
		{name: "non term day", year: 2024, month: 4, day: 5, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SolarTerm(tt.year, tt.month, tt.day); got != tt.want {
				t.Fatalf("SolarTerm(%d,%d,%d) = %q, want %q", tt.year, tt.month, tt.day, got, tt.want)
			}
		})
	}
}

func TestLunarInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "invalid solar date", err: lunarErr(SolarToLunar(2024, 2, 31))},
		{name: "solar out of range", err: lunarErr(SolarToLunar(1899, 12, 31))},
		{name: "lunar year out of range", err: solarErr(LunarToSolar(1899, 1, 1, false))},
		{name: "invalid lunar month", err: solarErr(LunarToSolar(2024, 13, 1, false))},
		{name: "invalid leap month", err: solarErr(LunarToSolar(2024, 1, 1, true))},
		{name: "invalid lunar day", err: solarErr(LunarToSolar(2024, 1, 31, false))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("error = %v, want ErrCodeInvalidInput", tt.err)
			}
		})
	}
}

func lunarErr(_ LunarDate, err error) error {
	return err
}

func solarErr(_ SolarDate, err error) error {
	return err
}
