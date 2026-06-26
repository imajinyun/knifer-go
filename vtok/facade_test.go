package vtok_test

import (
	"context"
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/vtok"
)

type providerFunc struct {
	tokenize func(context.Context, vtok.TokenizeRequest) (vtok.TokenizeResponse, error)
	keywords func(context.Context, vtok.KeywordsRequest) (vtok.KeywordsResponse, error)
}

func (p providerFunc) Tokenize(ctx context.Context, request vtok.TokenizeRequest) (vtok.TokenizeResponse, error) {
	return p.tokenize(ctx, request)
}

func (p providerFunc) Keywords(ctx context.Context, request vtok.KeywordsRequest) (vtok.KeywordsResponse, error) {
	return p.keywords(ctx, request)
}

func TestTokenizeFacade(t *testing.T) {
	provider := providerFunc{
		tokenize: func(ctx context.Context, request vtok.TokenizeRequest) (vtok.TokenizeResponse, error) {
			return vtok.TokenizeResponse{Text: request.Text, Tokens: []vtok.Token{{Text: "南京"}}}, nil
		},
	}

	response, err := vtok.Tokenize(context.Background(), provider, vtok.TokenizeRequest{Text: "南京"})
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	if len(response.Tokens) != 1 || response.Tokens[0].Text != "南京" {
		t.Fatalf("Tokenize response = %+v", response)
	}
}

func TestKeywordsFacade(t *testing.T) {
	provider := providerFunc{
		keywords: func(ctx context.Context, request vtok.KeywordsRequest) (vtok.KeywordsResponse, error) {
			return vtok.KeywordsResponse{Text: request.Text, Keywords: []vtok.Keyword{{Text: "南京", Score: 0.9}}}, nil
		},
	}

	response, err := vtok.Keywords(context.Background(), provider, vtok.KeywordsRequest{Text: "南京"})
	if err != nil {
		t.Fatalf("Keywords returned error: %v", err)
	}
	if len(response.Keywords) != 1 || response.Keywords[0].Text != "南京" {
		t.Fatalf("Keywords response = %+v", response)
	}
}

func TestFacadeExposesErrors(t *testing.T) {
	_, err := vtok.Tokenize(context.Background(), nil, vtok.TokenizeRequest{Text: "南京"})
	if !errors.Is(err, vtok.ErrMissingProvider) {
		t.Fatalf("Tokenize error = %v, want ErrMissingProvider", err)
	}

	_, err = vtok.Tokenize(context.Background(), providerFunc{}, vtok.TokenizeRequest{Text: "南京", Mode: vtok.Mode("bad")})
	if !errors.Is(err, vtok.ErrInvalidTokenizeRequest) {
		t.Fatalf("Tokenize error = %v, want ErrInvalidTokenizeRequest", err)
	}

	_, err = vtok.Keywords(context.Background(), providerFunc{}, vtok.KeywordsRequest{Text: "南京", MaxInputRunes: 1})
	if !errors.Is(err, vtok.ErrInputLimitExceeded) {
		t.Fatalf("Keywords error = %v, want ErrInputLimitExceeded", err)
	}
}
