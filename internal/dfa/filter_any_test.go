package dfa

import "testing"

func TestFilterAnyWithMatcherOptions(t *testing.T) {
	type payload struct {
		Text string `json:"text"`
	}
	Init([]string{"global"})
	got, err := FilterAnyWithOptions(payload{Text: "a local"}, true, nil, WithMatcherWords([]string{"local"}))
	if err != nil {
		t.Fatalf("FilterAnyWithOptions: %v", err)
	}
	if got.Text != "a *****" || Contains("local") {
		t.Fatalf("FilterAnyWithOptions = %#v globalContainsLocal=%v", got, Contains("local"))
	}
}

func TestFilterAny(t *testing.T) {
	type payload struct {
		Text string `json:"text"`
		Num  int    `json:"num"`
	}
	Init([]string{"大", "大土豆", "土豆", "刚出锅", "出锅"})
	got, err := FilterAny(payload{Text: sampleText, Num: 100}, true, nil)
	if err != nil {
		t.Fatalf("FilterAny() error = %v", err)
	}
	if got.Text != "我有一颗$****，***的" || got.Num != 100 {
		t.Fatalf("FilterAny() = %#v", got)
	}
}

func TestFilterAnyStringAndRealUserNestedScenario(t *testing.T) {
	Init([]string{"token", "secret"})
	filtered, err := FilterAny("token and secret", true, func(word FoundWord) string {
		return "[redacted:" + word.Word + "]"
	})
	if err != nil {
		t.Fatalf("FilterAny string = %v", err)
	}
	if filtered != "[redacted:token] and [redacted:secret]" {
		t.Fatalf("FilterAny string = %q", filtered)
	}

	type comment struct {
		Body string `json:"body"`
	}
	type payload struct {
		Title    string    `json:"title"`
		Comments []comment `json:"comments"`
	}
	got, err := FilterAny(payload{
		Title: "public token",
		Comments: []comment{
			{Body: "keep"},
			{Body: "secret value"},
		},
	}, true, nil)
	if err != nil {
		t.Fatalf("FilterAny nested payload = %v", err)
	}
	if got.Title != "public *****" || got.Comments[1].Body != "****** value" || got.Comments[0].Body != "keep" {
		t.Fatalf("FilterAny nested payload = %#v", got)
	}
}
