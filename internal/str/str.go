package str

import (
	"fmt"
	"hash/fnv"
	"math/bits"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

// This file provides string helpers aligned with the utility toolkit-core StrUtil and CharSequenceUtil.

var emojiPattern = regexp.MustCompile(`(?:[\x{1F1E6}-\x{1F1FF}]{2}|[#*0-9]\x{FE0F}?\x{20E3}|[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}])(?:\x{FE0F}|\x{200D}[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]\x{FE0F}?)*`)

type emojiConfig struct {
	matcher  func(string) bool
	replacer func(string) string
}

// EmojiOption customizes emoji helpers per call.
type EmojiOption func(*emojiConfig)

// WithEmojiMatcher sets the matcher used by ContainsEmojiWithOptions.
func WithEmojiMatcher(matcher func(string) bool) EmojiOption {
	return func(c *emojiConfig) {
		if matcher != nil {
			c.matcher = matcher
		}
	}
}

// WithEmojiReplacer sets the replacer used by RemoveEmojiWithOptions.
func WithEmojiReplacer(replacer func(string) string) EmojiOption {
	return func(c *emojiConfig) {
		if replacer != nil {
			c.replacer = replacer
		}
	}
}

func applyEmojiOptions(opts []EmojiOption) emojiConfig {
	cfg := emojiConfig{matcher: emojiPattern.MatchString, replacer: func(s string) string { return emojiPattern.ReplaceAllString(s, "") }}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.matcher == nil {
		cfg.matcher = emojiPattern.MatchString
	}
	if cfg.replacer == nil {
		cfg.replacer = func(s string) string { return emojiPattern.ReplaceAllString(s, "") }
	}
	return cfg
}

// IsEmpty reports whether s has zero length.
func IsEmpty(s string) bool { return len(s) == 0 }

// IsNotEmpty reports whether s is not empty.
func IsNotEmpty(s string) bool { return !IsEmpty(s) }

// IsBlank reports whether s is empty or contains only Unicode white space.
func IsBlank(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// IsNotBlank reports whether s is not blank.
func IsNotBlank(s string) bool { return !IsBlank(s) }

// HasEmpty reports whether any string is empty.
func HasEmpty(strs ...string) bool {
	for _, s := range strs {
		if IsEmpty(s) {
			return true
		}
	}
	return false
}

// HasBlank reports whether any string is blank.
func HasBlank(strs ...string) bool {
	for _, s := range strs {
		if IsBlank(s) {
			return true
		}
	}
	return false
}

// IsAllEmpty reports whether all strings are empty.
func IsAllEmpty(strs ...string) bool {
	for _, s := range strs {
		if IsNotEmpty(s) {
			return false
		}
	}
	return true
}

// IsAllBlank reports whether all strings are blank.
func IsAllBlank(strs ...string) bool {
	for _, s := range strs {
		if IsNotBlank(s) {
			return false
		}
	}
	return true
}

// Trim removes leading and trailing white space.
func Trim(s string) string { return strings.TrimSpace(s) }

// TrimToEmpty removes leading and trailing white space.
func TrimToEmpty(s string) string { return strings.TrimSpace(s) }

// TrimStart removes leading white space.
func TrimStart(s string) string { return strings.TrimLeftFunc(s, unicode.IsSpace) }

// TrimEnd removes trailing white space.
func TrimEnd(s string) string { return strings.TrimRightFunc(s, unicode.IsSpace) }

// Sub returns a substring by rune indexes and supports negative indexes from the end.
// fromIndex is inclusive and toIndex is exclusive; reversed ranges are normalized.
func Sub(s string, fromIndex, toIndex int) string {
	rs := []rune(s)
	n := len(rs)
	if n == 0 {
		return ""
	}
	if fromIndex < 0 {
		fromIndex += n
	}
	if toIndex < 0 {
		toIndex += n
	}
	if fromIndex < 0 {
		fromIndex = 0
	}
	if toIndex > n {
		toIndex = n
	}
	if fromIndex > toIndex {
		fromIndex, toIndex = toIndex, fromIndex
	}
	if fromIndex == toIndex {
		return ""
	}
	return string(rs[fromIndex:toIndex])
}

// SubBefore returns the text before sep. When isLastSeparator is true, the last sep is used.
func SubBefore(s, sep string, isLastSeparator bool) string {
	if s == "" || sep == "" {
		return s
	}
	var idx int
	if isLastSeparator {
		idx = strings.LastIndex(s, sep)
	} else {
		idx = strings.Index(s, sep)
	}
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// SubAfter returns the text after sep. When isLastSeparator is true, the last sep is used.
func SubAfter(s, sep string, isLastSeparator bool) string {
	if s == "" {
		return s
	}
	if sep == "" {
		return ""
	}
	var idx int
	if isLastSeparator {
		idx = strings.LastIndex(s, sep)
	} else {
		idx = strings.Index(s, sep)
	}
	if idx == -1 {
		return ""
	}
	return s[idx+len(sep):]
}

// Split splits s by sep and returns an empty slice for an empty input string.
func Split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// SplitTrim splits s, trims each part, and drops blank parts.
func SplitTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Repeat repeats s n times. Non-positive n returns an empty string.
func Repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(s, n)
}

// PadLeft pads s on the left to the requested rune length.
func PadLeft(s string, length int, pad rune) string {
	rs := []rune(s)
	if len(rs) >= length {
		return s
	}
	padding := make([]rune, length-len(rs))
	for i := range padding {
		padding[i] = pad
	}
	return string(padding) + s
}

// PadRight pads s on the right to the requested rune length.
func PadRight(s string, length int, pad rune) string {
	rs := []rune(s)
	if len(rs) >= length {
		return s
	}
	padding := make([]rune, length-len(rs))
	for i := range padding {
		padding[i] = pad
	}
	return s + string(padding)
}

// Contains reports whether s contains sub.
func Contains(s, sub string) bool { return strings.Contains(s, sub) }

// ContainsEmoji reports whether s contains emoji-like runes.
func ContainsEmoji(s string) bool { return ContainsEmojiWithOptions(s) }

// ContainsEmojiWithOptions reports whether s contains emoji-like runes with options.
func ContainsEmojiWithOptions(s string, opts ...EmojiOption) bool {
	return applyEmojiOptions(opts).matcher(s)
}

// RemoveEmoji removes emoji-like runes from s, including variation-selector and
// zero-width-joiner based emoji sequences.
func RemoveEmoji(s string) string { return RemoveEmojiWithOptions(s) }

// RemoveEmojiWithOptions removes emoji-like runes from s with options.
func RemoveEmojiWithOptions(s string, opts ...EmojiOption) string {
	return applyEmojiOptions(opts).replacer(s)
}

// ContainsAny reports whether s contains any candidate substring.
func ContainsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ContainsAll reports whether s contains all candidate substrings.
func ContainsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

// ContainsIgnoreCase reports whether s contains sub case-insensitively.
func ContainsIgnoreCase(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

// StartsWith reports whether s starts with prefix.
func StartsWith(s, prefix string) bool { return strings.HasPrefix(s, prefix) }

// EndsWith reports whether s ends with suffix.
func EndsWith(s, suffix string) bool { return strings.HasSuffix(s, suffix) }

// EqualsIgnoreCase compares strings case-insensitively.
func EqualsIgnoreCase(a, b string) bool { return strings.EqualFold(a, b) }

// Reverse reverses a string by rune, preserving multi-byte characters.
func Reverse(s string) string {
	rs := []rune(s)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	return string(rs)
}

// Format mimics the utility toolkit StrUtil.format by replacing {} placeholders in order.
//
//	Format("name={}, age={}", "tom", 12) -> "name=tom, age=12"
//
// Use \\{ to escape a literal opening brace.
func Format(template string, args ...any) string {
	if template == "" || len(args) == 0 {
		return template
	}
	var b strings.Builder
	b.Grow(len(template))
	idx := 0
	for i := 0; i < len(template); i++ {
		c := template[i]
		if c == '\\' && i+1 < len(template) && template[i+1] == '{' {
			b.WriteByte('{')
			i++
			continue
		}
		if c == '{' && i+1 < len(template) && template[i+1] == '}' {
			if idx < len(args) {
				fmt.Fprint(&b, args[idx])
				idx++
			} else {
				b.WriteString("{}")
			}
			i++
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// RemovePrefix removes prefix when present.
func RemovePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// RemoveSuffix removes suffix when present.
func RemoveSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// AddPrefixIfNot adds prefix when it is not already present.
func AddPrefixIfNot(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s
	}
	return prefix + s
}

// AddSuffixIfNot adds suffix when it is not already present.
func AddSuffixIfNot(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}

// Length returns the number of runes in s.
func Length(s string) int { return len([]rune(s)) }

// RuneLen returns the number of runes in s.
func RuneLen(s string) int { return Length(s) }

// EscapeUnicode escapes non-ASCII runes as Java-style \uXXXX sequences.
func EscapeUnicode(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 128 {
			b.WriteRune(r)
			continue
		}
		if r <= 0xFFFF {
			writeUnicodeEscape(&b, r)
			continue
		}
		for _, surrogate := range utf16.Encode([]rune{r}) {
			writeUnicodeEscape(&b, rune(surrogate))
		}
	}
	return b.String()
}

// UnescapeUnicode decodes Java-style \uXXXX sequences.
//
// Malformed escapes are preserved verbatim. Surrogate pairs are combined when
// both halves are present; lone surrogates are emitted as their rune value.
func UnescapeUnicode(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if i+6 > len(s) || s[i] != '\\' || s[i+1] != 'u' {
			b.WriteByte(s[i])
			i++
			continue
		}
		hi, ok := parseUnicodeHex(s[i+2 : i+6])
		if !ok {
			b.WriteByte(s[i])
			i++
			continue
		}
		i += 6
		if utf16.IsSurrogate(hi) && i+6 <= len(s) && s[i] == '\\' && s[i+1] == 'u' {
			lo, ok := parseUnicodeHex(s[i+2 : i+6])
			if ok {
				if decoded := utf16.DecodeRune(hi, lo); decoded != unicode.ReplacementChar {
					b.WriteRune(decoded)
					i += 6
					continue
				}
			}
		}
		b.WriteRune(hi)
	}
	return b.String()
}

// AntPathMatch reports whether path matches an Ant-style pattern using "/" as separator.
//
// Within a path segment, "*" matches any characters and "?" matches one rune.
// A segment that is exactly "**" matches zero or more path segments.
func AntPathMatch(pattern, path string) bool {
	return AntPathMatchWithSeparator(pattern, path, "/")
}

// AntPathMatchWithSeparator reports whether path matches an Ant-style pattern.
func AntPathMatchWithSeparator(pattern, path, separator string) bool {
	if separator == "" {
		separator = "/"
	}
	if pattern == path {
		return true
	}
	patternSegments := splitPathSegments(pattern, separator)
	pathSegments := splitPathSegments(path, separator)
	return matchAntSegments(patternSegments, pathSegments, 0, 0)
}

// JaccardSimilarity returns the Jaccard similarity of two strings by rune set.
//
// Unicode whitespace is ignored. Two empty effective inputs are considered
// identical and return 1.0.
func JaccardSimilarity(a, b string) float64 {
	left := runeSet(a)
	right := runeSet(b)
	return jaccard(left, right)
}

// NGramSimilarity returns the Jaccard similarity of rune n-gram sets.
//
// Whitespace is ignored before n-grams are built. If n is not positive the
// function returns 0. Two empty effective inputs are considered identical.
func NGramSimilarity(a, b string, n int) float64 {
	if n <= 0 {
		return 0
	}
	left := ngramSet(a, n)
	right := ngramSet(b, n)
	return jaccard(left, right)
}

// LevenshteinDistance returns the Unicode-aware edit distance between a and b.
func LevenshteinDistance(a, b string) int {
	left := []rune(a)
	right := []rune(b)
	if len(left) == 0 {
		return len(right)
	}
	if len(right) == 0 {
		return len(left)
	}

	prev := make([]int, len(right)+1)
	curr := make([]int, len(right)+1)
	for j := range prev {
		prev[j] = j
	}
	for i, lr := range left {
		curr[0] = i + 1
		for j, rr := range right {
			cost := 0
			if lr != rr {
				cost = 1
			}
			curr[j+1] = min(curr[j]+1, prev[j+1]+1, prev[j]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(right)]
}

// LevenshteinSimilarity returns a normalized similarity score in [0, 1].
func LevenshteinSimilarity(a, b string) float64 {
	leftLen := len([]rune(a))
	rightLen := len([]rune(b))
	maxLen := max(leftLen, rightLen)
	if maxLen == 0 {
		return 1
	}
	distance := LevenshteinDistance(a, b)
	return 1 - float64(distance)/float64(maxLen)
}

// SimHash returns a deterministic 64-bit SimHash for text.
//
// Whitespace-separated lower-cased fields are used as tokens. If the text has no
// fields, non-space runes are used as fallback tokens. Empty input returns 0.
func SimHash(text string) uint64 {
	tokens := simHashTokens(text)
	if len(tokens) == 0 {
		return 0
	}

	var vector [64]int
	for _, token := range tokens {
		hash := hashString64(token)
		for i := range vector {
			if hash&(uint64(1)<<i) != 0 {
				vector[i]++
			} else {
				vector[i]--
			}
		}
	}

	var result uint64
	for i, weight := range vector {
		if weight > 0 {
			result |= uint64(1) << i
		}
	}
	return result
}

// HammingDistance64 returns the number of different bits between a and b.
func HammingDistance64(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

func writeUnicodeEscape(b *strings.Builder, r rune) {
	fmt.Fprintf(b, `\u%04X`, r)
}

func parseUnicodeHex(s string) (rune, bool) {
	v, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		return 0, false
	}
	return rune(v), true
}

func splitPathSegments(s, separator string) []string {
	if s == "" {
		return []string{}
	}
	trimmed := strings.Trim(s, separator)
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, separator)
}

func matchAntSegments(pattern, path []string, pi, si int) bool {
	if pi == len(pattern) {
		return si == len(path)
	}
	if pattern[pi] == "**" {
		for next := si; next <= len(path); next++ {
			if matchAntSegments(pattern, path, pi+1, next) {
				return true
			}
		}
		return false
	}
	if si == len(path) {
		return false
	}
	return matchAntSegment(pattern[pi], path[si]) && matchAntSegments(pattern, path, pi+1, si+1)
}

func matchAntSegment(pattern, text string) bool {
	pr := []rune(pattern)
	tr := []rune(text)
	dp := make([][]bool, len(pr)+1)
	for i := range dp {
		dp[i] = make([]bool, len(tr)+1)
	}
	dp[0][0] = true
	for i := 1; i <= len(pr); i++ {
		if pr[i-1] == '*' {
			dp[i][0] = dp[i-1][0]
		}
	}
	for i := 1; i <= len(pr); i++ {
		for j := 1; j <= len(tr); j++ {
			switch pr[i-1] {
			case '*':
				dp[i][j] = dp[i-1][j] || dp[i][j-1]
			case '?':
				dp[i][j] = dp[i-1][j-1]
			default:
				dp[i][j] = pr[i-1] == tr[j-1] && dp[i-1][j-1]
			}
		}
	}
	return dp[len(pr)][len(tr)]
}

func runeSet(s string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		set[string(r)] = struct{}{}
	}
	return set
}

func ngramSet(s string, n int) map[string]struct{} {
	runes := effectiveRunes(s)
	set := map[string]struct{}{}
	if len(runes) == 0 {
		return set
	}
	if len(runes) < n {
		set[string(runes)] = struct{}{}
		return set
	}
	for i := 0; i+n <= len(runes); i++ {
		set[string(runes[i:i+n])] = struct{}{}
	}
	return set
}

func effectiveRunes(s string) []rune {
	out := []rune{}
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func jaccard(left, right map[string]struct{}) float64 {
	if len(left) == 0 && len(right) == 0 {
		return 1
	}
	if len(left) == 0 || len(right) == 0 {
		return 0
	}

	intersection := 0
	for item := range left {
		if _, ok := right[item]; ok {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	return float64(intersection) / float64(union)
}

func simHashTokens(text string) []string {
	fields := strings.Fields(strings.ToLower(text))
	if len(fields) > 0 {
		return fields
	}

	tokens := []string{}
	for _, r := range strings.ToLower(text) {
		if unicode.IsSpace(r) {
			continue
		}
		tokens = append(tokens, string(r))
	}
	return tokens
}

func hashString64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
