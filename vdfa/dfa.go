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

// DefaultSeparator is used by InitString when no separator is provided.
const DefaultSeparator = dfaimpl.DefaultSeparator

// NewWordTree creates an empty word tree using the default stop-rune filter.
func NewWordTree() *WordTree { return dfaimpl.NewWordTree() }

// WithCharFilter sets the character filter used by a WordTree.
func WithCharFilter(filter CharFilter) WordTreeOption { return dfaimpl.WithCharFilter(filter) }

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

// ContainsAny marshals value as JSON and checks it with the package-level matcher.
func ContainsAny(value any) bool { return dfaimpl.ContainsAny(value) }

// GetFoundFirst returns the first found word from the package-level matcher.
func GetFoundFirst(text string) (FoundWord, bool) { return dfaimpl.GetFoundFirst(text) }

// GetFoundFirstAny marshals value as JSON and returns the first found word.
func GetFoundFirstAny(value any) (FoundWord, bool) { return dfaimpl.GetFoundFirstAny(value) }

// GetFoundAll returns all found words from the package-level matcher.
func GetFoundAll(text string) []FoundWord { return dfaimpl.GetFoundAll(text) }

// GetFoundAllMode returns all found words with dense and greedy matching controls.
func GetFoundAllMode(text string, densityMatch, greedMatch bool) []FoundWord {
	return dfaimpl.GetFoundAllMode(text, densityMatch, greedMatch)
}

// GetFoundAllAny marshals value as JSON and returns all found words.
func GetFoundAllAny(value any) []FoundWord { return dfaimpl.GetFoundAllAny(value) }

// Filter replaces words found by the package-level matcher.
func Filter(text string) string { return dfaimpl.Filter(text) }

// FilterMode replaces words found by the package-level matcher using processor.
func FilterMode(text string, greedMatch bool, processor Processor) string {
	return dfaimpl.FilterMode(text, greedMatch, processor)
}

// FilterAny marshals value as JSON, filters matched text, and unmarshals it back.
func FilterAny[T any](value T, greedMatch bool, processor Processor) (T, error) {
	return dfaimpl.FilterAny(value, greedMatch, processor)
}
