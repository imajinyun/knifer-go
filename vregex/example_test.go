package vregex_test

import (
	"fmt"
	"regexp"

	"github.com/imajinyun/knifer-go/vregex"
)

func ExampleMatch() {
	fmt.Println(vregex.Match(`^\d+$`, "123"))
	fmt.Println(vregex.Match(`^\d+$`, "abc"))
	// Output:
	// true
	// false
}

func ExampleFind() {
	fmt.Println(vregex.Find(`\d+`, "abc123def456"))
	// Output: 123
}

func ExampleFindAll() {
	fmt.Println(vregex.FindAll(`\d+`, "a1b22c333"))
	// Output: [1 22 333]
}

func ExampleCount() {
	fmt.Println(vregex.Count(`\d+`, "a1b22c333"))
	// Output: 3
}

func ExampleContains() {
	fmt.Println(vregex.Contains(`go-\w+`, "hello knifer-go"))
	fmt.Println(vregex.Contains(`go-\w+`, "hello knifer"))
	// Output:
	// true
	// false
}

func ExampleIsMatch() {
	fmt.Println(vregex.IsMatch(`\d+`, "123"))
	fmt.Println(vregex.IsMatch(`\d+`, "abc123"))
	// Output:
	// true
	// false
}

func ExampleIndexOf() {
	match := vregex.IndexOf(`\d+`, "ab12cd34")
	fmt.Println(match.Text, match.Start, match.End)
	// Output: 12 2 4
}

func ExampleLastIndexOf() {
	match := vregex.LastIndexOf(`\d+`, "ab12cd34")
	fmt.Println(match.Text, match.Start, match.End)
	// Output: 34 6 8
}

func ExampleFirst() {
	vregex.First(regexp.MustCompile(`\d+`), "a1b22", func(match vregex.MatchResult) {
		fmt.Println(match.Text)
	})
	// Output: 1
}

func ExampleEach() {
	matches := make([]string, 0)
	vregex.Each(regexp.MustCompile(`\d+`), "a1b22", func(match vregex.MatchResult) {
		matches = append(matches, match.Text)
	})
	fmt.Println(matches)
	// Output: [1 22]
}

func ExampleGetFirstNumber() {
	n, ok := vregex.GetFirstNumber("build-2026-ready")
	fmt.Println(n, ok)
	// Output: 2026 true
}

func ExampleEscape() {
	fmt.Println(vregex.Escape(`a+b?(c)`))
	// Output: a\+b\?\(c\)
}

func ExampleEscapeChar() {
	fmt.Println(vregex.EscapeChar('+'))
	fmt.Println(vregex.EscapeChar('a'))
	// Output:
	// \+
	// a
}

func ExampleReplace() {
	result := vregex.Replace(`\d+`, "abc123def", "X")
	fmt.Println(result)
	// Output: abcXdef
}

func ExampleReplaceFirst() {
	fmt.Println(vregex.ReplaceFirst(`\d+`, "a123b456", "X"))
	// Output: aXb456
}

func ExampleReplaceAll() {
	fmt.Println(vregex.ReplaceAll("name:alice age:20", `(\w+):(\w+)`, `$1=[$2]`))
	// Output: name=[alice] age=[20]
}

func ExampleReplaceAllFunc() {
	result := vregex.ReplaceAllFunc("a1b22", `\d+`, func(match vregex.MatchResult) string {
		return "[" + match.Text + "]"
	})
	fmt.Println(result)
	// Output: a[1]b[22]
}

func ExampleExtractMulti() {
	fmt.Println(vregex.ExtractMulti(`(\d+)年(\d+)月`, "2026年5月", `$1-$2`))
	// Output: 2026-5
}

func ExampleExtractMultiAndDelPre() {
	content := "prefix id=42 tail"
	fmt.Println(vregex.ExtractMultiAndDelPre(`id=(\d+)`, &content, `#$1`))
	fmt.Println(content)
	// Output:
	// #42
	//  tail
}

func ExampleTemplateVars() {
	fmt.Println(vregex.TemplateVars(`$2/$1/$10`))
	// Output: [10 2 1]
}

func ExampleGetByName() {
	fmt.Println(vregex.GetByName(`(?<word>\w+)-(?<num>\d+)`, "item-42", "num"))
	// Output: 42
}

func ExampleGetGroup1() {
	fmt.Println(vregex.GetGroup1(`name=(\w+)`, "name=go"))
	// Output: go
}

func ExampleGetOK() {
	value, ok := vregex.GetOK(`(\w+)-(\d+)`, "item-42", 2)
	fmt.Println(value, ok)
	// Output: 42 true
}

func ExampleGetAllGroups() {
	fmt.Println(vregex.GetAllGroups(`(\w+)-(\d+)`, "item-42 next-7", true, true))
	// Output: [item-42 item 42 next-7 next 7]
}

func ExampleGetAllGroupNames() {
	groups := vregex.GetAllGroupNames(`(?<word>\w+)-(?<num>\d+)`, "item-42")
	fmt.Println(groups["word"], groups["num"])
	// Output: item 42
}

func ExampleFindAllGroup1() {
	fmt.Println(vregex.FindAllGroup1(`name=(\w+)`, "name=alice name=bob"))
	// Output: [alice bob]
}

func ExampleDelFirst() {
	fmt.Println(vregex.DelFirst(`\d+`, "a123b456"))
	// Output: ab456
}

func ExampleDelLast() {
	fmt.Println(vregex.DelLast(`\d+`, "a123b456"))
	// Output: a123b
}

func ExampleDelAll() {
	fmt.Println(vregex.DelAll(`\d+`, "a1b22"))
	// Output: ab
}

func ExampleDelPre() {
	fmt.Println(vregex.DelPre(`:`, "prefix:value"))
	// Output: value
}
