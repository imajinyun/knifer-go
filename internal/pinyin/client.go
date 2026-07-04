package pinyin

import (
	"context"
	"errors"
	"fmt"
)

// ErrMissingProvider reports that a pinyin call has no configured provider.
var ErrMissingProvider = errors.New("pinyin: missing provider")

// Provider performs provider-specific pinyin operations for validated requests.
type Provider interface {
	Convert(ctx context.Context, request ConvertRequest) (ConvertResponse, error)
	Initials(ctx context.Context, request InitialsRequest) (InitialsResponse, error)
}

// Client routes validated pinyin requests to an injected provider.
type Client struct {
	provider Provider
}

// Option customizes a Client.
type Option func(*Client)

// WithProvider sets the provider used by conversion and initials calls.
func WithProvider(provider Provider) Option {
	return func(c *Client) {
		if provider != nil {
			c.provider = provider
		}
	}
}

// New returns a provider-neutral pinyin client.
func New(opts ...Option) *Client {
	client := &Client{}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

// Convert validates request and delegates to the configured provider.
func (c *Client) Convert(ctx context.Context, request ConvertRequest) (ConvertResponse, error) {
	if err := request.Validate(); err != nil {
		return ConvertResponse{}, err
	}
	if c == nil || c.provider == nil {
		return ConvertResponse{}, ErrMissingProvider
	}
	response, err := c.provider.Convert(ctx, request.Clone())
	if err != nil {
		return ConvertResponse{}, fmt.Errorf("pinyin convert provider: %w", err)
	}
	return response.Clone(), nil
}

// Initials validates request and delegates to the configured provider.
func (c *Client) Initials(ctx context.Context, request InitialsRequest) (InitialsResponse, error) {
	if err := request.Validate(); err != nil {
		return InitialsResponse{}, err
	}
	if c == nil || c.provider == nil {
		return InitialsResponse{}, ErrMissingProvider
	}
	response, err := c.provider.Initials(ctx, request.Clone())
	if err != nil {
		return InitialsResponse{}, fmt.Errorf("pinyin initials provider: %w", err)
	}
	return response.Clone(), nil
}
