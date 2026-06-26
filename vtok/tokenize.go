package vtok

import (
	"context"

	tokenizeimpl "github.com/imajinyun/knifer-go/internal/tokenize"
)

// Mode describes a provider-specific tokenization mode.
type Mode = tokenizeimpl.Mode

const (
	// ModeDefault leaves segmentation behavior to the configured provider.
	ModeDefault = tokenizeimpl.ModeDefault
	// ModePrecise requests precise segmentation.
	ModePrecise = tokenizeimpl.ModePrecise
	// ModeSearch requests search-engine-oriented segmentation.
	ModeSearch = tokenizeimpl.ModeSearch
	// ModeFull requests full segmentation when supported by the provider.
	ModeFull = tokenizeimpl.ModeFull
)

// Token describes one token emitted by a tokenizer provider.
type Token = tokenizeimpl.Token

// Keyword describes one keyword emitted by a keyword extraction provider.
type Keyword = tokenizeimpl.Keyword

// TokenizeRequest describes a tokenization request.
type TokenizeRequest = tokenizeimpl.TokenizeRequest

// TokenizeResponse contains tokenization output and metadata.
type TokenizeResponse = tokenizeimpl.TokenizeResponse

// KeywordsRequest describes a keyword extraction request.
type KeywordsRequest = tokenizeimpl.KeywordsRequest

// KeywordsResponse contains extracted keywords and metadata.
type KeywordsResponse = tokenizeimpl.KeywordsResponse

// Provider performs provider-specific tokenization operations for validated requests.
type Provider = tokenizeimpl.Provider

// Client routes validated tokenization requests to an injected provider.
type Client = tokenizeimpl.Client

// Option customizes a Client.
type Option = tokenizeimpl.Option

var (
	// ErrInvalidTokenizeRequest reports a malformed tokenization request.
	ErrInvalidTokenizeRequest = tokenizeimpl.ErrInvalidTokenizeRequest
	// ErrInvalidKeywordsRequest reports a malformed keyword extraction request.
	ErrInvalidKeywordsRequest = tokenizeimpl.ErrInvalidKeywordsRequest
	// ErrInputLimitExceeded reports that input text exceeded its configured rune limit.
	ErrInputLimitExceeded = tokenizeimpl.ErrInputLimitExceeded
	// ErrTokenLimitExceeded reports that provider output exceeded its configured token limit.
	ErrTokenLimitExceeded = tokenizeimpl.ErrTokenLimitExceeded
	// ErrMissingProvider reports that a tokenization call has no configured provider.
	ErrMissingProvider = tokenizeimpl.ErrMissingProvider
)

// WithProvider sets the provider used by tokenization and keyword calls.
func WithProvider(provider Provider) Option { return tokenizeimpl.WithProvider(provider) }

// New returns a provider-neutral tokenization client.
func New(opts ...Option) *Client { return tokenizeimpl.New(opts...) }

// Tokenize validates request and delegates to provider through a short-lived Client.
func Tokenize(ctx context.Context, provider Provider, request TokenizeRequest) (TokenizeResponse, error) {
	return New(WithProvider(provider)).Tokenize(ctx, request)
}

// Keywords validates request and delegates to provider through a short-lived Client.
func Keywords(ctx context.Context, provider Provider, request KeywordsRequest) (KeywordsResponse, error) {
	return New(WithProvider(provider)).Keywords(ctx, request)
}
