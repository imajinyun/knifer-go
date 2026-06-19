package vdfa_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdfa"
)

func ExampleContains() {
	vdfa.Init([]string{"bad", "word"})

	fmt.Println(vdfa.Contains("this is a bad word"))
	fmt.Println(vdfa.Contains("clean text"))
	// Output:
	// true
	// false
}

func ExampleFilter() {
	vdfa.Init([]string{"bad"})

	result := vdfa.Filter("this has bad content")
	fmt.Println(result)
	// Output: this has *** content
}

func ExampleFilterWithOptions() {
	result := vdfa.FilterWithOptions(
		"this has bad content",
		vdfa.WithMatcherWords([]string{"bad"}),
	)

	fmt.Println(result)
	// Output: this has *** content
}

func ExampleGetFoundFirstWithOptions() {
	found, ok := vdfa.GetFoundFirstWithOptions(
		"this has a secret",
		vdfa.WithMatcherWords([]string{"secret"}),
	)

	fmt.Println(found.Word, ok)
	// Output: secret true
}

func ExampleContainsAnyWithOptions() {
	value := struct {
		Text string `json:"text"`
	}{Text: "has secret"}

	fmt.Println(vdfa.ContainsAnyWithOptions(
		value,
		vdfa.WithMatcherWords([]string{"secret"}),
	))
	// Output: true
}
