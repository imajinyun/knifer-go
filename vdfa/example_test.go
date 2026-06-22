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

func ExampleNewWordTree() {
	tree := vdfa.NewWordTree()
	tree.AddWords("secret", "classified")

	fmt.Println(tree.IsMatch("contains secret data"))
	fmt.Println(tree.MatchAll("secret and classified"))
	// Output:
	// true
	// [secret classified]
}

func ExampleWordTree_Filter() {
	tree := vdfa.NewWordTree().AddWords("secret")
	masked := tree.Filter("keep secret safe", false, nil)

	fmt.Println(masked)
	// Output: keep ****** safe
}

func ExampleWordTree_MatchWord() {
	tree := vdfa.NewWordTree().AddWord("secret")
	found, ok := tree.MatchWord("a secret appears")

	fmt.Println(found.Word, found.FoundWord, found.Start, found.End, ok)
	// Output: secret secret 2 7 true
}

func ExampleNewWordTreeWithOptions() {
	tree := vdfa.NewWordTreeWithOptions(vdfa.WithCharFilter(func(r rune) bool { return r != '-' }))
	tree.AddWord("ab")

	fmt.Println(tree.IsMatch("a-b"))
	// Output: true
}

func ExampleFilterModeWithOptions() {
	result := vdfa.FilterModeWithOptions(
		"hide secret",
		false,
		func(word vdfa.FoundWord) string { return "[" + word.Word + "]" },
		vdfa.WithMatcherWords([]string{"secret"}),
	)

	fmt.Println(result)
	// Output: hide [secret]
}

func ExampleGetFoundAllWithOptions() {
	found := vdfa.GetFoundAllWithOptions(
		"secret and classified",
		vdfa.WithMatcherWords([]string{"secret", "classified"}),
	)

	for _, word := range found {
		fmt.Println(word.Word)
	}
	// Output:
	// secret
	// classified
}
