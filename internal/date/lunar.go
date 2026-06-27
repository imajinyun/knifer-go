package date

import (
	"time"
)

// SolarDate represents a Gregorian calendar date.
type SolarDate struct {
	Year  int
	Month int
	Day   int
}

// LunarDate represents a Chinese lunar calendar date.
type LunarDate struct {
	Year        int
	Month       int
	Day         int
	IsLeapMonth bool
	YearGanZhi  string
	MonthGanZhi string
	DayGanZhi   string
	Zodiac      string
}

const (
	minLunarYear = 1900
	maxLunarYear = 2100
)

var (
	lunarBaseDate   = time.Date(1900, time.January, 31, 0, 0, 0, 0, time.UTC)
	heavenlyStems   = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	earthlyBranches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	zodiacAnimals   = []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	solarTermNames  = []string{
		"小寒", "大寒", "立春", "雨水", "惊蛰", "春分",
		"清明", "谷雨", "立夏", "小满", "芒种", "夏至",
		"小暑", "大暑", "立秋", "处暑", "白露", "秋分",
		"寒露", "霜降", "立冬", "小雪", "大雪", "冬至",
	}
	solarTermMinutes = []int{
		0, 21208, 42467, 63836, 85337, 107014,
		128867, 150921, 173149, 195551, 218072, 240693,
		263343, 285989, 308563, 331033, 353350, 375494,
		397447, 419210, 440795, 462224, 483532, 504758,
	}
)

// lunarInfo stores month-size and leap-month data for 1900 through 2100.
// Bits 16..5 encode lunar months 1..12; bit set means 30 days, unset means 29.
// Bit 16 also encodes leap-month length; low 4 bits encode the leap month.
var lunarInfo = []int{
	0x04bd8, 0x04ae0, 0x0a570, 0x054d5, 0x0d260, 0x0d950, 0x16554, 0x056a0, 0x09ad0, 0x055d2,
	0x04ae0, 0x0a5b6, 0x0a4d0, 0x0d250, 0x1d255, 0x0b540, 0x0d6a0, 0x0ada2, 0x095b0, 0x14977,
	0x04970, 0x0a4b0, 0x0b4b5, 0x06a50, 0x06d40, 0x1ab54, 0x02b60, 0x09570, 0x052f2, 0x04970,
	0x06566, 0x0d4a0, 0x0ea50, 0x06e95, 0x05ad0, 0x02b60, 0x186e3, 0x092e0, 0x1c8d7, 0x0c950,
	0x0d4a0, 0x1d8a6, 0x0b550, 0x056a0, 0x1a5b4, 0x025d0, 0x092d0, 0x0d2b2, 0x0a950, 0x0b557,
	0x06ca0, 0x0b550, 0x15355, 0x04da0, 0x0a5d0, 0x14573, 0x052d0, 0x0a9a8, 0x0e950, 0x06aa0,
	0x0aea6, 0x0ab50, 0x04b60, 0x0aae4, 0x0a570, 0x05260, 0x0f263, 0x0d950, 0x05b57, 0x056a0,
	0x096d0, 0x04dd5, 0x04ad0, 0x0a4d0, 0x0d4d4, 0x0d250, 0x0d558, 0x0b540, 0x0b6a0, 0x195a6,
	0x095b0, 0x049b0, 0x0a974, 0x0a4b0, 0x0b27a, 0x06a50, 0x06d40, 0x0af46, 0x0ab60, 0x09570,
	0x04af5, 0x04970, 0x064b0, 0x074a3, 0x0ea50, 0x06b58, 0x05ac0, 0x0ab60, 0x096d5, 0x092e0,
	0x0c960, 0x0d954, 0x0d4a0, 0x0da50, 0x07552, 0x056a0, 0x0abb7, 0x025d0, 0x092d0, 0x0cab5,
	0x0a950, 0x0b4a0, 0x0baa4, 0x0ad50, 0x055d9, 0x04ba0, 0x0a5b0, 0x15176, 0x052b0, 0x0a930,
	0x07954, 0x06aa0, 0x0ad50, 0x05b52, 0x04b60, 0x0a6e6, 0x0a4e0, 0x0d260, 0x0ea65, 0x0d530,
	0x05aa0, 0x076a3, 0x096d0, 0x04bd7, 0x04ad0, 0x0a4d0, 0x1d0b6, 0x0d250, 0x0d520, 0x0dd45,
	0x0b5a0, 0x056d0, 0x055b2, 0x049b0, 0x0a577, 0x0a4b0, 0x0aa50, 0x1b255, 0x06d20, 0x0ada0,
	0x14b63, 0x09370, 0x049f8, 0x04970, 0x064b0, 0x168a6, 0x0ea50, 0x06b20, 0x1a6c4, 0x0aae0,
	0x0a2e0, 0x0d2e3, 0x0c960, 0x0d557, 0x0d4a0, 0x0da50, 0x05d55, 0x056a0, 0x0a6d0, 0x055d4,
	0x052d0, 0x0a9b8, 0x0a950, 0x0b4a0, 0x0b6a6, 0x0ad50, 0x055a0, 0x0aba4, 0x0a5b0, 0x052b0,
	0x0b273, 0x06930, 0x07337, 0x06aa0, 0x0ad50, 0x14b55, 0x04b60, 0x0a570, 0x054e4, 0x0d160,
	0x0e968, 0x0d520, 0x0daa0, 0x16aa6, 0x056d0, 0x04ae0, 0x0a9d4, 0x0a2d0, 0x0d150, 0x0f252,
	0x0d520,
}

// SolarToLunar converts a Gregorian date to the Chinese lunar calendar.
func SolarToLunar(year, month, day int) (LunarDate, error) {
	solar, err := normalizeSolarDate(year, month, day)
	if err != nil {
		return LunarDate{}, err
	}
	if solar.Before(lunarBaseDate) || solar.After(maxSolarDate()) {
		return LunarDate{}, invalidDateInputf("solar date out of supported lunar range: %04d-%02d-%02d", year, month, day)
	}

	offset := int(solar.Sub(lunarBaseDate).Hours() / 24)
	lunarYear := minLunarYear
	for ; lunarYear <= maxLunarYear; lunarYear++ {
		days := LunarYearDays(lunarYear)
		if offset < days {
			break
		}
		offset -= days
	}
	if lunarYear > maxLunarYear {
		return LunarDate{}, invalidDateInputf("solar date out of supported lunar range: %04d-%02d-%02d", year, month, day)
	}

	leapMonth := LeapMonth(lunarYear)
	isLeap := false
	lunarMonth := 1
	for ; lunarMonth <= 12; lunarMonth++ {
		days := LunarMonthDays(lunarYear, lunarMonth, isLeap)
		if offset < days {
			break
		}
		offset -= days
		if leapMonth == lunarMonth && !isLeap {
			isLeap = true
			lunarMonth--
		} else {
			isLeap = false
		}
	}

	return buildLunarDate(lunarYear, lunarMonth, offset+1, isLeap, solar), nil
}

// LunarToSolar converts a Chinese lunar date to the Gregorian calendar.
func LunarToSolar(year, month, day int, isLeapMonth bool) (SolarDate, error) {
	if !validLunarYear(year) {
		return SolarDate{}, invalidDateInputf("lunar year out of supported range: %d", year)
	}
	if month < 1 || month > 12 {
		return SolarDate{}, invalidDateInputf("invalid lunar month: %d", month)
	}
	if isLeapMonth && LeapMonth(year) != month {
		return SolarDate{}, invalidDateInputf("lunar year %d has no leap month %d", year, month)
	}
	monthDays := LunarMonthDays(year, month, isLeapMonth)
	if day < 1 || day > monthDays {
		return SolarDate{}, invalidDateInputf("invalid lunar day: %d", day)
	}

	offset := 0
	for y := minLunarYear; y < year; y++ {
		offset += LunarYearDays(y)
	}
	leapMonth := LeapMonth(year)
	for m := 1; m < month; m++ {
		offset += LunarMonthDays(year, m, false)
		if leapMonth == m {
			offset += LunarMonthDays(year, m, true)
		}
	}
	if isLeapMonth {
		offset += LunarMonthDays(year, month, false)
	}
	offset += day - 1

	solar := lunarBaseDate.AddDate(0, 0, offset)
	return SolarDate{Year: solar.Year(), Month: int(solar.Month()), Day: solar.Day()}, nil
}

// LeapMonth returns the leap lunar month for year, or 0 when the year has none.
func LeapMonth(year int) int {
	if !validLunarYear(year) {
		return 0
	}
	return lunarInfo[year-minLunarYear] & 0xf
}

// IsLeapMonth reports whether month is the leap month in the lunar year.
func IsLeapMonth(year, month int) bool {
	return validLunarYear(year) && month >= 1 && month <= 12 && LeapMonth(year) == month
}

// LunarMonthDays returns the day count for a lunar month.
func LunarMonthDays(year, month int, isLeapMonth bool) int {
	if !validLunarYear(year) || month < 1 || month > 12 {
		return 0
	}
	info := lunarInfo[year-minLunarYear]
	if isLeapMonth {
		if LeapMonth(year) != month {
			return 0
		}
		if info&0x10000 != 0 {
			return 30
		}
		return 29
	}
	if info&(0x10000>>month) != 0 {
		return 30
	}
	return 29
}

// LunarYearDays returns the day count for the lunar year.
func LunarYearDays(year int) int {
	if !validLunarYear(year) {
		return 0
	}
	days := 348
	info := lunarInfo[year-minLunarYear]
	for bit := 0x8000; bit > 0x8; bit >>= 1 {
		if info&bit != 0 {
			days++
		}
	}
	if leap := LeapMonth(year); leap != 0 {
		days += LunarMonthDays(year, leap, true)
	}
	return days
}

// Zodiac returns the Chinese zodiac animal for a Gregorian or lunar year.
func Zodiac(year int) string {
	if year <= 0 {
		return ""
	}
	return zodiacAnimals[(year-4)%12]
}

// YearGanZhi returns the sexagenary cycle name for year.
func YearGanZhi(year int) string {
	if year <= 0 {
		return ""
	}
	return ganZhi(year - 4)
}

// MonthGanZhi returns an approximate sexagenary cycle name for a Gregorian month.
func MonthGanZhi(year, month int) string {
	if year <= 0 || month < 1 || month > 12 {
		return ""
	}
	stem := (year*2 + month + 3) % 10
	branch := (month + 1) % 12
	return heavenlyStems[stem] + earthlyBranches[branch]
}

// DayGanZhi returns the sexagenary cycle name for a Gregorian date.
func DayGanZhi(year, month, day int) string {
	solar, err := normalizeSolarDate(year, month, day)
	if err != nil {
		return ""
	}
	offset := int(solar.Sub(lunarBaseDate).Hours() / 24)
	return ganZhi(offset + 40)
}

// SolarTerm returns the solar term name that falls on the Gregorian date.
// It returns an empty string when the date is not a solar term day.
func SolarTerm(year, month, day int) string {
	if year < minLunarYear || year > maxLunarYear || month < 1 || month > 12 {
		return ""
	}
	for i := (month - 1) * 2; i <= (month-1)*2+1; i++ {
		if solarTermDay(year, i) == day {
			return solarTermNames[i]
		}
	}
	return ""
}

func buildLunarDate(year, month, day int, isLeap bool, solar time.Time) LunarDate {
	return LunarDate{
		Year:        year,
		Month:       month,
		Day:         day,
		IsLeapMonth: isLeap,
		YearGanZhi:  YearGanZhi(year),
		MonthGanZhi: MonthGanZhi(solar.Year(), int(solar.Month())),
		DayGanZhi:   DayGanZhi(solar.Year(), int(solar.Month()), solar.Day()),
		Zodiac:      Zodiac(year),
	}
}

func normalizeSolarDate(year, month, day int) (time.Time, error) {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}, invalidDateInputf("invalid solar date: %04d-%02d-%02d", year, month, day)
	}
	solar := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if solar.Year() != year || int(solar.Month()) != month || solar.Day() != day {
		return time.Time{}, invalidDateInputf("invalid solar date: %04d-%02d-%02d", year, month, day)
	}
	return solar, nil
}

func maxSolarDate() time.Time {
	offset := 0
	for y := minLunarYear; y <= maxLunarYear; y++ {
		offset += LunarYearDays(y)
	}
	return lunarBaseDate.AddDate(0, 0, offset-1)
}

func validLunarYear(year int) bool {
	return year >= minLunarYear && year <= maxLunarYear
}

func ganZhi(offset int) string {
	offset %= 60
	if offset < 0 {
		offset += 60
	}
	return heavenlyStems[offset%10] + earthlyBranches[offset%12]
}

func solarTermDay(year, termIndex int) int {
	const tropicalYearMillis = 31556925974.7
	base := time.Date(1900, time.January, 6, 2, 5, 0, 0, time.UTC)
	millis := tropicalYearMillis*float64(year-1900) + float64(solarTermMinutes[termIndex])*60*1000
	return base.Add(time.Duration(millis) * time.Millisecond).Day()
}
