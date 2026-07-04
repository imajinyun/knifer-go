package pinyin

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	convertRequests  []ConvertRequest
	initialsRequests []InitialsRequest
	convertResponse  ConvertResponse
	initialsResponse InitialsResponse
	err              error
}

func (p *fakeProvider) Convert(ctx context.Context, request ConvertRequest) (ConvertResponse, error) {
	select {
	case <-ctx.Done():
		return ConvertResponse{}, ctx.Err()
	default:
	}
	p.convertRequests = append(p.convertRequests, request)
	return p.convertResponse, p.err
}

func (p *fakeProvider) Initials(ctx context.Context, request InitialsRequest) (InitialsResponse, error) {
	select {
	case <-ctx.Done():
		return InitialsResponse{}, ctx.Err()
	default:
	}
	p.initialsRequests = append(p.initialsRequests, request)
	return p.initialsResponse, p.err
}

func TestClientConvertUsesProviderAndClones(t *testing.T) {
	provider := &fakeProvider{convertResponse: ConvertResponse{Text: "中国", Output: "zhong guo", Tokens: []Token{{Text: "中", Syllables: []string{"zhong"}}}}}
	client := New(WithProvider(provider))
	request := ConvertRequest{Text: "中国", Separator: " ", ToneStyle: ToneStylePlain, Metadata: map[string]string{"trace": "one"}}

	response, err := client.Convert(context.Background(), request)
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	request.Metadata["trace"] = "changed"
	provider.convertResponse.Tokens[0].Syllables[0] = "changed"
	if provider.convertRequests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.convertRequests[0])
	}
	if response.Tokens[0].Syllables[0] != "zhong" {
		t.Fatalf("response was not cloned: %+v", response)
	}
}

func TestClientInitialsUsesProviderAndClones(t *testing.T) {
	provider := &fakeProvider{initialsResponse: InitialsResponse{Text: "中国", Output: "zg", Initials: []string{"z", "g"}}}
	client := New(WithProvider(provider))
	request := InitialsRequest{Text: "中国", Metadata: map[string]string{"trace": "one"}}

	response, err := client.Initials(context.Background(), request)
	if err != nil {
		t.Fatalf("Initials returned error: %v", err)
	}
	request.Metadata["trace"] = "changed"
	provider.initialsResponse.Initials[0] = "x"
	if provider.initialsRequests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.initialsRequests[0])
	}
	if response.Initials[0] != "z" {
		t.Fatalf("response was not cloned: %+v", response)
	}
}

func TestNilProviderOptionDoesNotOverwriteConfiguredProvider(t *testing.T) {
	provider := &fakeProvider{convertResponse: ConvertResponse{Text: "中国", Output: "zhong guo"}}
	client := New(WithProvider(provider), WithProvider(nil))
	if _, err := client.Convert(context.Background(), ConvertRequest{Text: "中国"}); err != nil {
		t.Fatalf("Convert with nil overwrite option error = %v", err)
	}
	if len(provider.convertRequests) != 1 {
		t.Fatalf("provider calls = %d, want 1", len(provider.convertRequests))
	}
}

func TestClientRequiresProvider(t *testing.T) {
	client := New()
	if _, err := client.Convert(context.Background(), ConvertRequest{Text: "中国"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("Convert error = %v, want ErrMissingProvider", err)
	}
	if _, err := client.Initials(context.Background(), InitialsRequest{Text: "中国"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("Initials error = %v, want ErrMissingProvider", err)
	}
}

func TestClientValidatesBeforeProvider(t *testing.T) {
	provider := &fakeProvider{}
	client := New(WithProvider(provider))
	_, err := client.Convert(context.Background(), ConvertRequest{})
	if !errors.Is(err, ErrInvalidConvertRequest) {
		t.Fatalf("Convert error = %v, want ErrInvalidConvertRequest", err)
	}
	if len(provider.convertRequests) != 0 {
		t.Fatalf("provider was called for invalid request")
	}
}

func TestClientPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := New(WithProvider(&fakeProvider{}))
	_, err := client.Initials(ctx, InitialsRequest{Text: "中国"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Initials error = %v, want context.Canceled", err)
	}
}

func TestClientWrapsProviderError(t *testing.T) {
	providerErr := errors.New("provider failed")
	client := New(WithProvider(&fakeProvider{err: providerErr}))
	_, err := client.Convert(context.Background(), ConvertRequest{Text: "中国"})
	if !errors.Is(err, providerErr) {
		t.Fatalf("Convert error = %v, want provider error", err)
	}
}
