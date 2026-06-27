package vstr_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vstr"
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

func ExampleContainsAll() {
	fmt.Println(vstr.ContainsAll("go knifer", "go", "knife"))
	fmt.Println(vstr.ContainsAll("go knifer", "go", "rust"))
	// Output:
	// true
	// false
}

func ExampleContainsAny() {
	fmt.Println(vstr.ContainsAny("go knifer", "rust", "knife"))
	fmt.Println(vstr.ContainsAny("go knifer", "rust", "java"))
	// Output:
	// true
	// false
}

func ExampleContainsEmoji() {
	fmt.Println(vstr.ContainsEmoji("go 🚀"))
	fmt.Println(vstr.ContainsEmoji("go"))
	// Output:
	// true
	// false
}

func ExampleContainsEmojiWithOptions() {
	matcher := func(s string) bool { return vstr.Contains(s, ":smile:") }
	fmt.Println(vstr.ContainsEmojiWithOptions("hello :smile:", vstr.WithEmojiMatcher(matcher)))
	fmt.Println(vstr.ContainsEmojiWithOptions("hello", vstr.WithEmojiMatcher(matcher)))
	// Output:
	// true
	// false
}

func ExampleContainsIgnoreCase() {
	fmt.Println(vstr.ContainsIgnoreCase("Go Knifer", "go"))
	fmt.Println(vstr.ContainsIgnoreCase("Go Knifer", "rust"))
	// Output:
	// true
	// false
}

func ExampleDefaultIfBlank() {
	fmt.Println(vstr.DefaultIfBlank("  ", "fallback"))
	fmt.Println(vstr.DefaultIfBlank("go", "fallback"))
	// Output:
	// fallback
	// go
}

func ExampleDefaultIfEmpty() {
	fmt.Printf("%q\n", vstr.DefaultIfEmpty("", "fallback"))
	fmt.Printf("%q\n", vstr.DefaultIfEmpty(" ", "fallback"))
	// Output:
	// "fallback"
	// " "
}

func ExampleEndsWith() {
	fmt.Println(vstr.EndsWith("report.txt", ".txt"))
	fmt.Println(vstr.EndsWith("report.txt", ".md"))
	// Output:
	// true
	// false
}

func ExampleEqualsIgnoreCase() {
	fmt.Println(vstr.EqualsIgnoreCase("Go", "go"))
	fmt.Println(vstr.EqualsIgnoreCase("Go", "Rust"))
	// Output:
	// true
	// false
}

func ExampleEscapeHTML() {
	fmt.Println(vstr.EscapeHTML(`<span title="go">Tom & Jerry's</span>`))
	// Output: &lt;span title=&quot;go&quot;&gt;Tom &amp; Jerry&#39;s&lt;/span&gt;
}

func ExampleHammingDistance64() {
	fmt.Println(vstr.HammingDistance64(0b1010, 0b0011))
	// Output: 2
}

func ExampleHasBlank() {
	fmt.Println(vstr.HasBlank("go", "  "))
	fmt.Println(vstr.HasBlank("go", "knifer"))
	// Output:
	// true
	// false
}

func ExampleHasEmpty() {
	fmt.Println(vstr.HasEmpty("go", ""))
	fmt.Println(vstr.HasEmpty("go", " "))
	// Output:
	// true
	// false
}

func ExampleIsAllBlank() {
	fmt.Println(vstr.IsAllBlank("", "  "))
	fmt.Println(vstr.IsAllBlank("", "go"))
	// Output:
	// true
	// false
}

func ExampleIsAllEmpty() {
	fmt.Println(vstr.IsAllEmpty("", ""))
	fmt.Println(vstr.IsAllEmpty("", " "))
	// Output:
	// true
	// false
}

func ExampleIsAscii() {
	fmt.Println(vstr.IsAscii('A'))
	fmt.Println(vstr.IsAscii('界'))
	// Output:
	// true
	// false
}

func ExampleIsBlankChar() {
	fmt.Println(vstr.IsBlankChar(' '))
	fmt.Println(vstr.IsBlankChar('A'))
	// Output:
	// true
	// false
}

func ExampleIsDigit() {
	fmt.Println(vstr.IsDigit('7'))
	fmt.Println(vstr.IsDigit('七'))
	// Output:
	// true
	// false
}

func ExampleIsEmpty() {
	fmt.Println(vstr.IsEmpty(""))
	fmt.Println(vstr.IsEmpty(" "))
	// Output:
	// true
	// false
}

func ExampleIsLetter() {
	fmt.Println(vstr.IsLetter('G'))
	fmt.Println(vstr.IsLetter('7'))
	// Output:
	// true
	// false
}

func ExampleIsLetterOrDigit() {
	fmt.Println(vstr.IsLetterOrDigit('G'))
	fmt.Println(vstr.IsLetterOrDigit('7'))
	fmt.Println(vstr.IsLetterOrDigit('-'))
	// Output:
	// true
	// true
	// false
}

func ExampleIsNotBlank() {
	fmt.Println(vstr.IsNotBlank("go"))
	fmt.Println(vstr.IsNotBlank("  "))
	// Output:
	// true
	// false
}

func ExampleIsNotEmpty() {
	fmt.Println(vstr.IsNotEmpty("go"))
	fmt.Println(vstr.IsNotEmpty(""))
	// Output:
	// true
	// false
}

func ExampleLevenshteinDistance() {
	fmt.Println(vstr.LevenshteinDistance("kitten", "sitting"))
	// Output: 3
}

func ExampleLevenshteinSimilarity() {
	fmt.Printf("%.2f\n", vstr.LevenshteinSimilarity("kitten", "sitting"))
	// Output: 0.57
}

func ExamplePadLeft() {
	fmt.Println(vstr.PadLeft("go", 5, '.'))
	// Output: ...go
}

func ExamplePadRight() {
	fmt.Println(vstr.PadRight("go", 5, '.'))
	// Output: go...
}

func ExampleRemoveEmoji() {
	fmt.Printf("%q\n", vstr.RemoveEmoji("go 🚀 fast"))
	// Output: "go  fast"
}

func ExampleRemoveEmojiWithOptions() {
	replacer := func(s string) string { return vstr.RemoveSuffix(s, ":smile:") }
	fmt.Printf("%q\n", vstr.RemoveEmojiWithOptions("hello :smile:", vstr.WithEmojiReplacer(replacer)))
	// Output: "hello "
}

func ExampleRemoveSuffix() {
	fmt.Println(vstr.RemoveSuffix("report.txt", ".txt"))
	fmt.Println(vstr.RemoveSuffix("report", ".txt"))
	// Output:
	// report
	// report
}

func ExampleRepeat() {
	fmt.Printf("%q\n", vstr.Repeat("go", 3))
	fmt.Printf("%q\n", vstr.Repeat("go", 0))
	// Output:
	// "gogogo"
	// ""
}

func ExampleRuneLen() {
	fmt.Println(vstr.RuneLen("go语言"))
	// Output: 4
}

func ExampleSimHash() {
	fmt.Println(vstr.SimHash(""))
	// Output: 0
}

func ExampleSplit() {
	fmt.Println(vstr.Split("api,docs,tests", ","))
	fmt.Println(vstr.Split("", ","))
	// Output:
	// [api docs tests]
	// []
}

func ExampleStartsWith() {
	fmt.Println(vstr.StartsWith("/api/users", "/api"))
	fmt.Println(vstr.StartsWith("users", "/api"))
	// Output:
	// true
	// false
}

func ExampleSubAfter() {
	fmt.Println(vstr.SubAfter("api/v1/users", "/", false))
	fmt.Println(vstr.SubAfter("api/v1/users", "/", true))
	// Output:
	// v1/users
	// users
}

func ExampleSubBefore() {
	fmt.Println(vstr.SubBefore("api/v1/users", "/", false))
	fmt.Println(vstr.SubBefore("api/v1/users", "/", true))
	// Output:
	// api
	// api/v1
}

func ExampleToKebabCase() {
	fmt.Println(vstr.ToKebabCase("HelloWorld"))
	// Output: hello-world
}

func ExampleToPascalCase() {
	fmt.Println(vstr.ToPascalCase("hello_world"))
	// Output: HelloWorld
}

func ExampleTrimEnd() {
	fmt.Printf("%q\n", vstr.TrimEnd("  go knifer  "))
	// Output: "  go knifer"
}

func ExampleTrimStart() {
	fmt.Printf("%q\n", vstr.TrimStart("  go knifer  "))
	// Output: "go knifer  "
}

func ExampleTrimToEmpty() {
	fmt.Println(vstr.TrimToEmpty("  go knifer  "))
	// Output: go knifer
}

func ExampleUnescapeHTML() {
	fmt.Println(vstr.UnescapeHTML("&lt;span&gt;Tom &amp; Jerry&lt;/span&gt;"))
	// Output: <span>Tom & Jerry</span>
}

func ExampleUnescapeUnicode() {
	fmt.Println(vstr.UnescapeUnicode(`Hi, \u4E16\u754C`))
	// Output: Hi, 世界
}

func ExampleWithEmojiMatcher() {
	matcher := func(s string) bool { return vstr.Contains(s, ":rocket:") }
	fmt.Println(vstr.ContainsEmojiWithOptions("launch :rocket:", vstr.WithEmojiMatcher(matcher)))
	// Output: true
}

func ExampleWithEmojiReplacer() {
	replacer := func(s string) string { return vstr.RemoveSuffix(s, ":rocket:") }
	fmt.Printf("%q\n", vstr.RemoveEmojiWithOptions("launch :rocket:", vstr.WithEmojiReplacer(replacer)))
	// Output: "launch "
}

func ExampleHasBOM() {
	data := []byte{0xEF, 0xBB, 0xBF, 'g', 'o'}
	fmt.Println(vstr.HasBOM(data))
	fmt.Println(vstr.HasBOM([]byte("go")))
	// Output:
	// UTF-8
	//
}

func ExampleStripBOM() {
	data := []byte{0xEF, 0xBB, 0xBF, 'g', 'o'}
	fmt.Printf("%q\n", vstr.StripBOM(data))
	// Output: "go"
}

func ExampleToUTF8() {
	text, err := vstr.ToUTF8([]byte{0xE9}, "iso-8859-1")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(text))
	// Output: é
}

func ExampleFromUTF8() {
	encoded, err := vstr.FromUTF8([]byte("é"), "iso-8859-1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%X\n", encoded)
	// Output: E9
}
