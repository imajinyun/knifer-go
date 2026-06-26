package vjson_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vjson"
)

func TestFacadeConversionNamesWithoutJSONPrefix(t *testing.T) {
	type user struct {
		Name string `json:"name"`
	}

	var u user
	if err := vjson.ToBean(`{"name":"knifer-go"}`, &u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "knifer-go" {
		t.Fatalf("ToBean() name = %q", u.Name)
	}

	var list []user
	if err := vjson.ToList(`[{"name":"go"},{"name":"tool"}]`, &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[1].Name != "tool" {
		t.Fatalf("ToList() = %#v", list)
	}

	xmlObj, err := vjson.XMLToJSON(`<user><name>knifer-go</name></user>`)
	if err != nil {
		t.Fatal(err)
	}
	if got := objectString(xmlObj, "user", "name"); got != "knifer-go" {
		t.Fatalf("XMLToJSON() user.name = %q", got)
	}

	xmlStr, err := vjson.ToXML(vjson.NewObject().Set("name", "knifer-go"), "user")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xmlStr, "<name>knifer-go</name>") {
		t.Fatalf("ToXML() = %q", xmlStr)
	}
}

func objectString(obj *vjson.Object, objectKey, valueKey string) string {
	return obj.GetJSONObject(objectKey).GetString(valueKey)
}

func TestDynamicJSONContractMatrix(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want any
	}{
		{name: "nil becomes JSON null singleton", src: nil, want: vjson.Null},
		{name: "object string parses to object", src: `{"n":1}`, want: "1"},
		{name: "array bytes parse to array", src: []byte(`["go","knifer"]`), want: "knifer"},
		{name: "json number survives wrapping", src: map[string]any{"n": json.Number("42")}, want: json.Number("42")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vjson.Parse(tt.src)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			switch tt.name {
			case "nil becomes JSON null singleton":
				if got != tt.want {
					t.Fatalf("Parse(nil) = %#v, want JSON null", got)
				}
			case "object string parses to object":
				obj, ok := got.(*vjson.Object)
				value, exists := obj.Get("n")
				if !ok || !exists || fmt.Sprint(value) != tt.want {
					t.Fatalf("Parse(object) = %#v", got)
				}
			case "array bytes parse to array":
				arr, ok := got.(*vjson.Array)
				if !ok || arr.GetString(1) != tt.want {
					t.Fatalf("Parse(array) = %#v", got)
				}
			case "json number survives wrapping":
				obj, ok := got.(*vjson.Object)
				value, exists := obj.Get("n")
				if !ok || !exists || fmt.Sprint(value) != fmt.Sprint(tt.want) {
					t.Fatalf("Parse(map) = %#v", got)
				}
			}
		})
	}

	if _, err := vjson.ParseObj(`[1,2,3]`); err == nil {
		t.Fatal("ParseObj(array) error = nil")
	}
	if got := vjson.GetByPathOr(map[string]any{"user": map[string]any{"name": "go"}}, "user.missing", "fallback"); got != "fallback" {
		t.Fatalf("GetByPathOr missing = %#v", got)
	}
}

func FuzzDynamicJSONStringContract(f *testing.F) {
	f.Add(`{"k":"v"}`)
	f.Add(`[1,2,3]`)
	f.Add(`null`)
	f.Fuzz(func(t *testing.T, input string) {
		parsed, err := vjson.Parse(input)
		if err != nil {
			return
		}
		encoded, err := vjson.ToStr(parsed)
		if err != nil {
			t.Fatalf("ToStr(Parse(input)) error = %v", err)
		}
		if !vjson.IsJSON(encoded) {
			t.Fatalf("ToStr(Parse(input)) = %q, want valid JSON", encoded)
		}
	})
}
