package tokenize

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	// ErrInvalidTokenizeRequest reports a malformed tokenization request.
	ErrInvalidTokenizeRequest = errors.New("tokenize: invalid tokenize request")
	// ErrInvalidKeywordsRequest reports a malformed keyword extraction request.
	ErrInvalidKeywordsRequest = errors.New("tokenize: invalid keywords request")
	// ErrInputLimitExceeded reports that input text exceeded its configured rune limit.
	ErrInputLimitExceeded = errors.New("tokenize: input limit exceeded")
	// ErrTokenLimitExceeded reports that provider output exceeded its configured token limit.
	ErrTokenLimitExceeded = errors.New("tokenize: token limit exceeded")
)

// Mode describes a provider-specific tokenization mode.
type Mode string

const (
	// ModeDefault leaves segmentation behavior to the configured provider.
	ModeDefault Mode = ""
	// ModePrecise requests precise segmentation.
	ModePrecise Mode = "precise"
	// ModeSearch requests search-engine-oriented segmentation.
	ModeSearch Mode = "search"
	// ModeFull requests full segmentation when supported by the provider.
	ModeFull Mode = "full"
)

// Token describes one token emitted by a tokenizer provider.
type Token struct {
	Text     string
	Start    int
	End      int
	Position int
	Weight   float64
	Metadata map[string]string
}

// Keyword describes one keyword emitted by a keyword extraction provider.
type Keyword struct {
	Text     string
	Score    float64
	Metadata map[string]string
}

// TokenizeRequest describes a tokenization request.
type TokenizeRequest struct {
	Text          string
	Mode          Mode
	KeepPunct     bool
	MaxInputRunes int
	MaxTokens     int
	Metadata      map[string]string
}

// TokenizeResponse contains tokenization output and metadata.
type TokenizeResponse struct {
	Text     string
	Tokens   []Token
	Metadata map[string]string
}

// KeywordsRequest describes a keyword extraction request.
type KeywordsRequest struct {
	Text          string
	Limit         int
	MaxInputRunes int
	Metadata      map[string]string
}

// KeywordsResponse contains extracted keywords and metadata.
type KeywordsResponse struct {
	Text     string
	Keywords []Keyword
	Metadata map[string]string
}

// Validate checks whether r has valid input text, mode, and limits.
func (r TokenizeRequest) Validate() error {
	if err := validateText(r.Text, r.MaxInputRunes, ErrInvalidTokenizeRequest); err != nil {
		return err
	}
	if r.MaxTokens < 0 {
		return fmt.Errorf("%w: max tokens must be non-negative", ErrInvalidTokenizeRequest)
	}
	return validateMode(r.Mode)
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r TokenizeRequest) Clone() TokenizeRequest {
	return TokenizeRequest{
		Text:          r.Text,
		Mode:          r.Mode,
		KeepPunct:     r.KeepPunct,
		MaxInputRunes: r.MaxInputRunes,
		MaxTokens:     r.MaxTokens,
		Metadata:      cloneStringMap(r.Metadata),
	}
}

// Clone returns a response copy that callers can mutate independently.
func (r TokenizeResponse) Clone() TokenizeResponse {
	return TokenizeResponse{Text: r.Text, Tokens: cloneTokens(r.Tokens), Metadata: cloneStringMap(r.Metadata)}
}

// Validate checks whether r has valid input text and limits.
func (r KeywordsRequest) Validate() error {
	if err := validateText(r.Text, r.MaxInputRunes, ErrInvalidKeywordsRequest); err != nil {
		return err
	}
	if r.Limit < 0 {
		return fmt.Errorf("%w: limit must be non-negative", ErrInvalidKeywordsRequest)
	}
	return nil
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r KeywordsRequest) Clone() KeywordsRequest {
	return KeywordsRequest{
		Text:          r.Text,
		Limit:         r.Limit,
		MaxInputRunes: r.MaxInputRunes,
		Metadata:      cloneStringMap(r.Metadata),
	}
}

// Clone returns a response copy that callers can mutate independently.
func (r KeywordsResponse) Clone() KeywordsResponse {
	return KeywordsResponse{Text: r.Text, Keywords: cloneKeywords(r.Keywords), Metadata: cloneStringMap(r.Metadata)}
}

func validateText(text string, maxRunes int, sentinel error) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("%w: text is required", sentinel)
	}
	if strings.ContainsRune(text, '\x00') {
		return fmt.Errorf("%w: text contains nul byte", sentinel)
	}
	if maxRunes < 0 {
		return fmt.Errorf("%w: max input runes must be non-negative", sentinel)
	}
	if maxRunes > 0 {
		runes := utf8.RuneCountInString(text)
		if runes > maxRunes {
			return fmt.Errorf("%w: input has %d runes, max %d", ErrInputLimitExceeded, runes, maxRunes)
		}
	}
	return nil
}

func validateMode(mode Mode) error {
	switch mode {
	case ModeDefault, ModePrecise, ModeSearch, ModeFull:
		return nil
	default:
		return fmt.Errorf("%w: unsupported mode %q", ErrInvalidTokenizeRequest, mode)
	}
}

func cloneTokens(tokens []Token) []Token {
	if len(tokens) == 0 {
		return nil
	}
	clone := make([]Token, len(tokens))
	for i, token := range tokens {
		clone[i] = token
		clone[i].Metadata = cloneStringMap(token.Metadata)
	}
	return clone
}

func cloneKeywords(keywords []Keyword) []Keyword {
	if len(keywords) == 0 {
		return nil
	}
	clone := make([]Keyword, len(keywords))
	for i, keyword := range keywords {
		clone[i] = keyword
		clone[i].Metadata = cloneStringMap(keyword.Metadata)
	}
	return clone
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	clone := make(map[string]string, len(values))
	for k, v := range values {
		clone[k] = v
	}
	return clone
}
