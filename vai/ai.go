package vai

import (
	"context"

	aiimpl "github.com/imajinyun/knifer-go/internal/ai"
)

// Role identifies the speaker or source of a chat message.
type Role = aiimpl.Role

const (
	// RoleSystem marks instructions that steer model behavior.
	RoleSystem = aiimpl.RoleSystem
	// RoleUser marks user-provided input.
	RoleUser = aiimpl.RoleUser
	// RoleAssistant marks model-generated output.
	RoleAssistant = aiimpl.RoleAssistant
)

// Message contains one chat message.
type Message = aiimpl.Message

// Usage contains provider token accounting when available.
type Usage = aiimpl.Usage

// ProviderMetadata carries low-cardinality provider information.
type ProviderMetadata = aiimpl.ProviderMetadata

// ChatRequest contains provider-neutral chat input.
type ChatRequest = aiimpl.ChatRequest

// ChatResponse contains provider-neutral chat output.
type ChatResponse = aiimpl.ChatResponse

// EmbeddingRequest contains provider-neutral embedding input.
type EmbeddingRequest = aiimpl.EmbeddingRequest

// EmbeddingResponse contains embedding vectors aligned with request input.
type EmbeddingResponse = aiimpl.EmbeddingResponse

// ChatProvider generates chat responses for validated chat requests.
type ChatProvider = aiimpl.ChatProvider

// EmbeddingProvider generates embedding vectors for validated embedding requests.
type EmbeddingProvider = aiimpl.EmbeddingProvider

// Client routes requests to injected AI providers.
type Client = aiimpl.Client

// Option customizes a Client.
type Option = aiimpl.Option

var (
	// ErrInvalidChatRequest reports a malformed chat request.
	ErrInvalidChatRequest = aiimpl.ErrInvalidChatRequest
	// ErrInvalidEmbeddingRequest reports a malformed embedding request.
	ErrInvalidEmbeddingRequest = aiimpl.ErrInvalidEmbeddingRequest
	// ErrMissingChatProvider reports that a chat call has no configured provider.
	ErrMissingChatProvider = aiimpl.ErrMissingChatProvider
	// ErrMissingEmbeddingProvider reports that an embedding call has no configured provider.
	ErrMissingEmbeddingProvider = aiimpl.ErrMissingEmbeddingProvider
)

// WithChatProvider sets the provider used by Client.Chat.
func WithChatProvider(provider ChatProvider) Option { return aiimpl.WithChatProvider(provider) }

// WithEmbeddingProvider sets the provider used by Client.Embed.
func WithEmbeddingProvider(provider EmbeddingProvider) Option {
	return aiimpl.WithEmbeddingProvider(provider)
}

// New returns a provider-neutral AI client.
func New(opts ...Option) *Client { return aiimpl.New(opts...) }

// Chat validates request and delegates to provider through a short-lived Client.
func Chat(ctx context.Context, provider ChatProvider, request ChatRequest) (ChatResponse, error) {
	return New(WithChatProvider(provider)).Chat(ctx, request)
}

// Embed validates request and delegates to provider through a short-lived Client.
func Embed(ctx context.Context, provider EmbeddingProvider, request EmbeddingRequest) (EmbeddingResponse, error) {
	return New(WithEmbeddingProvider(provider)).Embed(ctx, request)
}

// Redact replaces obvious secret-like tokens for examples and diagnostic text.
func Redact(text string) string { return aiimpl.Redact(text) }
