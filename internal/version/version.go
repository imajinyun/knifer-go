package version

import (
	"fmt"
	"strings"
)

const DefaultVersionsDelimiter = ";"

// CompareVersion compares two version strings.
//
// It returns a negative value when version1 is smaller, a positive value when
// version1 is greater, and zero when both versions are equivalent.
func CompareVersion(version1, version2 string) int {
	return compareParsed(parseVersion(version1), parseVersion(version2))
}

// AnyMatch reports whether currentVersion matches any expression in compareVersions.
func AnyMatch(currentVersion string, compareVersions ...string) bool {
	return MatchEl(currentVersion, strings.Join(compareVersions, DefaultVersionsDelimiter))
}

// AnyMatchSlice reports whether currentVersion matches any expression in compareVersions.
func AnyMatchSlice(currentVersion string, compareVersions []string) bool {
	return AnyMatch(currentVersion, compareVersions...)
}

// IsGreaterThan reports whether currentVersion is greater than compareVersion.
func IsGreaterThan(currentVersion, compareVersion string) bool {
	return MatchEl(currentVersion, ">"+compareVersion)
}

// IsGreaterThanOrEqual reports whether currentVersion is greater than or equal to compareVersion.
func IsGreaterThanOrEqual(currentVersion, compareVersion string) bool {
	return MatchEl(currentVersion, ">="+compareVersion)
}

// IsLessThan reports whether currentVersion is less than compareVersion.
func IsLessThan(currentVersion, compareVersion string) bool {
	return MatchEl(currentVersion, "<"+compareVersion)
}

// IsLessThanOrEqual reports whether currentVersion is less than or equal to compareVersion.
func IsLessThanOrEqual(currentVersion, compareVersion string) bool {
	return MatchEl(currentVersion, "<="+compareVersion)
}

// MatchEl reports whether currentVersion satisfies a semicolon-separated version expression.
func MatchEl(currentVersion, versionEl string) bool {
	return MatchElWithDelimiter(currentVersion, versionEl, DefaultVersionsDelimiter)
}

// MatchElWithDelimiter reports whether currentVersion satisfies versionEl using versionsDelimiter.
//
// The expression may contain multiple alternatives separated by versionsDelimiter.
// Each alternative may be an exact version, a comparison expression such as
// ">=1.2.3", or an inclusive range such as "1.0.0-1.5.0". Open ranges are
// supported: "-1.5.0" means <=1.5.0 and "1.0.0-" means >=1.0.0.
func MatchElWithDelimiter(currentVersion, versionEl, versionsDelimiter string) bool {
	return MatchElWithDelimiterErr(currentVersion, versionEl, versionsDelimiter) == nil
}

// MatchElWithDelimiterErr validates the delimiter and reports expression matching errors.
func MatchElWithDelimiterErr(currentVersion, versionEl, versionsDelimiter string) error {
	if err := validateDelimiter(versionsDelimiter); err != nil {
		return err
	}
	if strings.TrimSpace(versionEl) == "" {
		return errNoMatch
	}
	trimmedVersion := strings.TrimSpace(currentVersion)
	parts := splitExpression(versionEl, versionsDelimiter)
	if len(parts) == 0 {
		return errNoMatch
	}
	for _, el := range parts {
		if matchOne(trimmedVersion, el) {
			return nil
		}
	}
	return errNoMatch
}

// MatchElByDelimiter is a bool-returning convenience wrapper around MatchElWithDelimiter.
func MatchElByDelimiter(currentVersion, versionEl, versionsDelimiter string) bool {
	return MatchElWithDelimiter(currentVersion, versionEl, versionsDelimiter)
}

var errNoMatch = fmt.Errorf("version: no expression matched")

type parsedVersion struct {
	sequence []token
	pre      []token
	build    []token
}

type token struct {
	number bool
	num    string
	text   string
}

func validateDelimiter(delimiter string) error {
	if strings.TrimSpace(delimiter) == "" || delimiter == "-" || compareOpPrefix(delimiter) != "" {
		return fmt.Errorf("version: invalid delimiter %q", delimiter)
	}
	return nil
}

func splitExpression(expr, delimiter string) []string {
	raw := strings.Split(expr, delimiter)
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func matchOne(currentVersion, expression string) bool {
	if op := compareOpPrefix(expression); op != "" {
		ver := strings.TrimSpace(strings.TrimPrefix(expression, op))
		var target *string
		if !strings.EqualFold(ver, "null") {
			target = &ver
		}
		cmp := compareNullable(&currentVersion, target)
		switch op {
		case ">=", "≥", "≥=":
			return cmp >= 0
		case "<=", "≤", "≤=":
			return cmp <= 0
		case ">":
			return cmp > 0
		case "<":
			return cmp < 0
		default:
			return false
		}
	}
	if strings.Contains(expression, "-") {
		idx := strings.Index(expression, "-")
		left := strings.TrimSpace(expression[:idx])
		right := strings.TrimSpace(expression[idx+1:])
		leftMatch := left == "" || CompareVersion(left, currentVersion) <= 0
		rightMatch := right == "" || CompareVersion(right, currentVersion) >= 0
		return leftMatch && rightMatch
	}
	return currentVersion == expression
}

func compareNullable(version1, version2 *string) int {
	if version1 == nil && version2 == nil {
		return 0
	}
	if version1 == nil {
		return -1
	}
	if version2 == nil {
		return 1
	}
	return CompareVersion(*version1, *version2)
}

func compareOpPrefix(s string) string {
	if s == "" {
		return ""
	}
	first := []rune(s)[0]
	if first != '>' && first != '<' && first != '≥' && first != '≤' {
		return ""
	}
	if len([]rune(s)) > 1 {
		runes := []rune(s)
		if runes[1] == '=' {
			return string(runes[:2])
		}
	}
	return string(first)
}

func parseVersion(v string) parsedVersion {
	n := len(v)
	p := parsedVersion{
		sequence: make([]token, 0, 4),
		pre:      make([]token, 0, 2),
		build:    make([]token, 0, 2),
	}
	if n == 0 {
		return p
	}
	i := 0
	c := v[i]
	i = takeInitialToken(v, i, &p.sequence)
	for i < n {
		c = v[i]
		if c == '.' {
			i++
			continue
		}
		if c == '-' || c == '+' {
			i++
			break
		}
		if isASCIIDigit(c) {
			i = takeNumber(v, i, &p.sequence)
		} else {
			i = takeString(v, i, &p.sequence)
		}
	}
	if c == '-' && i >= n {
		return p
	}
	for i < n {
		c = v[i]
		if isASCIIDigit(c) {
			i = takeNumber(v, i, &p.pre)
		} else {
			i = takeString(v, i, &p.pre)
		}
		if i >= n {
			break
		}
		c = v[i]
		if c == '.' || c == '-' {
			i++
			continue
		}
		if c == '+' {
			i++
			break
		}
	}
	if c == '+' && i >= n {
		return p
	}
	for i < n {
		c = v[i]
		if isASCIIDigit(c) {
			i = takeNumber(v, i, &p.build)
		} else {
			i = takeString(v, i, &p.build)
		}
		if i >= n {
			break
		}
		c = v[i]
		if c == '.' || c == '-' || c == '+' {
			i++
		}
	}
	return p
}

func takeInitialToken(s string, i int, acc *[]token) int {
	if isASCIIDigit(s[i]) {
		return takeNumber(s, i, acc)
	}
	// Preserve the original comparison rule for leading non-digit ASCII tokens:
	// the first byte is treated as its byte value offset by '0'.
	*acc = append(*acc, token{number: true, num: normalizeNumber(fmt.Sprint(int(s[i] - '0'))), text: fmt.Sprint(int(s[i] - '0'))})
	return i + 1
}

func takeNumber(s string, i int, acc *[]token) int {
	start := i
	for i < len(s) && isASCIIDigit(s[i]) {
		i++
	}
	raw := s[start:i]
	*acc = append(*acc, token{number: true, num: normalizeNumber(raw), text: normalizeNumber(raw)})
	return i
}

func takeString(s string, i int, acc *[]token) int {
	start := i
	for i < len(s) {
		c := s[i]
		if c == '.' || c == '-' || c == '+' || isASCIIDigit(c) {
			break
		}
		i++
	}
	*acc = append(*acc, token{text: s[start:i]})
	return i
}

func normalizeNumber(s string) string {
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = strings.TrimPrefix(s, "-")
	}
	s = strings.TrimLeft(s, "0")
	if s == "" {
		s = "0"
	}
	if neg && s != "0" {
		return "-" + s
	}
	return s
}

func compareParsed(v1, v2 parsedVersion) int {
	if c := compareTokens(v1.sequence, v2.sequence); c != 0 {
		return c
	}
	if len(v1.pre) == 0 && len(v2.pre) != 0 {
		return 1
	}
	if len(v1.pre) != 0 && len(v2.pre) == 0 {
		return -1
	}
	if c := compareTokens(v1.pre, v2.pre); c != 0 {
		return c
	}
	return compareTokens(v1.build, v2.build)
}

func compareTokens(ts1, ts2 []token) int {
	n := min(len(ts1), len(ts2))
	for i := 0; i < n; i++ {
		o1, o2 := ts1[i], ts2[i]
		var c int
		switch {
		case o1.number && o2.number:
			c = compareNumberString(o1.num, o2.num)
		case !o1.number && !o2.number, o1.number != o2.number:
			c = strings.Compare(o1.text, o2.text)
		}
		if c != 0 {
			return c
		}
	}
	rest := ts1
	if len(ts2) > len(ts1) {
		rest = ts2
	}
	for i := n; i < len(rest); i++ {
		if rest[i].number && rest[i].num == "0" {
			continue
		}
		return len(ts1) - len(ts2)
	}
	return 0
}

func compareNumberString(a, b string) int {
	aNeg, bNeg := strings.HasPrefix(a, "-"), strings.HasPrefix(b, "-")
	if aNeg != bNeg {
		if aNeg {
			return -1
		}
		return 1
	}
	aa, bb := strings.TrimPrefix(a, "-"), strings.TrimPrefix(b, "-")
	var c int
	switch {
	case len(aa) < len(bb):
		c = -1
	case len(aa) > len(bb):
		c = 1
	default:
		c = strings.Compare(aa, bb)
	}
	if aNeg {
		return -c
	}
	return c
}

func isASCIIDigit(b byte) bool { return b >= '0' && b <= '9' }
