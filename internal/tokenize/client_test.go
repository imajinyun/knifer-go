package tokenize

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	tokenizeRequests []TokenizeRequest
	keywordsRequests []KeywordsRequest
	tokenizeResponse TokenizeResponse
	keywordsResponse KeywordsResponse
	err              error
}

func (p *fakeProvider) Tokenize(ctx context.Context, request TokenizeRequest) (TokenizeResponse, error) {
	select {
	case <-ctx.Done():
		return TokenizeResponse{}, ctx.Err()
	default:
	}
	p.tokenizeRequests = append(p.tokenizeRequests, request)
	return p.tokenizeResponse, p.err
}

func (p *fakeProvider) Keywords(ctx context.Context, request KeywordsRequest) (KeywordsResponse, error) {
	select {
	case <-ctx.Done():
		return KeywordsResponse{}, ctx.Err()
	default:
	}
	p.keywordsRequests = append(p.keywordsRequests, request)
	return p.keywordsResponse, p.err
}

func TestClientTokenizeUsesProviderAndClones(t *testing.T) {
	provider := &fakeProvider{tokenizeResponse: TokenizeResponse{Text: "南京", Tokens: []Token{{Text: "南京", Metadata: map[string]string{"pos": "ns"}}}}}
	client := New(WithProvider(provider))
	request := TokenizeRequest{Text: "南京", Mode: ModePrecise, Metadata: map[string]string{"trace": "one"}}

	response, err := client.Tokenize(context.Background(), request)
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	request.Metadata["trace"] = "changed"
	provider.tokenizeResponse.Tokens[0].Metadata["pos"] = "changed"
	if provider.tokenizeRequests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.tokenizeRequests[0])
	}
	if response.Tokens[0].Metadata["pos"] != "ns" {
		t.Fatalf("response was not cloned: %+v", response)
	}
}

func TestClientKeywordsUsesProviderAndClones(t *testing.T) {
	provider := &fakeProvider{keywordsResponse: KeywordsResponse{Text: "南京", Keywords: []Keyword{{Text: "南京", Score: 0.9, Metadata: map[string]string{"source": "tfidf"}}}}}
	client := New(WithProvider(provider))
	request := KeywordsRequest{Text: "南京", Metadata: map[string]string{"trace": "one"}}

	response, err := client.Keywords(context.Background(), request)
	if err != nil {
		t.Fatalf("Keywords returned error: %v", err)
	}
	request.Metadata["trace"] = "changed"
	provider.keywordsResponse.Keywords[0].Metadata["source"] = "changed"
	if provider.keywordsRequests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.keywordsRequests[0])
	}
	if response.Keywords[0].Metadata["source"] != "tfidf" {
		t.Fatalf("response was not cloned: %+v", response)
	}
}

func TestClientRequiresProvider(t *testing.T) {
	client := New()
	if _, err := client.Tokenize(context.Background(), TokenizeRequest{Text: "南京"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("Tokenize error = %v, want ErrMissingProvider", err)
	}
	if _, err := client.Keywords(context.Background(), KeywordsRequest{Text: "南京"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("Keywords error = %v, want ErrMissingProvider", err)
	}
}

func TestClientValidatesBeforeProvider(t *testing.T) {
	provider := &fakeProvider{}
	client := New(WithProvider(provider))
	_, err := client.Tokenize(context.Background(), TokenizeRequest{})
	if !errors.Is(err, ErrInvalidTokenizeRequest) {
		t.Fatalf("Tokenize error = %v, want ErrInvalidTokenizeRequest", err)
	}
	if len(provider.tokenizeRequests) != 0 {
		t.Fatalf("provider was called for invalid request")
	}
}

func TestClientPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := New(WithProvider(&fakeProvider{}))
	_, err := client.Keywords(ctx, KeywordsRequest{Text: "南京"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Keywords error = %v, want context.Canceled", err)
	}
}

func TestClientWrapsProviderError(t *testing.T) {
	providerErr := errors.New("provider failed")
	client := New(WithProvider(&fakeProvider{err: providerErr}))
	_, err := client.Tokenize(context.Background(), TokenizeRequest{Text: "南京"})
	if !errors.Is(err, providerErr) {
		t.Fatalf("Tokenize error = %v, want provider error", err)
	}
}

func TestClientTokenizeEnforcesTokenLimit(t *testing.T) {
	client := New(WithProvider(&fakeProvider{tokenizeResponse: TokenizeResponse{Tokens: []Token{{Text: "南"}, {Text: "京"}}}}))
	_, err := client.Tokenize(context.Background(), TokenizeRequest{Text: "南京", MaxTokens: 1})
	if !errors.Is(err, ErrTokenLimitExceeded) {
		t.Fatalf("Tokenize error = %v, want ErrTokenLimitExceeded", err)
	}
}
