package vdfa

import dfaimpl "github.com/imajinyun/go-knifer/internal/dfa"

// CharFilter decides whether a rune participates in matching.
type CharFilter = dfaimpl.CharFilter

// Processor replaces a found word during text filtering.
type Processor = dfaimpl.Processor

// FoundWord describes a matched dictionary word and its location in the input.
type FoundWord = dfaimpl.FoundWord

// WordTree stores words in a rune trie and matches them with DFA-style scans.
type WordTree = dfaimpl.WordTree

// WordTreeOption customizes WordTree creation and package-level matcher initialization.
type WordTreeOption = dfaimpl.WordTreeOption

// MatcherOption customizes one package-level matcher operation.
type MatcherOption = dfaimpl.MatcherOption

// DefaultSeparator is used by InitString when no separator is provided.
const DefaultSeparator = dfaimpl.DefaultSeparator

// NewWordTree creates an empty word tree using the default stop-rune filter.
func NewWordTree() *WordTree { return dfaimpl.NewWordTree() }

// WithCharFilter sets the character filter used by a WordTree.
func WithCharFilter(filter CharFilter) WordTreeOption { return dfaimpl.WithCharFilter(filter) }

// WithMatcher sets the word tree used by one package-level matcher operation.
func WithMatcher(tree *WordTree) MatcherOption { return dfaimpl.WithMatcher(tree) }

// WithMatcherWords creates an isolated word tree for one package-level matcher operation.
func WithMatcherWords(words []string, opts ...WordTreeOption) MatcherOption {
	return dfaimpl.WithMatcherWords(words, opts...)
}

// WithJSONMarshal sets the marshal function used by Any helpers.
func WithJSONMarshal(marshal func(any) ([]byte, error)) MatcherOption {
	return dfaimpl.WithJSONMarshal(marshal)
}

// WithJSONUnmarshal sets the unmarshal function used by FilterAnyWithOptions.
func WithJSONUnmarshal(unmarshal func([]byte, any) error) MatcherOption {
	return dfaimpl.WithJSONUnmarshal(unmarshal)
}

// NewWordTreeWithOptions creates an empty word tree customized by options.
func NewWordTreeWithOptions(opts ...WordTreeOption) *WordTree {
	return dfaimpl.NewWordTreeWithOptions(opts...)
}

// DefaultProcessor replaces each rune of the matched text with an asterisk.
func DefaultProcessor(word FoundWord) string { return dfaimpl.DefaultProcessor(word) }

// IsStopChar reports whether r should be ignored by the default matcher.
func IsStopChar(r rune) bool { return dfaimpl.IsStopChar(r) }

// IsNotStopChar reports whether r should participate in matching.
func IsNotStopChar(r rune) bool { return dfaimpl.IsNotStopChar(r) }

// IsInited reports whether the package-level matcher contains words.
func IsInited() bool { return dfaimpl.IsInited() }

// Init replaces the package-level matcher with words.
func Init(words []string) { dfaimpl.Init(words) }

// InitWithOptions replaces the package-level matcher with words and initialization options.
func InitWithOptions(words []string, opts ...WordTreeOption) { dfaimpl.InitWithOptions(words, opts...) }

// InitAsync initializes the package-level matcher in a new goroutine.
func InitAsync(words []string) { dfaimpl.InitAsync(words) }

// ConfigureAsyncRunner sets the runner used by asynchronous package-level matcher initialization.
func ConfigureAsyncRunner(runner func(func())) { dfaimpl.ConfigureAsyncRunner(runner) }

// ResetAsyncRunner restores the default goroutine runner used by asynchronous initialization.
func ResetAsyncRunner() { dfaimpl.ResetAsyncRunner() }

// InitString initializes the package-level matcher from a separated string.
func InitString(words string, separator rune) { dfaimpl.InitString(words, separator) }

// InitStringWithOptions initializes the package-level matcher from a separated string and options.
func InitStringWithOptions(words string, separator rune, opts ...WordTreeOption) {
	dfaimpl.InitStringWithOptions(words, separator, opts...)
}

// InitStringAsync initializes the package-level matcher from a separated string in a new goroutine.
func InitStringAsync(words string, separator rune) { dfaimpl.InitStringAsync(words, separator) }

// InitStringAsyncWithOptions initializes the package-level matcher from a separated string in a new goroutine.
func InitStringAsyncWithOptions(words string, separator rune, opts ...WordTreeOption) {
	dfaimpl.InitStringAsyncWithOptions(words, separator, opts...)
}

// SetCharFilter sets the filter used by the package-level matcher.
func SetCharFilter(filter CharFilter) { dfaimpl.SetCharFilter(filter) }

// Contains reports whether text contains a word from the package-level matcher.
func Contains(text string) bool { return dfaimpl.Contains(text) }

// ContainsWithOptions reports whether text contains a word from the selected matcher.
func ContainsWithOptions(text string, opts ...MatcherOption) bool {
	return dfaimpl.ContainsWithOptions(text, opts...)
}

// ContainsAny marshals value as JSON and checks it with the package-level matcher.
func ContainsAny(value any) bool { return dfaimpl.ContainsAny(value) }

// ContainsAnyWithOptions marshals value as JSON and checks it with the selected matcher.
func ContainsAnyWithOptions(value any, opts ...MatcherOption) bool {
	return dfaimpl.ContainsAnyWithOptions(value, opts...)
}

// GetFoundFirst returns the first found word from the package-level matcher.
func GetFoundFirst(text string) (FoundWord, bool) { return dfaimpl.GetFoundFirst(text) }

// GetFoundFirstWithOptions returns the first found word from the selected matcher.
func GetFoundFirstWithOptions(text string, opts ...MatcherOption) (FoundWord, bool) {
	return dfaimpl.GetFoundFirstWithOptions(text, opts...)
}

// GetFoundFirstAny marshals value as JSON and returns the first found word.
func GetFoundFirstAny(value any) (FoundWord, bool) { return dfaimpl.GetFoundFirstAny(value) }

// GetFoundFirstAnyWithOptions marshals value as JSON and returns the first found word from the selected matcher.
func GetFoundFirstAnyWithOptions(value any, opts ...MatcherOption) (FoundWord, bool) {
	return dfaimpl.GetFoundFirstAnyWithOptions(value, opts...)
}

// GetFoundAll returns all found words from the package-level matcher.
func GetFoundAll(text string) []FoundWord { return dfaimpl.GetFoundAll(text) }

// GetFoundAllWithOptions returns all found words from the selected matcher.
func GetFoundAllWithOptions(text string, opts ...MatcherOption) []FoundWord {
	return dfaimpl.GetFoundAllWithOptions(text, opts...)
}

// GetFoundAllMode returns all found words with dense and greedy matching controls.
func GetFoundAllMode(text string, densityMatch, greedMatch bool) []FoundWord {
	return dfaimpl.GetFoundAllMode(text, densityMatch, greedMatch)
}

// GetFoundAllModeWithOptions returns all found words from the selected matcher with dense and greedy matching controls.
func GetFoundAllModeWithOptions(text string, densityMatch, greedMatch bool, opts ...MatcherOption) []FoundWord {
	return dfaimpl.GetFoundAllModeWithOptions(text, densityMatch, greedMatch, opts...)
}

// GetFoundAllAny marshals value as JSON and returns all found words.
func GetFoundAllAny(value any) []FoundWord { return dfaimpl.GetFoundAllAny(value) }

// GetFoundAllAnyWithOptions marshals value as JSON and returns all found words from the selected matcher.
func GetFoundAllAnyWithOptions(value any, opts ...MatcherOption) []FoundWord {
	return dfaimpl.GetFoundAllAnyWithOptions(value, opts...)
}

// Filter replaces words found by the package-level matcher.
func Filter(text string) string { return dfaimpl.Filter(text) }

// FilterWithOptions replaces words found by the selected matcher.
func FilterWithOptions(text string, opts ...MatcherOption) string {
	return dfaimpl.FilterWithOptions(text, opts...)
}

// FilterMode replaces words found by the package-level matcher using processor.
func FilterMode(text string, greedMatch bool, processor Processor) string {
	return dfaimpl.FilterMode(text, greedMatch, processor)
}

// FilterModeWithOptions replaces words found by the selected matcher using processor.
func FilterModeWithOptions(text string, greedMatch bool, processor Processor, opts ...MatcherOption) string {
	return dfaimpl.FilterModeWithOptions(text, greedMatch, processor, opts...)
}

// FilterAny marshals value as JSON, filters matched text, and unmarshals it back.
func FilterAny[T any](value T, greedMatch bool, processor Processor) (T, error) {
	return dfaimpl.FilterAny(value, greedMatch, processor)
}

// FilterAnyWithOptions marshals value, filters matched text with the selected matcher, and unmarshals it back.
func FilterAnyWithOptions[T any](value T, greedMatch bool, processor Processor, opts ...MatcherOption) (T, error) {
	return dfaimpl.FilterAnyWithOptions(value, greedMatch, processor, opts...)
}
