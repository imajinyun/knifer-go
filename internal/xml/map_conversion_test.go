package xml

import (
	"errors"
	"testing"
)

func TestMapConversions(t *testing.T) {
	m, err := XMLToMap(`<root enabled="true"><name>alice</name><age>30</age><score>3.5</score><none>null</none><tags>a</tags><tags>b</tags></root>`)
	if err != nil {
		t.Fatalf("XMLToMap failed: %v", err)
	}
	root := m["root"].(map[string]any)
	if root["enabled"] != true || root["name"] != "alice" || root["age"] != int64(30) || root["score"] != 3.5 || root["none"] != nil {
		t.Fatalf("XMLToMap root = %#v", root)
	}
	if tags, ok := root["tags"].([]any); !ok || len(tags) != 2 {
		t.Fatalf("XMLToMap tags = %#v", root["tags"])
	}
	merged, err := XMLToMapInto(`<x><a>1</a></x>`, map[string]any{"old": true})
	if err != nil || merged["old"] != true || merged["x"] == nil {
		t.Fatalf("XMLToMapInto = %#v, %v", merged, err)
	}
	stripped, err := XMLToMapWithOptions(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`, WithNamespaceAware(false))
	if err != nil || stripped["root"].(map[string]any)["a"] != int64(1) {
		t.Fatalf("XMLToMapWithOptions = %#v, %v", stripped, err)
	}
	limited, err := XMLToMapIntoWithOptions(`<root><a>1</a></root>`, nil, WithMaxBytes(6))
	if err == nil || limited != nil {
		t.Fatalf("XMLToMapIntoWithOptions should fail with max bytes: %#v, %v", limited, err)
	}
	customMap, err := XMLToMapWithOptions(`<root><n>custom-int</n><f>custom-float</f></root>`,
		WithScalarIntParser(func(s string, base, bitSize int) (int64, error) {
			if s == "custom-int" {
				return 99, nil
			}
			return 0, errors.New("not int")
		}),
		WithScalarFloatParser(func(s string, bitSize int) (float64, error) {
			if s == "custom-float" {
				return 6.25, nil
			}
			return 0, errors.New("not float")
		}),
	)
	if err != nil {
		t.Fatalf("XMLToMapWithOptions scalar providers: %v", err)
	}
	customRoot := customMap["root"].(map[string]any)
	if customRoot["n"] != int64(99) || customRoot["f"] != 6.25 {
		t.Fatalf("custom scalar parsers = %#v", customRoot)
	}
	bigMap, err := XMLToMap(`<root><n>9223372036854775808</n></root>`)
	if err != nil {
		t.Fatalf("XMLToMap large scalar: %v", err)
	}
	if got := bigMap["root"].(map[string]any)["n"]; got != uint64(9223372036854775808) {
		t.Fatalf("large scalar = %#v (%T), want exact uint64", got, got)
	}
	if got := XMLNodeToMapInto(nil, nil); len(got) != 0 {
		t.Fatalf("XMLNodeToMapInto nil = %#v", got)
	}

	xmlStr, err := MarshalMap(map[string]any{"name": "alice", "tags": []string{"a", "b"}}, WithRootName("user"), WithOmitDeclaration(true))
	if err != nil || xmlStr != `<user><name>alice</name><tags>a</tags><tags>b</tags></user>` {
		t.Fatalf("MarshalMap = %q, %v", xmlStr, err)
	}
}
