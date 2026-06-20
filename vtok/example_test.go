package vtok_test

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vtok"
)

type exampleProvider struct{}

func (exampleProvider) Tokenize(ctx context.Context, request vtok.TokenizeRequest) (vtok.TokenizeResponse, error) {
	return vtok.TokenizeResponse{
		Text: request.Text,
		Tokens: []vtok.Token{
			{Text: "南京", Start: 0, End: 2, Position: 0},
			{Text: "长江大桥", Start: 3, End: 7, Position: 1},
		},
	}, nil
}

func (exampleProvider) Keywords(ctx context.Context, request vtok.KeywordsRequest) (vtok.KeywordsResponse, error) {
	return vtok.KeywordsResponse{
		Text:     request.Text,
		Keywords: []vtok.Keyword{{Text: "南京", Score: 0.9}, {Text: "长江大桥", Score: 0.8}},
	}, nil
}

func ExampleNew() {
	client := vtok.New(vtok.WithProvider(exampleProvider{}))
	response, _ := client.Tokenize(context.Background(), vtok.TokenizeRequest{Text: "南京市长江大桥", Mode: vtok.ModePrecise})
	fmt.Println(response.Tokens[0].Text)
	// Output: 南京
}

func ExampleWithProvider() {
	client := vtok.New(vtok.WithProvider(exampleProvider{}))
	response, _ := client.Keywords(context.Background(), vtok.KeywordsRequest{Text: "南京市长江大桥", Limit: 2})
	fmt.Println(response.Keywords[0].Text)
	// Output: 南京
}

func ExampleTokenize() {
	response, _ := vtok.Tokenize(context.Background(), exampleProvider{}, vtok.TokenizeRequest{Text: "南京市长江大桥", Mode: vtok.ModePrecise})
	fmt.Println(len(response.Tokens), response.Tokens[1].Text)
	// Output: 2 长江大桥
}

func ExampleKeywords() {
	response, _ := vtok.Keywords(context.Background(), exampleProvider{}, vtok.KeywordsRequest{Text: "南京市长江大桥", Limit: 2})
	fmt.Println(len(response.Keywords), response.Keywords[0].Score)
	// Output: 2 0.9
}
