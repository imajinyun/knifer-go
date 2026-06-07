package dfa

import (
	"encoding/json"
	"strings"
	"sync"
	"unicode"
)

// CharFilter decides whether a rune participates in matching.
type CharFilter func(rune) bool

// Processor replaces a found word during text filtering.
type Processor func(FoundWord) string

// FoundWord describes a matched dictionary word and its location in the input.
type FoundWord struct {
	Word      string
	FoundWord string
	Start     int
	End       int
}

// String returns the matched text as it appeared in the input.
func (w FoundWord) String() string { return w.FoundWord }

type node struct {
	children map[rune]*node
	end      bool
	word     string
}

func newNode() *node { return &node{children: make(map[rune]*node)} }

// WordTree stores words in a rune trie and matches them with DFA-style scans.
type WordTree struct {
	root       *node
	charFilter CharFilter
}

type wordTreeConfig struct {
	charFilter CharFilter
}

// WordTreeOption customizes WordTree creation and package-level matcher initialization.
type WordTreeOption func(*wordTreeConfig)

// WithCharFilter sets the character filter used by a WordTree.
func WithCharFilter(filter CharFilter) WordTreeOption {
	return func(c *wordTreeConfig) {
		if filter != nil {
			c.charFilter = filter
		}
	}
}

func applyWordTreeOptions(opts []WordTreeOption) wordTreeConfig {
	cfg := wordTreeConfig{charFilter: IsNotStopChar}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.charFilter == nil {
		cfg.charFilter = IsNotStopChar
	}
	return cfg
}

// NewWordTree creates an empty word tree using the default stop-rune filter.
func NewWordTree() *WordTree {
	return NewWordTreeWithOptions()
}

// NewWordTreeWithOptions creates an empty word tree customized by options.
func NewWordTreeWithOptions(opts ...WordTreeOption) *WordTree {
	cfg := applyWordTreeOptions(opts)
	return &WordTree{root: newNode(), charFilter: cfg.charFilter}
}

// SetCharFilter sets the filter used to decide whether a rune participates in matching.
func (t *WordTree) SetCharFilter(filter CharFilter) *WordTree {
	if filter != nil {
		t.charFilter = filter
	}
	return t
}

// AddWords adds all words to the tree.
func (t *WordTree) AddWords(words ...string) *WordTree {
	seen := make(map[string]struct{}, len(words))
	for _, word := range words {
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = struct{}{}
		t.AddWord(word)
	}
	return t
}

// AddWord adds a word to the tree after filtering ignored runes.
func (t *WordTree) AddWord(word string) *WordTree {
	if t.root == nil {
		t.root = newNode()
	}
	filter := t.filter()
	current := t.root
	var accepted []rune
	for _, r := range word {
		if !filter(r) {
			continue
		}
		child := current.children[r]
		if child == nil {
			child = newNode()
			current.children[r] = child
		}
		current = child
		accepted = append(accepted, r)
	}
	if len(accepted) > 0 {
		current.end = true
		current.word = string(accepted)
	}
	return t
}

// Clear removes all words from the tree.
func (t *WordTree) Clear() {
	t.root = newNode()
}

// IsEmpty reports whether the tree contains no words.
func (t *WordTree) IsEmpty() bool {
	return t == nil || t.root == nil || len(t.root.children) == 0
}

// IsMatch reports whether text contains at least one word from the tree.
func (t *WordTree) IsMatch(text string) bool {
	_, ok := t.MatchWord(text)
	return ok
}

// Match returns the first matched text.
func (t *WordTree) Match(text string) (string, bool) {
	found, ok := t.MatchWord(text)
	if !ok {
		return "", false
	}
	return found.FoundWord, true
}

// MatchWord returns the first matched word with position metadata.
func (t *WordTree) MatchWord(text string) (FoundWord, bool) {
	words := t.MatchAllWords(text, 1, false, false)
	if len(words) == 0 {
		return FoundWord{}, false
	}
	return words[0], true
}

// MatchAll returns all matched texts without a limit.
func (t *WordTree) MatchAll(text string) []string {
	return t.MatchAllLimit(text, -1)
}

// MatchAllLimit returns matched texts up to limit. Non-positive limit means no limit.
func (t *WordTree) MatchAllLimit(text string, limit int) []string {
	return t.MatchAllMode(text, limit, false, false)
}

// MatchAllMode returns matched texts with dense and greedy matching controls.
func (t *WordTree) MatchAllMode(text string, limit int, densityMatch, greedMatch bool) []string {
	words := t.MatchAllWords(text, limit, densityMatch, greedMatch)
	result := make([]string, 0, len(words))
	for _, word := range words {
		result = append(result, word.FoundWord)
	}
	return result
}

// MatchAllWords returns found words with dense and greedy matching controls.
func (t *WordTree) MatchAllWords(text string, limit int, densityMatch, greedMatch bool) []FoundWord {
	if t == nil || t.root == nil || text == "" {
		return nil
	}
	runes := []rune(text)
	found := make([]FoundWord, 0)
	filter := t.filter()
	for i := 0; i < len(runes); i++ {
		if !filter(runes[i]) {
			continue
		}
		current := t.root
		var foundRunes []rune
		var keyRunes []rune
		for j := i; j < len(runes); j++ {
			r := runes[j]
			if !filter(r) {
				if len(foundRunes) > 0 {
					foundRunes = append(foundRunes, r)
				}
				continue
			}
			child := current.children[r]
			if child == nil {
				break
			}
			foundRunes = append(foundRunes, r)
			keyRunes = append(keyRunes, r)
			current = child
			if current.end {
				word := current.word
				if word == "" {
					word = string(keyRunes)
				}
				found = append(found, FoundWord{Word: word, FoundWord: string(foundRunes), Start: i, End: j})
				if limit > 0 && len(found) >= limit {
					return found
				}
				if !densityMatch {
					i = j
					break
				}
				if !greedMatch {
					break
				}
			}
		}
	}
	return found
}

// Filter replaces matched words in text using processor or the default mask.
func (t *WordTree) Filter(text string, greedMatch bool, processor Processor) string {
	if text == "" {
		return text
	}
	found := t.MatchAllWords(text, -1, true, greedMatch)
	if len(found) == 0 {
		return text
	}
	if processor == nil {
		processor = DefaultProcessor
	}
	byStart := make(map[int]FoundWord, len(found))
	for _, word := range found {
		byStart[word.Start] = word
	}
	runes := []rune(text)
	var builder strings.Builder
	for i := 0; i < len(runes); i++ {
		if word, ok := byStart[i]; ok {
			builder.WriteString(processor(word))
			i = word.End
			continue
		}
		builder.WriteRune(runes[i])
	}
	return builder.String()
}

func (t *WordTree) filter() CharFilter {
	if t.charFilter == nil {
		return IsNotStopChar
	}
	return t.charFilter
}

// DefaultProcessor replaces each rune of the matched text with an asterisk.
func DefaultProcessor(word FoundWord) string {
	return strings.Repeat("*", len([]rune(word.FoundWord)))
}

var stopRunes = map[rune]struct{}{}

func init() {
	for _, r := range " '\u3001。·ˉˇ々—～‖…‘’“”〔〕〈〉《》「」『』〖〗【】±＋－×÷∧∨∑∏∪∩∈√⊥⊙∫∮≡≌≈∽∝≠≮≯≤≥∞∶∵∴∷♂♀°′〃℃＄¤￠￡‰§☆★〇○●◎◇◆□■△▽⊿▲▼◣◤◢◥▁▂▃▄▅▆▇█▉▊▋▌▍▎▏▓※→←↑↓↖↗↘↙〓ⅰⅱⅲⅳⅴⅵⅶⅷⅸⅹ①②③④⑤⑥⑦⑧⑨⑩⒈⒉⒊⒋⒌⒍⒎⒏⒐⒑⒒⒓⒔⒕⒖⒗⒘⒙⒚⒛⑴⑵⑶⑷⑸⑹⑺⑻⑼⑽⑾⑿⒀⒁⒂⒃⒄⒅⒆⒇ⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩⅪⅫ！＃￥％＆（）＊，．／０１２３４５６７８９：；＜＝＞？＠＼＾＿｛｜｝ΡΥΦΧΨΩαβγδεζηθικλμνξοπρστυφχψω﹊﹍╭╮╰╯_^/\\\"<>`{}~()-$@*&#卐㎎㎏㎜㎝㎞㎡㏄㏎㏑㏒㏕+=?:.!;]|%" {
		stopRunes[r] = struct{}{}
	}
}

// IsStopChar reports whether r should be ignored by the default matcher.
func IsStopChar(r rune) bool {
	if unicode.IsSpace(r) {
		return true
	}
	_, ok := stopRunes[r]
	return ok
}

// IsNotStopChar reports whether r should participate in matching.
func IsNotStopChar(r rune) bool { return !IsStopChar(r) }

// DefaultSeparator is used by InitString when no separator is provided.
const DefaultSeparator = ','

var defaultMatcher = struct {
	sync.RWMutex
	tree *WordTree
}{tree: NewWordTree()}

var defaultAsyncRunner = struct {
	sync.RWMutex
	runner func(func())
}{runner: goRun}

// MatcherOption customizes one package-level matcher operation.
type MatcherOption func(*matcherConfig)

type matcherConfig struct {
	tree      *WordTree
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

// WithMatcher sets the word tree used by one package-level matcher operation.
func WithMatcher(tree *WordTree) MatcherOption {
	return func(cfg *matcherConfig) {
		if tree != nil {
			cfg.tree = tree
		}
	}
}

// WithMatcherWords creates an isolated word tree for one package-level matcher operation.
func WithMatcherWords(words []string, opts ...WordTreeOption) MatcherOption {
	return WithMatcher(NewWordTreeWithOptions(opts...).AddWords(words...))
}

// WithJSONMarshal sets the marshal function used by Any helpers.
func WithJSONMarshal(marshal func(any) ([]byte, error)) MatcherOption {
	return func(cfg *matcherConfig) {
		if marshal != nil {
			cfg.marshal = marshal
		}
	}
}

// WithJSONUnmarshal sets the unmarshal function used by FilterAnyWithOptions.
func WithJSONUnmarshal(unmarshal func([]byte, any) error) MatcherOption {
	return func(cfg *matcherConfig) {
		if unmarshal != nil {
			cfg.unmarshal = unmarshal
		}
	}
}

func currentMatcher() *WordTree {
	defaultMatcher.RLock()
	defer defaultMatcher.RUnlock()
	return defaultMatcher.tree
}

func applyMatcherOptions(opts []MatcherOption) *WordTree {
	return applyMatcherConfig(opts).tree
}

func applyMatcherConfig(opts []MatcherOption) matcherConfig {
	cfg := matcherConfig{tree: currentMatcher()}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.tree == nil {
		cfg.tree = NewWordTree()
	}
	if cfg.marshal == nil {
		cfg.marshal = json.Marshal
	}
	if cfg.unmarshal == nil {
		cfg.unmarshal = json.Unmarshal
	}
	return cfg
}

// ConfigureAsyncRunner sets the runner used by asynchronous package-level matcher initialization.
// Passing nil restores the default goroutine runner.
func ConfigureAsyncRunner(runner func(func())) {
	defaultAsyncRunner.Lock()
	defer defaultAsyncRunner.Unlock()
	if runner == nil {
		defaultAsyncRunner.runner = goRun
		return
	}
	defaultAsyncRunner.runner = runner
}

// ResetAsyncRunner restores the default goroutine runner used by asynchronous initialization.
func ResetAsyncRunner() { ConfigureAsyncRunner(nil) }

func runAsync(fn func()) {
	defaultAsyncRunner.RLock()
	runner := defaultAsyncRunner.runner
	defaultAsyncRunner.RUnlock()
	if runner == nil {
		runner = goRun
	}
	runner(fn)
}

func goRun(fn func()) { go fn() }

// IsInited reports whether the package-level matcher contains words.
func IsInited() bool {
	defaultMatcher.RLock()
	defer defaultMatcher.RUnlock()
	return !defaultMatcher.tree.IsEmpty()
}

// Init replaces the package-level matcher with words.
func Init(words []string) {
	InitWithOptions(words)
}

// InitWithOptions replaces the package-level matcher with words and initialization options.
func InitWithOptions(words []string, opts ...WordTreeOption) {
	tree := NewWordTreeWithOptions(opts...).AddWords(words...)
	defaultMatcher.Lock()
	defaultMatcher.tree = tree
	defaultMatcher.Unlock()
}

// InitAsync initializes the package-level matcher in a new goroutine.
func InitAsync(words []string) { runAsync(func() { Init(words) }) }

// InitString initializes the package-level matcher from a separated string.
func InitString(words string, separator rune) {
	InitStringWithOptions(words, separator)
}

// InitStringWithOptions initializes the package-level matcher from a separated string and options.
func InitStringWithOptions(words string, separator rune, opts ...WordTreeOption) {
	if separator == 0 {
		separator = DefaultSeparator
	}
	parts := strings.Split(words, string(separator))
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			clean = append(clean, part)
		}
	}
	InitWithOptions(clean, opts...)
}

// InitStringAsync initializes the package-level matcher from a separated string in a new goroutine.
func InitStringAsync(words string, separator rune) { runAsync(func() { InitString(words, separator) }) }

// InitStringAsyncWithOptions initializes the package-level matcher from a separated string in a new goroutine.
func InitStringAsyncWithOptions(words string, separator rune, opts ...WordTreeOption) {
	runAsync(func() { InitStringWithOptions(words, separator, opts...) })
}

// SetCharFilter sets the filter used by the package-level matcher.
func SetCharFilter(filter CharFilter) {
	if filter == nil {
		return
	}
	defaultMatcher.Lock()
	defaultMatcher.tree.SetCharFilter(filter)
	defaultMatcher.Unlock()
}

// Contains reports whether text contains a word from the package-level matcher.
func Contains(text string) bool {
	return ContainsWithOptions(text)
}

// ContainsWithOptions reports whether text contains a word from the selected matcher.
func ContainsWithOptions(text string, opts ...MatcherOption) bool {
	if len(opts) > 0 {
		return applyMatcherOptions(opts).IsMatch(text)
	}
	defaultMatcher.RLock()
	defer defaultMatcher.RUnlock()
	return defaultMatcher.tree.IsMatch(text)
}

// ContainsAny marshals value as JSON and checks it with the package-level matcher.
func ContainsAny(value any) bool { return ContainsAnyWithOptions(value) }

// ContainsAnyWithOptions marshals value as JSON and checks it with the selected matcher.
func ContainsAnyWithOptions(value any, opts ...MatcherOption) bool {
	cfg := applyMatcherConfig(opts)
	return cfg.tree.IsMatch(jsonTextWithMarshal(value, cfg.marshal))
}

// GetFoundFirst returns the first found word from the package-level matcher.
func GetFoundFirst(text string) (FoundWord, bool) {
	return GetFoundFirstWithOptions(text)
}

// GetFoundFirstWithOptions returns the first found word from the selected matcher.
func GetFoundFirstWithOptions(text string, opts ...MatcherOption) (FoundWord, bool) {
	if len(opts) > 0 {
		return applyMatcherOptions(opts).MatchWord(text)
	}
	defaultMatcher.RLock()
	defer defaultMatcher.RUnlock()
	return defaultMatcher.tree.MatchWord(text)
}

// GetFoundFirstAny marshals value as JSON and returns the first found word.
func GetFoundFirstAny(value any) (FoundWord, bool) { return GetFoundFirstAnyWithOptions(value) }

// GetFoundFirstAnyWithOptions marshals value as JSON and returns the first found word from the selected matcher.
func GetFoundFirstAnyWithOptions(value any, opts ...MatcherOption) (FoundWord, bool) {
	cfg := applyMatcherConfig(opts)
	return cfg.tree.MatchWord(jsonTextWithMarshal(value, cfg.marshal))
}

// GetFoundAll returns all found words from the package-level matcher.
func GetFoundAll(text string) []FoundWord {
	return GetFoundAllWithOptions(text)
}

// GetFoundAllWithOptions returns all found words from the selected matcher.
func GetFoundAllWithOptions(text string, opts ...MatcherOption) []FoundWord {
	return GetFoundAllModeWithOptions(text, false, false, opts...)
}

// GetFoundAllMode returns all found words with dense and greedy matching controls.
func GetFoundAllMode(text string, densityMatch, greedMatch bool) []FoundWord {
	return GetFoundAllModeWithOptions(text, densityMatch, greedMatch)
}

// GetFoundAllModeWithOptions returns all found words from the selected matcher with dense and greedy matching controls.
func GetFoundAllModeWithOptions(text string, densityMatch, greedMatch bool, opts ...MatcherOption) []FoundWord {
	if len(opts) > 0 {
		return applyMatcherOptions(opts).MatchAllWords(text, -1, densityMatch, greedMatch)
	}
	defaultMatcher.RLock()
	defer defaultMatcher.RUnlock()
	return defaultMatcher.tree.MatchAllWords(text, -1, densityMatch, greedMatch)
}

// GetFoundAllAny marshals value as JSON and returns all found words.
func GetFoundAllAny(value any) []FoundWord { return GetFoundAllAnyWithOptions(value) }

// GetFoundAllAnyWithOptions marshals value as JSON and returns all found words from the selected matcher.
func GetFoundAllAnyWithOptions(value any, opts ...MatcherOption) []FoundWord {
	cfg := applyMatcherConfig(opts)
	return cfg.tree.MatchAllWords(jsonTextWithMarshal(value, cfg.marshal), -1, false, false)
}

// Filter replaces words found by the package-level matcher.
func Filter(text string) string { return FilterWithOptions(text) }

// FilterWithOptions replaces words found by the selected matcher.
func FilterWithOptions(text string, opts ...MatcherOption) string {
	return FilterModeWithOptions(text, true, nil, opts...)
}

// FilterMode replaces words found by the package-level matcher using processor.
func FilterMode(text string, greedMatch bool, processor Processor) string {
	return FilterModeWithOptions(text, greedMatch, processor)
}

// FilterModeWithOptions replaces words found by the selected matcher using processor.
func FilterModeWithOptions(text string, greedMatch bool, processor Processor, opts ...MatcherOption) string {
	if len(opts) > 0 {
		return applyMatcherOptions(opts).Filter(text, greedMatch, processor)
	}
	defaultMatcher.RLock()
	defer defaultMatcher.RUnlock()
	return defaultMatcher.tree.Filter(text, greedMatch, processor)
}

// FilterAny marshals value as JSON, filters matched text, and unmarshals it back.
func FilterAny[T any](value T, greedMatch bool, processor Processor) (T, error) {
	return FilterAnyWithOptions(value, greedMatch, processor)
}

// FilterAnyWithOptions marshals value, filters matched text with the selected matcher, and unmarshals it back.
func FilterAnyWithOptions[T any](value T, greedMatch bool, processor Processor, opts ...MatcherOption) (T, error) {
	if s, ok := any(value).(string); ok {
		return any(FilterModeWithOptions(s, greedMatch, processor, opts...)).(T), nil
	}
	var result T
	cfg := applyMatcherConfig(opts)
	data, err := cfg.marshal(value)
	if err != nil {
		return result, err
	}
	filtered := cfg.tree.Filter(string(data), greedMatch, processor)
	if err := cfg.unmarshal([]byte(filtered), &result); err != nil {
		return result, err
	}
	return result, nil
}

func jsonTextWithMarshal(value any, marshal func(any) ([]byte, error)) string {
	if s, ok := value.(string); ok {
		return s
	}
	if marshal == nil {
		marshal = json.Marshal
	}
	data, err := marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}
