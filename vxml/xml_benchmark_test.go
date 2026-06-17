package vxml

import "testing"

var benchmarkXMLPayload = `<user><name>go-knifer</name><age>5</age><tag>go</tag><tag>xml</tag></user>`

func BenchmarkXMLToMap(b *testing.B) {
	for b.Loop() {
		m, err := XMLToMap(benchmarkXMLPayload)
		if err != nil || m == nil {
			b.Fatalf("XMLToMap: %v", err)
		}
	}
}

func BenchmarkMarshalMap(b *testing.B) {
	payload := map[string]any{"name": "go-knifer", "age": 5, "tag": []string{"go", "xml"}}
	for b.Loop() {
		if _, err := MarshalMap(payload, WithRootName("user"), WithOmitDeclaration(true)); err != nil {
			b.Fatalf("MarshalMap: %v", err)
		}
	}
}

func BenchmarkFormat(b *testing.B) {
	for b.Loop() {
		if _, err := Format(benchmarkXMLPayload); err != nil {
			b.Fatalf("Format: %v", err)
		}
	}
}
