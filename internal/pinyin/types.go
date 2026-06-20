package pinyin

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	// ErrInvalidConvertRequest reports a malformed pinyin conversion request.
	ErrInvalidConvertRequest = errors.New("pinyin: invalid convert request")
	// ErrInvalidInitialsRequest reports a malformed pinyin initials request.
	ErrInvalidInitialsRequest = errors.New("pinyin: invalid initials request")
	// ErrInputLimitExceeded reports that input text exceeded its configured rune limit.
	ErrInputLimitExceeded = errors.New("pinyin: input limit exceeded")
)

// ToneStyle describes how converted syllables should represent tones.
type ToneStyle string

const (
	// ToneStyleDefault leaves tone rendering to the configured provider.
	ToneStyleDefault ToneStyle = ""
	// ToneStylePlain requests syllables without tone marks or tone numbers.
	ToneStylePlain ToneStyle = "plain"
	// ToneStyleNumber requests syllables with numeric tone suffixes.
	ToneStyleNumber ToneStyle = "number"
	// ToneStyleMark requests syllables with diacritic tone marks.
	ToneStyleMark ToneStyle = "mark"
)

// Token describes one converted input segment and its provider-specific syllables.
type Token struct {
	Text      string
	Syllables []string
	Metadata  map[string]string
}

// ConvertRequest describes a pinyin conversion request.
type ConvertRequest struct {
	Text          string
	Separator     string
	ToneStyle     ToneStyle
	Heteronym     bool
	MaxInputRunes int
	Metadata      map[string]string
}

// ConvertResponse contains pinyin conversion output and token metadata.
type ConvertResponse struct {
	Text     string
	Output   string
	Tokens   []Token
	Metadata map[string]string
}

// InitialsRequest describes a pinyin initials extraction request.
type InitialsRequest struct {
	Text          string
	Separator     string
	MaxInputRunes int
	Metadata      map[string]string
}

// InitialsResponse contains initials extraction output and metadata.
type InitialsResponse struct {
	Text     string
	Output   string
	Initials []string
	Metadata map[string]string
}

// Validate checks whether r has valid input text, tone style, and input rune limit.
func (r ConvertRequest) Validate() error {
	if err := validateText(r.Text, r.MaxInputRunes, ErrInvalidConvertRequest); err != nil {
		return err
	}
	return validateToneStyle(r.ToneStyle)
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r ConvertRequest) Clone() ConvertRequest {
	return ConvertRequest{
		Text:          r.Text,
		Separator:     r.Separator,
		ToneStyle:     r.ToneStyle,
		Heteronym:     r.Heteronym,
		MaxInputRunes: r.MaxInputRunes,
		Metadata:      cloneStringMap(r.Metadata),
	}
}

// Clone returns a response copy that callers can mutate independently.
func (r ConvertResponse) Clone() ConvertResponse {
	return ConvertResponse{Text: r.Text, Output: r.Output, Tokens: cloneTokens(r.Tokens), Metadata: cloneStringMap(r.Metadata)}
}

// Validate checks whether r has valid input text and input rune limit.
func (r InitialsRequest) Validate() error {
	return validateText(r.Text, r.MaxInputRunes, ErrInvalidInitialsRequest)
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r InitialsRequest) Clone() InitialsRequest {
	return InitialsRequest{
		Text:          r.Text,
		Separator:     r.Separator,
		MaxInputRunes: r.MaxInputRunes,
		Metadata:      cloneStringMap(r.Metadata),
	}
}

// Clone returns a response copy that callers can mutate independently.
func (r InitialsResponse) Clone() InitialsResponse {
	return InitialsResponse{
		Text:     r.Text,
		Output:   r.Output,
		Initials: append([]string(nil), r.Initials...),
		Metadata: cloneStringMap(r.Metadata),
	}
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

func validateToneStyle(style ToneStyle) error {
	switch style {
	case ToneStyleDefault, ToneStylePlain, ToneStyleNumber, ToneStyleMark:
		return nil
	default:
		return fmt.Errorf("%w: unsupported tone style %q", ErrInvalidConvertRequest, style)
	}
}

func cloneTokens(tokens []Token) []Token {
	if len(tokens) == 0 {
		return nil
	}
	clone := make([]Token, len(tokens))
	for i, token := range tokens {
		clone[i] = token
		clone[i].Syllables = append([]string(nil), token.Syllables...)
		clone[i].Metadata = cloneStringMap(token.Metadata)
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
