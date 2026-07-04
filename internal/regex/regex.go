// Package regex provides regular-expression helpers.
package regex

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	// REChinese matches a single Chinese Han character.
	REChinese = `\p{Han}`
	// REChineses matches a non-empty string made only of Chinese Han characters.
	REChineses = `^\p{Han}+$`
)

var (
	reKeys         = map[rune]struct{}{'$': {}, '(': {}, ')': {}, '*': {}, '+': {}, '.': {}, '[': {}, ']': {}, '?': {}, '\\': {}, '^': {}, '{': {}, '}': {}, '|': {}}
	groupVarRegexp = regexp.MustCompile(`\$(\d+)`)
	numbersRegexp  = regexp.MustCompile(`\d+`)
	namedGroupRe   = regexp.MustCompile(`\(\?<([A-Za-z_][A-Za-z0-9_]*)>`)
)

type regexConfig struct {
	compile             func(string) (*regexp.Regexp, error)
	dotAll              bool
	groupVarRegexp      *regexp.Regexp
	numbersRegexp       *regexp.Regexp
	namedGroupRegexp    *regexp.Regexp
	normalizeNamedGroup func(string) string
}

// Option customizes pattern-string regex helpers per call.
type Option func(*regexConfig)

// WithCompileFunc sets the compiler used by pattern-string regex helpers.
func WithCompileFunc(compile func(string) (*regexp.Regexp, error)) Option {
	return func(c *regexConfig) {
		if compile != nil {
			c.compile = compile
		}
	}
}

// WithDotAll controls whether pattern-string helpers wrap patterns with (?s:...).
func WithDotAll(dotAll bool) Option { return func(c *regexConfig) { c.dotAll = dotAll } }

// WithGroupVarRegexp sets the regexp used by TemplateVarsWithOptions.
func WithGroupVarRegexp(re *regexp.Regexp) Option {
	return func(c *regexConfig) {
		if re != nil {
			c.groupVarRegexp = re
		}
	}
}

// WithNumbersRegexp sets the regexp used by GetFirstNumberWithOptions.
func WithNumbersRegexp(re *regexp.Regexp) Option {
	return func(c *regexConfig) {
		if re != nil {
			c.numbersRegexp = re
		}
	}
}

// WithNamedGroupRegexp sets the regexp used to normalize (?<name>...) groups before compiling.
func WithNamedGroupRegexp(re *regexp.Regexp) Option {
	return func(c *regexConfig) { c.namedGroupRegexp = re }
}

// WithNamedGroupNormalizer sets the normalizer used before compiling pattern strings.
func WithNamedGroupNormalizer(normalize func(string) string) Option {
	return func(c *regexConfig) { c.normalizeNamedGroup = normalize }
}

func applyOptions(opts []Option) regexConfig {
	cfg := regexConfig{
		compile:          regexp.Compile,
		dotAll:           true,
		groupVarRegexp:   groupVarRegexp,
		numbersRegexp:    numbersRegexp,
		namedGroupRegexp: namedGroupRe,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.compile == nil {
		cfg.compile = regexp.Compile
	}
	if cfg.groupVarRegexp == nil {
		cfg.groupVarRegexp = groupVarRegexp
	}
	if cfg.numbersRegexp == nil {
		cfg.numbersRegexp = numbersRegexp
	}
	if cfg.namedGroupRegexp == nil {
		cfg.namedGroupRegexp = namedGroupRe
	}
	return cfg
}

// MatchResult describes a single regular-expression match.
// Start and End are byte offsets in the original string.
type MatchResult struct {
	Text       string
	Start      int
	End        int
	Groups     []string
	GroupNames map[string]string
}

// ReMatch reports whether s contains a match for pattern. Invalid patterns return false.
func ReMatch(pattern, s string) bool {
	return ReMatchWithOptions(pattern, s)
}

// ReMatchWithOptions reports whether s contains a match for pattern with options.
func ReMatchWithOptions(pattern, s string, opts ...Option) bool {
	re, err := compile(pattern, opts...)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// ReFind returns the first match, or an empty string when there is no match or the pattern is invalid.
func ReFind(pattern, s string) string {
	return ReFindWithOptions(pattern, s)
}

// ReFindWithOptions returns the first match with options.
func ReFindWithOptions(pattern, s string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return ""
	}
	return re.FindString(s)
}

// ReFindAll returns all whole-match results, or nil when the pattern is invalid.
func ReFindAll(pattern, s string) []string {
	return ReFindAllWithOptions(pattern, s)
}

// ReFindAllWithOptions returns all whole-match results with options.
func ReFindAllWithOptions(pattern, s string, opts ...Option) []string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return nil
	}
	return re.FindAllString(s, -1)
}

// ReReplace replaces matches of pattern with replacement. Invalid patterns return the original string.
func ReReplace(pattern, s, replacement string) string {
	return ReReplaceWithOptions(pattern, s, replacement)
}

// ReReplaceWithOptions replaces matches of pattern with options.
func ReReplaceWithOptions(pattern, s, replacement string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, replacement)
}

// GetGroup0 returns the full text of the first match.
func GetGroup0(pattern, content string) string { return GetGroup0WithOptions(pattern, content) }

// GetGroup0WithOptions returns the full text of the first match with options.
func GetGroup0WithOptions(pattern, content string, opts ...Option) string {
	return GetWithOptions(pattern, content, 0, opts...)
}

// GetGroup1 returns the first capture group of the first match.
func GetGroup1(pattern, content string) string { return GetGroup1WithOptions(pattern, content) }

// GetGroup1WithOptions returns the first capture group of the first match with options.
func GetGroup1WithOptions(pattern, content string, opts ...Option) string {
	return GetWithOptions(pattern, content, 1, opts...)
}

// Get returns a capture group from the first match. Missing matches or invalid patterns return an empty string.
func Get(pattern, content string, groupIndex int) string {
	return GetWithOptions(pattern, content, groupIndex)
}

// GetWithOptions returns a capture group from the first match with options.
func GetWithOptions(pattern, content string, groupIndex int, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return ""
	}
	return GetRe(re, content, groupIndex)
}

// GetOK returns a capture group from the first match and reports whether it exists.
func GetOK(pattern, content string, groupIndex int) (string, bool) {
	return GetOKWithOptions(pattern, content, groupIndex)
}

// GetOKWithOptions returns a capture group from the first match with options and reports whether it exists.
func GetOKWithOptions(pattern, content string, groupIndex int, opts ...Option) (string, bool) {
	re, err := compile(pattern, opts...)
	if err != nil {
		return "", false
	}
	return GetReOK(re, content, groupIndex)
}

// GetByName returns a named capture group from the first match.
func GetByName(pattern, content, groupName string) string {
	return GetByNameWithOptions(pattern, content, groupName)
}

// GetByNameWithOptions returns a named capture group from the first match with options.
func GetByNameWithOptions(pattern, content, groupName string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return ""
	}
	return GetByNameRe(re, content, groupName)
}

// GetRe returns a capture group from the first match of a compiled expression.
func GetRe(re *regexp.Regexp, content string, groupIndex int) string {
	value, _ := GetReOK(re, content, groupIndex)
	return value
}

// GetReOK returns a capture group from the first match of a compiled expression and reports whether it exists.
func GetReOK(re *regexp.Regexp, content string, groupIndex int) (string, bool) {
	if re == nil || groupIndex < 0 {
		return "", false
	}
	groups := re.FindStringSubmatch(content)
	if groupIndex >= len(groups) {
		return "", false
	}
	return groups[groupIndex], true
}

// GetByNameRe returns a named capture group from the first match of a compiled expression.
func GetByNameRe(re *regexp.Regexp, content, groupName string) string {
	if re == nil || groupName == "" {
		return ""
	}
	groups := re.FindStringSubmatch(content)
	if groups == nil {
		return ""
	}
	for i, name := range re.SubexpNames() {
		if i > 0 && name == groupName && i < len(groups) {
			return groups[i]
		}
	}
	return ""
}

// First calls consumer with the first match of re.
func First(re *regexp.Regexp, content string, consumer func(MatchResult)) {
	if re == nil || consumer == nil {
		return
	}
	if loc := re.FindStringSubmatchIndex(content); loc != nil {
		consumer(buildMatchResult(re, content, loc))
	}
}

// GetAllGroups returns capture groups from matches. Group 0 is included when withGroup0 is true.
func GetAllGroups(pattern, content string, withGroup0 bool, findAll bool) []string {
	return GetAllGroupsWithOptions(pattern, content, withGroup0, findAll)
}

// GetAllGroupsWithOptions returns capture groups from matches with options.
func GetAllGroupsWithOptions(pattern, content string, withGroup0 bool, findAll bool, opts ...Option) []string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return nil
	}
	return GetAllGroupsRe(re, content, withGroup0, findAll)
}

// GetAllGroupsRe returns capture groups from matches of a compiled expression.
func GetAllGroupsRe(re *regexp.Regexp, content string, withGroup0 bool, findAll bool) []string {
	if re == nil {
		return nil
	}
	all := re.FindAllStringSubmatch(content, allLimit(findAll))
	result := make([]string, 0)
	start := 1
	if withGroup0 {
		start = 0
	}
	for _, groups := range all {
		for i := start; i < len(groups); i++ {
			result = append(result, groups[i])
		}
	}
	return result
}

// GetAllGroupNames returns named capture groups from the first match.
func GetAllGroupNames(pattern, content string) map[string]string {
	return GetAllGroupNamesWithOptions(pattern, content)
}

// GetAllGroupNamesWithOptions returns named capture groups from the first match with options.
func GetAllGroupNamesWithOptions(pattern, content string, opts ...Option) map[string]string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return nil
	}
	return GetAllGroupNamesRe(re, content)
}

// GetAllGroupNamesRe returns named capture groups from the first match of a compiled expression.
func GetAllGroupNamesRe(re *regexp.Regexp, content string) map[string]string {
	if re == nil {
		return nil
	}
	result := map[string]string{}
	groups := re.FindStringSubmatch(content)
	if groups == nil {
		return result
	}
	for i, name := range re.SubexpNames() {
		if i > 0 && name != "" && i < len(groups) {
			result[name] = groups[i]
		}
	}
	return result
}

// ExtractMulti builds a string from the first match using $1, $2, ... placeholders.
func ExtractMulti(pattern, content, template string) string {
	return ExtractMultiWithOptions(pattern, content, template)
}

// ExtractMultiWithOptions builds a string from the first match using $1, $2, ... placeholders with options.
func ExtractMultiWithOptions(pattern, content, template string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return ""
	}
	return ExtractMultiRe(re, content, template)
}

// ExtractMultiRe builds a string from the first match of a compiled expression using $1, $2, ... placeholders.
func ExtractMultiRe(re *regexp.Regexp, content, template string) string {
	if re == nil {
		return ""
	}
	loc := re.FindStringSubmatchIndex(content)
	if loc == nil {
		return ""
	}
	return string(re.ExpandString(nil, template, content, loc))
}

// ExtractMultiAndDelPre extracts with a template and removes the consumed prefix from contentHolder.
func ExtractMultiAndDelPre(pattern string, contentHolder *string, template string) string {
	return ExtractMultiAndDelPreWithOptions(pattern, contentHolder, template)
}

// ExtractMultiAndDelPreWithOptions extracts with a template and removes the consumed prefix with options.
func ExtractMultiAndDelPreWithOptions(
	pattern string,
	contentHolder *string,
	template string,
	opts ...Option,
) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return ""
	}
	return ExtractMultiAndDelPreRe(re, contentHolder, template)
}

// ExtractMultiAndDelPreRe extracts with a template and removes the consumed prefix from contentHolder.
func ExtractMultiAndDelPreRe(re *regexp.Regexp, contentHolder *string, template string) string {
	if re == nil || contentHolder == nil {
		return ""
	}
	content := *contentHolder
	loc := re.FindStringSubmatchIndex(content)
	if loc == nil {
		return ""
	}
	*contentHolder = content[loc[1]:]
	return string(re.ExpandString(nil, template, content, loc))
}

// DelFirst deletes the first match.
func DelFirst(pattern, content string) string {
	return DelFirstWithOptions(pattern, content)
}

// DelFirstWithOptions deletes the first match with options.
func DelFirstWithOptions(pattern, content string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return content
	}
	return DelFirstRe(re, content)
}

// DelFirstRe deletes the first match of a compiled expression.
func DelFirstRe(re *regexp.Regexp, content string) string { return ReplaceFirstRe(re, content, "") }

// ReplaceFirst replaces the first match.
func ReplaceFirst(pattern, content, replacement string) string {
	return ReplaceFirstWithOptions(pattern, content, replacement)
}

// ReplaceFirstWithOptions replaces the first match with options.
func ReplaceFirstWithOptions(pattern, content, replacement string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return content
	}
	return ReplaceFirstRe(re, content, replacement)
}

// ReplaceFirstRe replaces the first match of a compiled expression.
func ReplaceFirstRe(re *regexp.Regexp, content, replacement string) string {
	if re == nil || content == "" {
		return content
	}
	loc := re.FindStringSubmatchIndex(content)
	if loc == nil {
		return content
	}
	repl := re.ExpandString(nil, replacement, content, loc)
	return content[:loc[0]] + string(repl) + content[loc[1]:]
}

// DelLast deletes the last match.
func DelLast(pattern, content string) string {
	return DelLastWithOptions(pattern, content)
}

// DelLastWithOptions deletes the last match with options.
func DelLastWithOptions(pattern, content string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return content
	}
	return DelLastRe(re, content)
}

// DelLastRe deletes the last match of a compiled expression.
func DelLastRe(re *regexp.Regexp, content string) string {
	match := LastIndexOfRe(re, content)
	if match == nil {
		return content
	}
	return content[:match.Start] + content[match.End:]
}

// DelAll deletes every match.
func DelAll(pattern, content string) string { return DelAllWithOptions(pattern, content) }

// DelAllWithOptions deletes every match with options.
func DelAllWithOptions(pattern, content string, opts ...Option) string {
	return ReReplaceWithOptions(pattern, content, "", opts...)
}

// DelAllRe deletes every match of a compiled expression.
func DelAllRe(re *regexp.Regexp, content string) string {
	if re == nil || content == "" {
		return content
	}
	return re.ReplaceAllString(content, "")
}

// DelPre deletes everything through the first match. If no match exists, content is returned unchanged.
func DelPre(pattern, content string) string {
	return DelPreWithOptions(pattern, content)
}

// DelPreWithOptions deletes everything through the first match with options.
func DelPreWithOptions(pattern, content string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return content
	}
	return DelPreRe(re, content)
}

// DelPreRe deletes everything through the first match of a compiled expression.
func DelPreRe(re *regexp.Regexp, content string) string {
	match := IndexOfRe(re, content)
	if match == nil {
		return content
	}
	return content[match.End:]
}

// FindAllGroup0 returns all full-match strings.
func FindAllGroup0(pattern, content string) []string {
	return FindAllGroup0WithOptions(pattern, content)
}

// FindAllGroup0WithOptions returns all full-match strings with options.
func FindAllGroup0WithOptions(pattern, content string, opts ...Option) []string {
	return FindAllWithOptions(pattern, content, 0, opts...)
}

// FindAllGroup1 returns all first capture groups.
func FindAllGroup1(pattern, content string) []string {
	return FindAllGroup1WithOptions(pattern, content)
}

// FindAllGroup1WithOptions returns all first capture groups with options.
func FindAllGroup1WithOptions(pattern, content string, opts ...Option) []string {
	return FindAllWithOptions(pattern, content, 1, opts...)
}

// FindAll returns all values for a capture group.
func FindAll(pattern, content string, group int) []string {
	return FindAllWithOptions(pattern, content, group)
}

// FindAllWithOptions returns all values for a capture group with options.
func FindAllWithOptions(pattern, content string, group int, opts ...Option) []string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return nil
	}
	return FindAllRe(re, content, group)
}

// FindAllRe returns all values for a capture group of a compiled expression.
func FindAllRe(re *regexp.Regexp, content string, group int) []string {
	if re == nil || group < 0 {
		return nil
	}
	all := re.FindAllStringSubmatch(content, -1)
	result := make([]string, 0, len(all))
	for _, groups := range all {
		if group < len(groups) {
			result = append(result, groups[group])
		}
	}
	return result
}

// Each calls consumer for every match.
func Each(re *regexp.Regexp, content string, consumer func(MatchResult)) {
	if re == nil || consumer == nil {
		return
	}
	for _, loc := range re.FindAllStringSubmatchIndex(content, -1) {
		consumer(buildMatchResult(re, content, loc))
	}
}

// Count returns the number of matches.
func Count(pattern, content string) int {
	return CountWithOptions(pattern, content)
}

// CountWithOptions returns the number of matches with options.
func CountWithOptions(pattern, content string, opts ...Option) int {
	re, err := compile(pattern, opts...)
	if err != nil {
		return 0
	}
	return CountRe(re, content)
}

// CountRe returns the number of matches for a compiled expression.
func CountRe(re *regexp.Regexp, content string) int {
	if re == nil {
		return 0
	}
	return len(re.FindAllStringIndex(content, -1))
}

// Contains reports whether content contains a match.
func Contains(pattern, content string) bool { return ReMatch(pattern, content) }

// ContainsWithOptions reports whether content contains a match with options.
func ContainsWithOptions(pattern, content string, opts ...Option) bool {
	return ReMatchWithOptions(pattern, content, opts...)
}

// ContainsRe reports whether content contains a match for a compiled expression.
func ContainsRe(re *regexp.Regexp, content string) bool { return re != nil && re.MatchString(content) }

// IndexOf returns the first match result.
func IndexOf(pattern, content string) *MatchResult {
	return IndexOfWithOptions(pattern, content)
}

// IndexOfWithOptions returns the first match result with options.
func IndexOfWithOptions(pattern, content string, opts ...Option) *MatchResult {
	re, err := compile(pattern, opts...)
	if err != nil {
		return nil
	}
	return IndexOfRe(re, content)
}

// IndexOfRe returns the first match result for a compiled expression.
func IndexOfRe(re *regexp.Regexp, content string) *MatchResult {
	if re == nil {
		return nil
	}
	loc := re.FindStringSubmatchIndex(content)
	if loc == nil {
		return nil
	}
	result := buildMatchResult(re, content, loc)
	return &result
}

// LastIndexOf returns the last match result.
func LastIndexOf(pattern, content string) *MatchResult {
	return LastIndexOfWithOptions(pattern, content)
}

// LastIndexOfWithOptions returns the last match result with options.
func LastIndexOfWithOptions(pattern, content string, opts ...Option) *MatchResult {
	re, err := compile(pattern, opts...)
	if err != nil {
		return nil
	}
	return LastIndexOfRe(re, content)
}

// LastIndexOfRe returns the last match result for a compiled expression.
func LastIndexOfRe(re *regexp.Regexp, content string) *MatchResult {
	if re == nil {
		return nil
	}
	locs := re.FindAllStringSubmatchIndex(content, -1)
	if len(locs) == 0 {
		return nil
	}
	result := buildMatchResult(re, content, locs[len(locs)-1])
	return &result
}

// GetFirstNumber returns the first integer in content.
func GetFirstNumber(content string) (int, bool) {
	return GetFirstNumberWithOptions(content)
}

// GetFirstNumberWithOptions returns the first integer in content with options.
func GetFirstNumberWithOptions(content string, opts ...Option) (int, bool) {
	number := applyOptions(opts).numbersRegexp.FindString(content)
	if number == "" {
		return 0, false
	}
	v, err := strconv.Atoi(number)
	if err != nil {
		return 0, false
	}
	return v, true
}

// IsMatch reports whether the whole content matches pattern. Empty pattern matches all non-empty inputs.
func IsMatch(pattern, content string) bool {
	return IsMatchWithOptions(pattern, content)
}

// IsMatchWithOptions reports whether the whole content matches pattern with options.
func IsMatchWithOptions(pattern, content string, opts ...Option) bool {
	if pattern == "" {
		return content != ""
	}
	re, err := compile(pattern, opts...)
	if err != nil {
		return false
	}
	return IsMatchRe(re, content)
}

// IsMatchRe reports whether the whole content matches a compiled expression.
func IsMatchRe(re *regexp.Regexp, content string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(content)
	return loc != nil && loc[0] == 0 && loc[1] == len(content)
}

// ReplaceAll replaces all matches using a template with $1, $2, ... placeholders.
func ReplaceAll(content, pattern, replacementTemplate string) string {
	return ReplaceAllWithOptions(content, pattern, replacementTemplate)
}

// ReplaceAllWithOptions replaces all matches using a template with options.
func ReplaceAllWithOptions(content, pattern, replacementTemplate string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return content
	}
	return ReplaceAllRe(content, re, replacementTemplate)
}

// ReplaceAllRe replaces all matches of a compiled expression using a template with $1, $2, ... placeholders.
func ReplaceAllRe(content string, re *regexp.Regexp, replacementTemplate string) string {
	if re == nil || content == "" {
		return content
	}
	return re.ReplaceAllString(content, replacementTemplate)
}

// ReplaceAllFunc replaces all matches using a custom function.
func ReplaceAllFunc(content, pattern string, replaceFunc func(MatchResult) string) string {
	return ReplaceAllFuncWithOptions(content, pattern, replaceFunc)
}

// ReplaceAllFuncWithOptions replaces all matches using a custom function with options.
func ReplaceAllFuncWithOptions(content, pattern string, replaceFunc func(MatchResult) string, opts ...Option) string {
	re, err := compile(pattern, opts...)
	if err != nil {
		return content
	}
	return ReplaceAllFuncRe(content, re, replaceFunc)
}

// ReplaceAllFuncRe replaces all matches of a compiled expression using a custom function.
func ReplaceAllFuncRe(content string, re *regexp.Regexp, replaceFunc func(MatchResult) string) string {
	if re == nil || replaceFunc == nil || content == "" {
		return content
	}
	locs := re.FindAllStringSubmatchIndex(content, -1)
	if len(locs) == 0 {
		return content
	}
	var b strings.Builder
	last := 0
	for _, loc := range locs {
		b.WriteString(content[last:loc[0]])
		b.WriteString(replaceFunc(buildMatchResult(re, content, loc)))
		last = loc[1]
	}
	b.WriteString(content[last:])
	return b.String()
}

// EscapeChar escapes a single regular-expression keyword character.
func EscapeChar(c rune) string {
	if _, ok := reKeys[c]; ok {
		return `\` + string(c)
	}
	return string(c)
}

// Escape escapes regular-expression keyword characters in content.
func Escape(content string) string {
	if strings.TrimSpace(content) == "" {
		return content
	}
	var b strings.Builder
	for _, r := range content {
		b.WriteString(EscapeChar(r))
	}
	return b.String()
}

// TemplateVars returns numeric placeholders referenced by a replacement template, longest first.
func TemplateVars(template string) []int {
	return TemplateVarsWithOptions(template)
}

// TemplateVarsWithOptions returns numeric placeholders referenced by a replacement template with options.
func TemplateVarsWithOptions(template string, opts ...Option) []int {
	matches := applyOptions(opts).groupVarRegexp.FindAllStringSubmatch(template, -1)
	seen := map[int]struct{}{}
	result := make([]int, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		v, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	slices.SortFunc(result, func(a, b int) int { return cmp.Compare(b, a) })
	return result
}

func compile(pattern string, opts ...Option) (*regexp.Regexp, error) {
	cfg := applyOptions(opts)
	pattern = normalizeNamedGroupsWithConfig(pattern, cfg)
	if cfg.dotAll {
		pattern = "(?s:" + pattern + ")"
	}
	return cfg.compile(pattern)
}

func normalizeNamedGroupsWithConfig(pattern string, cfg regexConfig) string {
	if cfg.normalizeNamedGroup != nil {
		return cfg.normalizeNamedGroup(pattern)
	}
	return cfg.namedGroupRegexp.ReplaceAllString(pattern, `(?P<$1>`)
}

func allLimit(findAll bool) int {
	if findAll {
		return -1
	}
	return 1
}

func buildMatchResult(re *regexp.Regexp, content string, loc []int) MatchResult {
	result := MatchResult{
		Text:       content[loc[0]:loc[1]],
		Start:      loc[0],
		End:        loc[1],
		Groups:     make([]string, len(loc)/2),
		GroupNames: map[string]string{},
	}
	names := re.SubexpNames()
	for i := 0; i < len(loc); i += 2 {
		group := i / 2
		if loc[i] >= 0 && loc[i+1] >= 0 {
			result.Groups[group] = content[loc[i]:loc[i+1]]
		}
		if group > 0 && group < len(names) && names[group] != "" {
			result.GroupNames[names[group]] = result.Groups[group]
		}
	}
	return result
}
