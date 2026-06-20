package vhan

import (
	"context"

	pinyinimpl "github.com/imajinyun/go-knifer/internal/pinyin"
)

// ToneStyle describes how converted syllables should represent tones.
type ToneStyle = pinyinimpl.ToneStyle

const (
	// ToneStyleDefault leaves tone rendering to the configured provider.
	ToneStyleDefault = pinyinimpl.ToneStyleDefault
	// ToneStylePlain requests syllables without tone marks or tone numbers.
	ToneStylePlain = pinyinimpl.ToneStylePlain
	// ToneStyleNumber requests syllables with numeric tone suffixes.
	ToneStyleNumber = pinyinimpl.ToneStyleNumber
	// ToneStyleMark requests syllables with diacritic tone marks.
	ToneStyleMark = pinyinimpl.ToneStyleMark
)

// Token describes one converted input segment and its provider-specific syllables.
type Token = pinyinimpl.Token

// ConvertRequest describes a pinyin conversion request.
type ConvertRequest = pinyinimpl.ConvertRequest

// ConvertResponse contains pinyin conversion output and token metadata.
type ConvertResponse = pinyinimpl.ConvertResponse

// InitialsRequest describes a pinyin initials extraction request.
type InitialsRequest = pinyinimpl.InitialsRequest

// InitialsResponse contains initials extraction output and metadata.
type InitialsResponse = pinyinimpl.InitialsResponse

// Provider performs provider-specific pinyin operations for validated requests.
type Provider = pinyinimpl.Provider

// Client routes validated pinyin requests to an injected provider.
type Client = pinyinimpl.Client

// Option customizes a Client.
type Option = pinyinimpl.Option

var (
	// ErrInvalidConvertRequest reports a malformed pinyin conversion request.
	ErrInvalidConvertRequest = pinyinimpl.ErrInvalidConvertRequest
	// ErrInvalidInitialsRequest reports a malformed pinyin initials request.
	ErrInvalidInitialsRequest = pinyinimpl.ErrInvalidInitialsRequest
	// ErrInputLimitExceeded reports that input text exceeded its configured rune limit.
	ErrInputLimitExceeded = pinyinimpl.ErrInputLimitExceeded
	// ErrMissingProvider reports that a pinyin call has no configured provider.
	ErrMissingProvider = pinyinimpl.ErrMissingProvider
)

// WithProvider sets the provider used by conversion and initials calls.
func WithProvider(provider Provider) Option { return pinyinimpl.WithProvider(provider) }

// New returns a provider-neutral Han text romanization client.
func New(opts ...Option) *Client { return pinyinimpl.New(opts...) }

// Convert validates request and delegates to provider through a short-lived Client.
func Convert(ctx context.Context, provider Provider, request ConvertRequest) (ConvertResponse, error) {
	return New(WithProvider(provider)).Convert(ctx, request)
}

// Initials validates request and delegates to provider through a short-lived Client.
func Initials(ctx context.Context, provider Provider, request InitialsRequest) (InitialsResponse, error) {
	return New(WithProvider(provider)).Initials(ctx, request)
}
