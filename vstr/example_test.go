package vstr_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vstr"
)

func ExampleToCamelCase() {
	fmt.Println(vstr.ToCamelCase("hello_world"))
	// Output: helloWorld
}

func ExampleToUnderlineCase() {
	fmt.Println(vstr.ToUnderlineCase("HelloWorld"))
	// Output: hello_world
}

func ExampleIsBlank() {
	fmt.Println(vstr.IsBlank("  "))
	fmt.Println(vstr.IsBlank("go"))
	// Output:
	// true
	// false
}

func ExampleTrim() {
	fmt.Println(vstr.Trim("  go knifer  "))
	// Output: go knifer
}

func ExampleContains() {
	fmt.Println(vstr.Contains("go knifer", "knife"))
	fmt.Println(vstr.ContainsIgnoreCase("Go Knifer", "go"))
	// Output:
	// true
	// true
}

func ExampleSplitTrim() {
	fmt.Println(vstr.SplitTrim(" api, docs, tests ", ","))
	// Output: [api docs tests]
}

func ExampleReverse() {
	fmt.Println(vstr.Reverse("你好"))
	// Output: 好你
}

func ExampleSub() {
	fmt.Println(vstr.Sub("你好世界", 1, 3))
	// Output: 好世
}

func ExampleFormat() {
	fmt.Println(vstr.Format("name={}, age={}", "tom", 12))
	fmt.Println(vstr.Format(`literal \{} stays, value={}`, "ok"))
	// Output:
	// name=tom, age=12
	// literal {} stays, value=ok
}

func ExampleAddPrefixIfNot() {
	fmt.Println(vstr.AddPrefixIfNot("api/users", "/"))
	fmt.Println(vstr.AddPrefixIfNot("/api/users", "/"))
	// Output:
	// /api/users
	// /api/users
}

func ExampleAddSuffixIfNot() {
	fmt.Println(vstr.AddSuffixIfNot("report", ".txt"))
	fmt.Println(vstr.AddSuffixIfNot("report.txt", ".txt"))
	// Output:
	// report.txt
	// report.txt
}

func ExampleRemovePrefix() {
	fmt.Println(vstr.RemovePrefix("/api/users", "/api"))
	fmt.Println(vstr.RemovePrefix("users", "/api"))
	// Output:
	// /users
	// users
}

func ExampleEscapeUnicode() {
	escaped := vstr.EscapeUnicode("Hi, 世界 🌍")
	fmt.Println(escaped)
	fmt.Println(vstr.UnescapeUnicode(escaped))
	// Output:
	// Hi, \u4E16\u754C \uD83C\uDF0D
	// Hi, 世界 🌍
}

func ExampleAntPathMatch() {
	fmt.Println(vstr.AntPathMatch("/api/**/users", "/api/v1/admin/users"))
	fmt.Println(vstr.AntPathMatch("/api/*/users", "/api/v1/admin/users"))
	// Output:
	// true
	// false
}

func ExampleAntPathMatchWithSeparator() {
	fmt.Println(vstr.AntPathMatchWithSeparator("api.**.users", "api.v1.admin.users", "."))
	// Output: true
}

func ExampleJaccardSimilarity() {
	fmt.Printf("%.2f\n", vstr.JaccardSimilarity("night", "nacht"))
	// Output: 0.43
}

func ExampleNGramSimilarity() {
	fmt.Printf("%.2f\n", vstr.NGramSimilarity("night", "nacht", 2))
	// Output: 0.14
}

func ExampleLength() {
	fmt.Println(vstr.Length("go语言"))
	fmt.Println(vstr.RuneLen("go语言"))
	// Output:
	// 4
	// 4
}
