package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrMissingChatProvider reports that a chat call has no configured provider.
	ErrMissingChatProvider = errors.New("ai: missing chat provider")
	// ErrMissingEmbeddingProvider reports that an embedding call has no configured provider.
	ErrMissingEmbeddingProvider = errors.New("ai: missing embedding provider")
)

// ChatProvider generates chat responses for validated chat requests.
type ChatProvider interface {
	Chat(ctx context.Context, request ChatRequest) (ChatResponse, error)
}

// EmbeddingProvider generates embedding vectors for validated embedding requests.
type EmbeddingProvider interface {
	Embed(ctx context.Context, request EmbeddingRequest) (EmbeddingResponse, error)
}

// Client routes requests to injected AI providers.
type Client struct {
	chatProvider      ChatProvider
	embeddingProvider EmbeddingProvider
}

// Option customizes a Client.
type Option func(*Client)

// WithChatProvider sets the provider used by Client.Chat.
func WithChatProvider(provider ChatProvider) Option {
	return func(c *Client) { c.chatProvider = provider }
}

// WithEmbeddingProvider sets the provider used by Client.Embed.
func WithEmbeddingProvider(provider EmbeddingProvider) Option {
	return func(c *Client) { c.embeddingProvider = provider }
}

// New returns a provider-neutral AI client.
func New(opts ...Option) *Client {
	client := &Client{}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

// Chat validates request and delegates to the configured chat provider.
func (c *Client) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	if err := request.Validate(); err != nil {
		return ChatResponse{}, err
	}
	if c == nil || c.chatProvider == nil {
		return ChatResponse{}, ErrMissingChatProvider
	}
	response, err := c.chatProvider.Chat(ctx, request.Clone())
	if err != nil {
		return ChatResponse{}, fmt.Errorf("chat provider: %w", err)
	}
	return response.Clone(), nil
}

// Embed validates request and delegates to the configured embedding provider.
func (c *Client) Embed(ctx context.Context, request EmbeddingRequest) (EmbeddingResponse, error) {
	if err := request.Validate(); err != nil {
		return EmbeddingResponse{}, err
	}
	if c == nil || c.embeddingProvider == nil {
		return EmbeddingResponse{}, ErrMissingEmbeddingProvider
	}
	response, err := c.embeddingProvider.Embed(ctx, request.Clone())
	if err != nil {
		return EmbeddingResponse{}, fmt.Errorf("embedding provider: %w", err)
	}
	return response.Clone(), nil
}

// Redact replaces obvious secret-like tokens for examples and diagnostic text.
func Redact(text string) string {
	fields := strings.Fields(text)
	for i, field := range fields {
		lower := strings.ToLower(strings.Trim(field, "\"'`.,;:()[]{}<>"))
		if lower == "secret" || lower == "password" || strings.HasPrefix(lower, "sk-") {
			fields[i] = "[REDACTED]"
		}
	}
	return strings.Join(fields, " ")
}
