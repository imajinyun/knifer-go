package ai

import (
	"errors"
	"reflect"
	"testing"
)

func TestChatRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request ChatRequest
		wantErr error
	}{
		{
			name:    "missing model",
			request: ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hello"}}},
			wantErr: ErrInvalidChatRequest,
		},
		{
			name:    "missing messages",
			request: ChatRequest{Model: "model-a"},
			wantErr: ErrInvalidChatRequest,
		},
		{
			name:    "message missing role",
			request: ChatRequest{Model: "model-a", Messages: []Message{{Content: "hello"}}},
			wantErr: ErrInvalidChatRequest,
		},
		{
			name:    "message missing content",
			request: ChatRequest{Model: "model-a", Messages: []Message{{Role: RoleUser}}},
			wantErr: ErrInvalidChatRequest,
		},
		{
			name:    "valid minimal request",
			request: ChatRequest{Model: "model-a", Messages: []Message{{Role: RoleUser, Content: "hello"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmbeddingRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request EmbeddingRequest
		wantErr error
	}{
		{
			name:    "missing model",
			request: EmbeddingRequest{Input: []string{"hello"}},
			wantErr: ErrInvalidEmbeddingRequest,
		},
		{
			name:    "missing input",
			request: EmbeddingRequest{Model: "embed-a"},
			wantErr: ErrInvalidEmbeddingRequest,
		},
		{
			name:    "blank input",
			request: EmbeddingRequest{Model: "embed-a", Input: []string{""}},
			wantErr: ErrInvalidEmbeddingRequest,
		},
		{
			name:    "valid request",
			request: EmbeddingRequest{Model: "embed-a", Input: []string{"hello"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestChatRequestCloneDefensivelyCopiesMutableFields(t *testing.T) {
	request := ChatRequest{
		Model:       "model-a",
		Messages:    []Message{{Role: RoleUser, Content: "hello"}},
		Temperature: 0.2,
		Metadata:    map[string]string{"trace": "one"},
	}
	clone := request.Clone()
	request.Messages[0].Content = "changed"
	request.Metadata["trace"] = "changed"

	if clone.Messages[0].Content != "hello" {
		t.Fatalf("cloned messages changed: %#v", clone.Messages)
	}
	if clone.Metadata["trace"] != "one" {
		t.Fatalf("cloned metadata changed: %#v", clone.Metadata)
	}
	if clone.Temperature != 0.2 {
		t.Fatalf("cloned temperature = %v, want 0.2", clone.Temperature)
	}
}

func TestEmbeddingRequestCloneDefensivelyCopiesMutableFields(t *testing.T) {
	request := EmbeddingRequest{
		Model:    "embed-a",
		Input:    []string{"hello"},
		Metadata: map[string]string{"trace": "one"},
	}
	clone := request.Clone()
	request.Input[0] = "changed"
	request.Metadata["trace"] = "changed"

	if clone.Input[0] != "hello" {
		t.Fatalf("cloned input changed: %#v", clone.Input)
	}
	if clone.Metadata["trace"] != "one" {
		t.Fatalf("cloned metadata changed: %#v", clone.Metadata)
	}
}

func TestEmbeddingResponseCloneDefensivelyCopiesVectors(t *testing.T) {
	response := EmbeddingResponse{Vectors: [][]float32{{1, 2}, {3, 4}}}
	clone := response.Clone()
	response.Vectors[0][0] = 99

	want := [][]float32{{1, 2}, {3, 4}}
	if !reflect.DeepEqual(clone.Vectors, want) {
		t.Fatalf("Clone vectors = %#v, want %#v", clone.Vectors, want)
	}
}
