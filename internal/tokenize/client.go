package tokenize

import (
	"context"
	"errors"
	"fmt"
)

// ErrMissingProvider reports that a tokenization call has no configured provider.
var ErrMissingProvider = errors.New("tokenize: missing provider")

// Provider performs provider-specific tokenization operations for validated requests.
type Provider interface {
	Tokenize(ctx context.Context, request TokenizeRequest) (TokenizeResponse, error)
	Keywords(ctx context.Context, request KeywordsRequest) (KeywordsResponse, error)
}

// Client routes validated tokenization requests to an injected provider.
type Client struct {
	provider Provider
}

// Option customizes a Client.
type Option func(*Client)

// WithProvider sets the provider used by tokenization and keyword calls.
func WithProvider(provider Provider) Option {
	return func(c *Client) {
		if provider != nil {
			c.provider = provider
		}
	}
}

// New returns a provider-neutral tokenization client.
func New(opts ...Option) *Client {
	client := &Client{}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

// Tokenize validates request and delegates to the configured provider.
func (c *Client) Tokenize(ctx context.Context, request TokenizeRequest) (TokenizeResponse, error) {
	if err := request.Validate(); err != nil {
		return TokenizeResponse{}, err
	}
	if c == nil || c.provider == nil {
		return TokenizeResponse{}, ErrMissingProvider
	}
	response, err := c.provider.Tokenize(ctx, request.Clone())
	if err != nil {
		return TokenizeResponse{}, fmt.Errorf("tokenize provider: %w", err)
	}
	if request.MaxTokens > 0 && len(response.Tokens) > request.MaxTokens {
		return TokenizeResponse{}, fmt.Errorf("%w: provider returned %d tokens, max %d", ErrTokenLimitExceeded, len(response.Tokens), request.MaxTokens)
	}
	return response.Clone(), nil
}

// Keywords validates request and delegates to the configured provider.
func (c *Client) Keywords(ctx context.Context, request KeywordsRequest) (KeywordsResponse, error) {
	if err := request.Validate(); err != nil {
		return KeywordsResponse{}, err
	}
	if c == nil || c.provider == nil {
		return KeywordsResponse{}, ErrMissingProvider
	}
	response, err := c.provider.Keywords(ctx, request.Clone())
	if err != nil {
		return KeywordsResponse{}, fmt.Errorf("tokenize keywords provider: %w", err)
	}
	return response.Clone(), nil
}
