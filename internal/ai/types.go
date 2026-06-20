package ai

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidChatRequest reports a malformed chat request.
	ErrInvalidChatRequest = errors.New("ai: invalid chat request")
	// ErrInvalidEmbeddingRequest reports a malformed embedding request.
	ErrInvalidEmbeddingRequest = errors.New("ai: invalid embedding request")
)

// Role identifies the speaker or source of a chat message.
type Role string

const (
	// RoleSystem marks instructions that steer model behavior.
	RoleSystem Role = "system"
	// RoleUser marks user-provided input.
	RoleUser Role = "user"
	// RoleAssistant marks model-generated output.
	RoleAssistant Role = "assistant"
)

// Message contains one chat message.
type Message struct {
	Role    Role
	Content string
}

// Usage contains provider token accounting when available.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// ProviderMetadata carries low-cardinality provider information.
type ProviderMetadata struct {
	Name    string
	Model   string
	TraceID string
}

// ChatRequest contains provider-neutral chat input.
type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature float64
	Metadata    map[string]string
}

// ChatResponse contains provider-neutral chat output.
type ChatResponse struct {
	Message  Message
	Usage    Usage
	Provider ProviderMetadata
}

// EmbeddingRequest contains provider-neutral embedding input.
type EmbeddingRequest struct {
	Model    string
	Input    []string
	Metadata map[string]string
}

// EmbeddingResponse contains embedding vectors aligned with request input.
type EmbeddingResponse struct {
	Vectors  [][]float32
	Usage    Usage
	Provider ProviderMetadata
}

// Validate checks whether r has the required chat fields.
func (r ChatRequest) Validate() error {
	if strings.TrimSpace(r.Model) == "" {
		return fmt.Errorf("%w: model is required", ErrInvalidChatRequest)
	}
	if len(r.Messages) == 0 {
		return fmt.Errorf("%w: at least one message is required", ErrInvalidChatRequest)
	}
	for i, msg := range r.Messages {
		if strings.TrimSpace(string(msg.Role)) == "" {
			return fmt.Errorf("%w: message %d role is required", ErrInvalidChatRequest, i)
		}
		if strings.TrimSpace(msg.Content) == "" {
			return fmt.Errorf("%w: message %d content is required", ErrInvalidChatRequest, i)
		}
	}
	return nil
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r ChatRequest) Clone() ChatRequest {
	return ChatRequest{
		Model:       r.Model,
		Messages:    cloneMessages(r.Messages),
		Temperature: r.Temperature,
		Metadata:    cloneStringMap(r.Metadata),
	}
}

// Clone returns a response copy that callers can mutate independently.
func (r ChatResponse) Clone() ChatResponse {
	return r
}

// Validate checks whether r has the required embedding fields.
func (r EmbeddingRequest) Validate() error {
	if strings.TrimSpace(r.Model) == "" {
		return fmt.Errorf("%w: model is required", ErrInvalidEmbeddingRequest)
	}
	if len(r.Input) == 0 {
		return fmt.Errorf("%w: at least one input is required", ErrInvalidEmbeddingRequest)
	}
	for i, input := range r.Input {
		if strings.TrimSpace(input) == "" {
			return fmt.Errorf("%w: input %d is empty", ErrInvalidEmbeddingRequest, i)
		}
	}
	return nil
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r EmbeddingRequest) Clone() EmbeddingRequest {
	return EmbeddingRequest{
		Model:    r.Model,
		Input:    append([]string(nil), r.Input...),
		Metadata: cloneStringMap(r.Metadata),
	}
}

// Clone returns a response copy that callers can mutate independently.
func (r EmbeddingResponse) Clone() EmbeddingResponse {
	return EmbeddingResponse{
		Vectors:  cloneVectors(r.Vectors),
		Usage:    r.Usage,
		Provider: r.Provider,
	}
}

func cloneMessages(messages []Message) []Message {
	if len(messages) == 0 {
		return nil
	}
	return append([]Message(nil), messages...)
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

func cloneVectors(vectors [][]float32) [][]float32 {
	if len(vectors) == 0 {
		return nil
	}
	clone := make([][]float32, len(vectors))
	for i := range vectors {
		clone[i] = append([]float32(nil), vectors[i]...)
	}
	return clone
}
