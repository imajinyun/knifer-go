package vhan_test

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vhan"
)

type exampleProvider struct{}

func (exampleProvider) Convert(ctx context.Context, request vhan.ConvertRequest) (vhan.ConvertResponse, error) {
	return vhan.ConvertResponse{
		Text:   request.Text,
		Output: "zhong guo",
		Tokens: []vhan.Token{{Text: "中", Syllables: []string{"zhong"}}, {Text: "国", Syllables: []string{"guo"}}},
	}, nil
}

func (exampleProvider) Initials(ctx context.Context, request vhan.InitialsRequest) (vhan.InitialsResponse, error) {
	return vhan.InitialsResponse{Text: request.Text, Output: "zg", Initials: []string{"z", "g"}}, nil
}

func ExampleNew() {
	client := vhan.New(vhan.WithProvider(exampleProvider{}))
	response, _ := client.Convert(context.Background(), vhan.ConvertRequest{Text: "中国", Separator: " ", ToneStyle: vhan.ToneStylePlain})
	fmt.Println(response.Output)
	// Output: zhong guo
}

func ExampleWithProvider() {
	client := vhan.New(vhan.WithProvider(exampleProvider{}))
	response, _ := client.Initials(context.Background(), vhan.InitialsRequest{Text: "中国"})
	fmt.Println(response.Output)
	// Output: zg
}

func ExampleConvert() {
	response, _ := vhan.Convert(context.Background(), exampleProvider{}, vhan.ConvertRequest{Text: "中国", Separator: " ", ToneStyle: vhan.ToneStylePlain})
	fmt.Println(len(response.Tokens), response.Output)
	// Output: 2 zhong guo
}

func ExampleInitials() {
	response, _ := vhan.Initials(context.Background(), exampleProvider{}, vhan.InitialsRequest{Text: "中国"})
	fmt.Println(response.Output, response.Initials[0])
	// Output: zg z
}
