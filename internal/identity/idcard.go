package identity

import (
	"regexp"
	"strings"
	"time"
)

const (
	chinaIDMinLength = 15
	chinaIDMaxLength = 18
)

type ageConfig struct {
	clock func() time.Time
}

type birthConfig struct {
	location *time.Location
}

type idCardConfig struct {
	digits func(string) bool
	tw     func(string) bool
	macau  func(string) bool
	hk     func(string) bool
}

// AgeOption customizes AgeWithOptions.
type AgeOption func(*ageConfig)

// BirthOption customizes birthday parsing helpers.
type BirthOption func(*birthConfig)

// IDCardOption customizes identity-card validation helpers per call.
type IDCardOption func(*idCardConfig)

// WithDigitsMatcher sets the decimal-digits matcher used by mainland ID card helpers.
func WithDigitsMatcher(matcher func(string) bool) IDCardOption {
	return func(c *idCardConfig) { c.digits = matcher }
}

// WithTWCardMatcher sets the format matcher used by Taiwan ID card helpers.
func WithTWCardMatcher(matcher func(string) bool) IDCardOption {
	return func(c *idCardConfig) { c.tw = matcher }
}

// WithMacauCardMatcher sets the format matcher used by Macau ID card helpers.
func WithMacauCardMatcher(matcher func(string) bool) IDCardOption {
	return func(c *idCardConfig) { c.macau = matcher }
}

// WithHKCardMatcher sets the format matcher used by Hong Kong ID card helpers.
func WithHKCardMatcher(matcher func(string) bool) IDCardOption {
	return func(c *idCardConfig) { c.hk = matcher }
}

// WithAgeTime sets the time used by AgeWithOptions.
func WithAgeTime(at time.Time) AgeOption {
	return func(c *ageConfig) { c.clock = func() time.Time { return at } }
}

// WithAgeClock sets the clock used by AgeWithOptions.
func WithAgeClock(clock func() time.Time) AgeOption {
	return func(c *ageConfig) {
		if clock != nil {
			c.clock = clock
		}
	}
}

// WithBirthLocation sets the location used to parse yyyyMMdd birthdays.
func WithBirthLocation(location *time.Location) BirthOption {
	return func(c *birthConfig) {
		if location != nil {
			c.location = location
		}
	}
}

func applyAgeOptions(opts []AgeOption) ageConfig {
	cfg := ageConfig{clock: time.Now}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.clock == nil {
		cfg.clock = time.Now
	}
	return cfg
}

func applyBirthOptions(opts []BirthOption) birthConfig {
	cfg := birthConfig{location: time.Local}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.location == nil {
		cfg.location = time.Local
	}
	return cfg
}

func applyIDCardOptions(opts []IDCardOption) idCardConfig {
	cfg := idCardConfig{
		digits: rxDigits.MatchString,
		tw:     rxTWCard.MatchString,
		macau:  rxMacauID.MatchString,
		hk:     rxHKIDCard.MatchString,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.digits == nil {
		cfg.digits = rxDigits.MatchString
	}
	if cfg.tw == nil {
		cfg.tw = rxTWCard.MatchString
	}
	if cfg.macau == nil {
		cfg.macau = rxMacauID.MatchString
	}
	if cfg.hk == nil {
		cfg.hk = rxHKIDCard.MatchString
	}
	return cfg
}

var (
	idCardPower = [...]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	cityCodes   = map[string]string{
		"11": "北京",
		"12": "天津",
		"13": "河北",
		"14": "山西",
		"15": "内蒙古",
		"21": "辽宁",
		"22": "吉林",
		"23": "黑龙江",
		"31": "上海",
		"32": "江苏",
		"33": "浙江",
		"34": "安徽",
		"35": "福建",
		"36": "江西",
		"37": "山东",
		"41": "河南",
		"42": "湖北",
		"43": "湖南",
		"44": "广东",
		"45": "广西",
		"46": "海南",
		"50": "重庆",
		"51": "四川",
		"52": "贵州",
		"53": "云南",
		"54": "西藏",
		"61": "陕西",
		"62": "甘肃",
		"63": "青海",
		"64": "宁夏",
		"65": "新疆",
		"71": "台湾",
		"81": "香港",
		"82": "澳门",
		"83": "台湾",
		"91": "国外",
	}
	twFirstCode = map[byte]int{
		'A': 10,
		'B': 11,
		'C': 12,
		'D': 13,
		'E': 14,
		'F': 15,
		'G': 16,
		'H': 17,
		'J': 18,
		'K': 19,
		'L': 20,
		'M': 21,
		'N': 22,
		'P': 23,
		'Q': 24,
		'R': 25,
		'S': 26,
		'T': 27,
		'U': 28,
		'V': 29,
		'X': 30,
		'Y': 31,
		'W': 32,
		'Z': 33,
		'I': 34,
		'O': 35,
	}
	rxDigits   = regexp.MustCompile(`^\d+$`)
	rxTWCard   = regexp.MustCompile(`^[a-zA-Z][0-9]{9}$`)
	rxMacauID  = regexp.MustCompile(`^[157][0-9]{6}\(?[0-9A-Z]\)?$`)
	rxHKIDCard = regexp.MustCompile(`^[A-Z]{1,2}[0-9]{6}\(?[0-9A]\)?$`)
)

// Gender identifies the gender encoded in an identity card number.
type Gender int

const (
	// GenderUnknown means the gender is not encoded or cannot be determined.
	GenderUnknown Gender = -1
	// GenderFemale represents female.
	GenderFemale Gender = 0
	// GenderMale represents male.
	GenderMale Gender = 1
)

// IDCardInfo contains parsed information from a mainland China identity card.
type IDCardInfo struct {
	ProvinceCode string
	Province     string
	CityCode     string
	DistrictCode string
	Birth        time.Time
	Gender       Gender
	Age          int
}

// RegionCardInfo contains parsed validation information for Hong Kong, Macau or Taiwan cards.
type RegionCardInfo struct {
	Region string
	Gender string
	Valid  bool
}

// Convert15To18 converts a 15-digit mainland China identity card number to 18 digits.
func Convert15To18(idCard string) (string, bool) {
	return Convert15To18WithOptions(idCard)
}

// Convert15To18WithOptions converts a 15-digit mainland China identity card number to 18 digits with options.
func Convert15To18WithOptions(idCard string, opts ...IDCardOption) (string, bool) {
	cfg := applyIDCardOptions(opts)
	if len(idCard) != chinaIDMinLength || !cfg.digits(idCard) {
		return "", false
	}
	id18 := idCard[:6] + "19" + idCard[6:]
	return id18 + string(CheckCode18(id18[:17])), true
}

// Convert18To15 converts a valid 18-digit mainland China identity card number to 15 digits.
func Convert18To15(idCard string) (string, bool) {
	return Convert18To15WithOptions(idCard)
}

// Convert18To15WithOptions converts a valid 18-digit mainland China identity card number to 15 digits with options.
func Convert18To15WithOptions(idCard string, opts ...IDCardOption) (string, bool) {
	if !IsValidIDCard18WithOptions(idCard, opts...) {
		return "", false
	}
	return idCard[:6] + idCard[8:17], true
}

// IsValidIDCard reports whether idCard is a valid 18-digit, 15-digit, or Hong Kong/Macau/Taiwan card number.
func IsValidIDCard(idCard string) bool {
	return IsValidIDCardWithOptions(idCard)
}

// IsValidIDCardWithOptions reports whether idCard is valid with options.
func IsValidIDCardWithOptions(idCard string, opts ...IDCardOption) bool {
	if strings.TrimSpace(idCard) == "" {
		return false
	}
	switch len(idCard) {
	case chinaIDMaxLength:
		return IsValidIDCard18WithOptions(idCard, opts...)
	case chinaIDMinLength:
		return IsValidIDCard15WithOptions(idCard, opts...)
	case 10:
		info, ok := ParseRegionCardWithOptions(idCard, opts...)
		return ok && info.Valid
	default:
		return false
	}
}

// IsValidIDCard18 reports whether idCard is a valid 18-digit mainland China identity card number.
func IsValidIDCard18(idCard string) bool { return IsValidIDCard18WithOptions(idCard) }

// IsValidIDCard18WithOptions reports whether idCard is a valid 18-digit mainland China identity card number with options.
func IsValidIDCard18WithOptions(idCard string, opts ...IDCardOption) bool {
	return isValidIDCard18(idCard, true, applyIDCardOptions(opts))
}

// IsValidIDCard18WithIgnoreCase validates an 18-digit identity card number and controls X/x comparison.
func IsValidIDCard18WithIgnoreCase(idCard string, ignoreCase bool) bool {
	return IsValidIDCard18WithIgnoreCaseAndOptions(idCard, ignoreCase)
}

// IsValidIDCard18WithIgnoreCaseAndOptions validates an 18-digit identity card number with options.
func IsValidIDCard18WithIgnoreCaseAndOptions(idCard string, ignoreCase bool, opts ...IDCardOption) bool {
	return isValidIDCard18(idCard, ignoreCase, applyIDCardOptions(opts))
}

func isValidIDCard18(idCard string, ignoreCase bool, cfg idCardConfig) bool {
	if len(idCard) != chinaIDMaxLength {
		return false
	}
	provinceCode := idCard[:2]
	if strings.HasPrefix(idCard, "9") {
		provinceCode = idCard[1:3]
	}
	if _, ok := cityCodes[provinceCode]; !ok {
		return false
	}
	if !IsValidBirthday(idCard[6:14]) {
		return false
	}
	code17 := idCard[:17]
	if !cfg.digits(code17) {
		return false
	}
	check := CheckCode18(code17)
	actual := idCard[17]
	if ignoreCase {
		return strings.EqualFold(string(check), string(actual))
	}
	return check == actual
}

// IsValidIDCard15 reports whether idCard is a valid 15-digit mainland China identity card number.
func IsValidIDCard15(idCard string) bool {
	return IsValidIDCard15WithOptions(idCard)
}

// IsValidIDCard15WithOptions reports whether idCard is a valid 15-digit mainland China identity card number with options.
func IsValidIDCard15WithOptions(idCard string, opts ...IDCardOption) bool {
	cfg := applyIDCardOptions(opts)
	if len(idCard) != chinaIDMinLength || !cfg.digits(idCard) {
		return false
	}
	if _, ok := cityCodes[idCard[:2]]; !ok {
		return false
	}
	return IsValidBirthday("19" + idCard[6:12])
}

// ParseRegionCard validates a Hong Kong, Macau or Taiwan identity card number.
func ParseRegionCard(idCard string) (RegionCardInfo, bool) {
	return ParseRegionCardWithOptions(idCard)
}

// ParseRegionCardWithOptions validates a Hong Kong, Macau or Taiwan identity card number with options.
func ParseRegionCardWithOptions(idCard string, opts ...IDCardOption) (RegionCardInfo, bool) {
	if strings.TrimSpace(idCard) == "" {
		return RegionCardInfo{}, false
	}
	cfg := applyIDCardOptions(opts)
	idCard = strings.ReplaceAll(idCard, "（", "(")
	idCard = strings.ReplaceAll(idCard, "）", ")")
	card := strings.NewReplacer("(", "", ")", "").Replace(idCard)
	if len(card) != 8 && len(card) != 9 && len(idCard) != 10 {
		return RegionCardInfo{}, false
	}
	if cfg.tw(idCard) {
		info := RegionCardInfo{Region: "台湾", Gender: "N"}
		switch idCard[1] {
		case '1':
			info.Gender = "M"
		case '2':
			info.Gender = "F"
		default:
			info.Valid = false
			return info, true
		}
		info.Valid = IsValidTWIDCardWithOptions(idCard, opts...)
		return info, true
	}
	if cfg.macau(idCard) {
		return RegionCardInfo{Region: "澳门", Gender: "N", Valid: true}, true
	}
	if cfg.hk(idCard) {
		return RegionCardInfo{Region: "香港", Gender: "N", Valid: IsValidHKIDCardWithOptions(idCard, opts...)}, true
	}
	return RegionCardInfo{}, false
}

// IsValidTWIDCard reports whether idCard is a valid Taiwan identity card number.
func IsValidTWIDCard(idCard string) bool {
	return IsValidTWIDCardWithOptions(idCard)
}

// IsValidTWIDCardWithOptions reports whether idCard is a valid Taiwan identity card number with options.
func IsValidTWIDCardWithOptions(idCard string, opts ...IDCardOption) bool {
	cfg := applyIDCardOptions(opts)
	if !cfg.tw(idCard) {
		return false
	}
	if len(idCard) != 10 {
		return false
	}
	start, ok := twFirstCode[idCard[0]]
	if !ok {
		return false
	}
	sum := start/10 + (start%10)*9
	weight := 8
	for i := 1; i < 9; i++ {
		if idCard[i] < '0' || idCard[i] > '9' {
			return false
		}
		sum += int(idCard[i]-'0') * weight
		weight--
	}
	if idCard[9] < '0' || idCard[9] > '9' {
		return false
	}
	check := 0
	if mod := sum % 10; mod != 0 {
		check = 10 - mod
	}
	return check == int(idCard[9]-'0')
}

// IsValidHKIDCard reports whether idCard is a valid Hong Kong identity card number.
func IsValidHKIDCard(idCard string) bool {
	return IsValidHKIDCardWithOptions(idCard)
}

// IsValidHKIDCardWithOptions reports whether idCard is a valid Hong Kong identity card number with options.
func IsValidHKIDCardWithOptions(idCard string, opts ...IDCardOption) bool {
	cfg := applyIDCardOptions(opts)
	if !cfg.hk(idCard) {
		return false
	}
	card := strings.NewReplacer("(", "", ")", "").Replace(idCard)
	sum := 0
	if len(card) == 9 {
		sum = int(card[0]-55)*9 + int(card[1]-55)*8
		card = card[1:]
	} else {
		sum = 522 + int(card[0]-55)*8
	}
	weight := 7
	for i := 1; i < 7; i++ {
		if card[i] < '0' || card[i] > '9' {
			return false
		}
		sum += int(card[i]-'0') * weight
		weight--
	}
	end := card[7]
	switch {
	case end == 'A' || end == 'a':
		sum += 10
	case end >= '0' && end <= '9':
		sum += int(end - '0')
	default:
		return false
	}
	return sum%11 == 0
}

// BirthString returns the birthday encoded in idCard as yyyyMMdd.
func BirthString(idCard string) (string, bool) {
	if len(idCard) < chinaIDMinLength {
		return "", false
	}
	if len(idCard) == chinaIDMinLength {
		converted, ok := Convert15To18(idCard)
		if !ok {
			return "", false
		}
		idCard = converted
	}
	if len(idCard) < chinaIDMaxLength {
		return "", false
	}
	birth := idCard[6:14]
	return birth, IsValidBirthday(birth)
}

// BirthDate returns the birthday encoded in idCard.
func BirthDate(idCard string) (time.Time, bool) {
	return BirthDateWithOptions(idCard)
}

// BirthDateWithOptions returns the birthday encoded in idCard using custom parsing options.
func BirthDateWithOptions(idCard string, opts ...BirthOption) (time.Time, bool) {
	birth, ok := BirthString(idCard)
	if !ok {
		return time.Time{}, false
	}
	cfg := applyBirthOptions(opts)
	t, err := time.ParseInLocation("20060102", birth, cfg.location)
	return t, err == nil
}

// Age returns the current age encoded in idCard.
func Age(idCard string) (int, bool) { return AgeWithOptions(idCard) }

// AgeWithOptions returns the age encoded in idCard using custom time options.
func AgeWithOptions(idCard string, opts ...AgeOption) (int, bool) {
	cfg := applyAgeOptions(opts)
	return AgeAt(idCard, cfg.clock())
}

// AgeAt returns the age encoded in idCard at the specified time.
func AgeAt(idCard string, at time.Time) (int, bool) {
	birth, ok := BirthDate(idCard)
	if !ok {
		return 0, false
	}
	age := at.Year() - birth.Year()
	anniversary := time.Date(at.Year(), birth.Month(), birth.Day(), 0, 0, 0, 0, at.Location())
	if time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location()).Before(anniversary) {
		age--
	}
	return age, true
}

// Year returns the birth year encoded in idCard.
func Year(idCard string) (int, bool) {
	birth, ok := BirthString(idCard)
	if !ok {
		return 0, false
	}
	return atoi4(birth[0:4]), true
}

// Month returns the birth month encoded in idCard.
func Month(idCard string) (int, bool) {
	birth, ok := BirthString(idCard)
	if !ok {
		return 0, false
	}
	return int((birth[4]-'0')*10 + birth[5] - '0'), true
}

// Day returns the birth day encoded in idCard.
func Day(idCard string) (int, bool) {
	birth, ok := BirthString(idCard)
	if !ok {
		return 0, false
	}
	return int((birth[6]-'0')*10 + birth[7] - '0'), true
}

// GenderOf returns the gender encoded in a 15- or 18-digit identity card number.
func GenderOf(idCard string) (Gender, bool) {
	if len(idCard) != chinaIDMinLength && len(idCard) != chinaIDMaxLength {
		return GenderUnknown, false
	}
	if len(idCard) == chinaIDMinLength {
		converted, ok := Convert15To18(idCard)
		if !ok {
			return GenderUnknown, false
		}
		idCard = converted
	}
	if idCard[16] < '0' || idCard[16] > '9' {
		return GenderUnknown, false
	}
	if (idCard[16]-'0')%2 != 0 {
		return GenderMale, true
	}
	return GenderFemale, true
}

// ProvinceCode returns the province code encoded in a 15- or 18-digit identity card number.
func ProvinceCode(idCard string) (string, bool) {
	if len(idCard) == chinaIDMinLength || len(idCard) == chinaIDMaxLength {
		return idCard[:2], true
	}
	return "", false
}

// Province returns the province name encoded in a 15- or 18-digit identity card number.
func Province(idCard string) (string, bool) {
	code, ok := ProvinceCode(idCard)
	if !ok {
		return "", false
	}
	name, ok := cityCodes[code]
	return name, ok
}

// CityCode returns the city-level code encoded in a 15- or 18-digit identity card number.
func CityCode(idCard string) (string, bool) {
	if len(idCard) == chinaIDMinLength || len(idCard) == chinaIDMaxLength {
		return idCard[:4], true
	}
	return "", false
}

// DistrictCode returns the district-level code encoded in a 15- or 18-digit identity card number.
func DistrictCode(idCard string) (string, bool) {
	if len(idCard) == chinaIDMinLength || len(idCard) == chinaIDMaxLength {
		return idCard[:6], true
	}
	return "", false
}

// ParseIDCard parses a valid 15- or 18-digit mainland China identity card number.
func ParseIDCard(idCard string) (IDCardInfo, bool) {
	if !IsValidIDCard18(idCard) && !IsValidIDCard15(idCard) {
		return IDCardInfo{}, false
	}
	provinceCode, _ := ProvinceCode(idCard)
	province, _ := Province(idCard)
	cityCode, _ := CityCode(idCard)
	districtCode, _ := DistrictCode(idCard)
	birth, _ := BirthDate(idCard)
	gender, _ := GenderOf(idCard)
	age, _ := Age(idCard)
	return IDCardInfo{
		ProvinceCode: provinceCode,
		Province:     province,
		CityCode:     cityCode,
		DistrictCode: districtCode,
		Birth:        birth,
		Gender:       gender,
		Age:          age,
	}, true
}

// Hide replaces runes in [start, end) with '*'. Indexes are rune based.
func Hide(idCard string, start, end int) string {
	runes := []rune(idCard)
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start >= end {
		return idCard
	}
	for i := start; i < end; i++ {
		runes[i] = '*'
	}
	return string(runes)
}

// CheckCode18 returns the 18th check code for a 17-digit identity card body.
func CheckCode18(code17 string) byte {
	if len(code17) != len(idCardPower) || !rxDigits.MatchString(code17) {
		return ' '
	}
	sum := 0
	for i := 0; i < len(code17); i++ {
		sum += int(code17[i]-'0') * idCardPower[i]
	}
	return "10X98765432"[sum%11]
}

// IsValidBirthday reports whether s is a valid yyyyMMdd date.
func IsValidBirthday(s string) bool {
	return IsValidBirthdayWithOptions(s)
}

// IsValidBirthdayWithOptions reports whether s is a valid yyyyMMdd date using custom parsing options.
func IsValidBirthdayWithOptions(s string, opts ...BirthOption) bool {
	if len(s) != 8 || !rxDigits.MatchString(s) {
		return false
	}
	cfg := applyBirthOptions(opts)
	t, err := time.ParseInLocation("20060102", s, cfg.location)
	if err != nil {
		return false
	}
	return t.Format("20060102") == s
}

func atoi4(s string) int {
	return int(s[0]-'0')*1000 + int(s[1]-'0')*100 + int(s[2]-'0')*10 + int(s[3]-'0')
}
