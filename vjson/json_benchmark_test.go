package vjson_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vjson"
)

var benchmarkJSONPayload = `{"user":{"name":"go-knifer","age":5,"tags":["go","tool","json"]}}`

func BenchmarkParseObj(b *testing.B) {
	for b.Loop() {
		obj, err := vjson.ParseObj(benchmarkJSONPayload)
		if err != nil || obj == nil {
			b.Fatalf("ParseObj: %v", err)
		}
	}
}

func BenchmarkToStr(b *testing.B) {
	payload := map[string]any{"name": "go-knifer", "tags": []string{"go", "tool", "json"}}
	for b.Loop() {
		if _, err := vjson.ToStr(payload); err != nil {
			b.Fatalf("ToStr: %v", err)
		}
	}
}

func BenchmarkGetByPath(b *testing.B) {
	root, err := vjson.Parse(benchmarkJSONPayload)
	if err != nil {
		b.Fatalf("Parse: %v", err)
	}
	for b.Loop() {
		if got := vjson.GetByPath(root, "user.name"); got != "go-knifer" {
			b.Fatalf("GetByPath = %v", got)
		}
	}
}

func BenchmarkXMLToJSON(b *testing.B) {
	const payload = `<user><name>go-knifer</name><age>5</age><tag>go</tag><tag>tool</tag></user>`
	for b.Loop() {
		obj, err := vjson.XMLToJSON(payload)
		if err != nil || obj == nil {
			b.Fatalf("XMLToJSON: %v", err)
		}
	}
}
