package vjson_test

import (
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vjson"
)

func TestFacadeUsesNamesWithoutJSONPrefix(t *testing.T) {
	obj := vjson.NewObject().
		Set("name", "go-knifer").
		Set("tags", []string{"go", "tool"})

	if got := obj.GetString("name"); got != "go-knifer" {
		t.Fatalf("GetString(name) = %q", got)
	}

	var parsed *vjson.Object
	parsed, err := vjson.ParseObj(obj.ToString())
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.GetString("name"); got != "go-knifer" {
		t.Fatalf("ParseObj().GetString(name) = %q", got)
	}

	arr, err := vjson.ParseArray(`[1,"two",true]`)
	if err != nil {
		t.Fatal(err)
	}
	if got := arrayString(arr, 1); got != "two" {
		t.Fatalf("Array.GetString(1) = %q", got)
	}
}

func TestFacadeHelperNamesWithoutJSONPrefix(t *testing.T) {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	obj := vjson.NewObjectWithConfig(cfg).
		Set("name", "go-knifer").
		Set("empty", vjson.Null)

	if !vjson.IsNull(vjson.Null) {
		t.Fatal("IsNull(Null) = false")
	}
	if !vjson.IsJSON(obj.ToString()) || !vjson.IsObj(obj.ToString()) {
		t.Fatalf("object string should be valid JSON object: %s", obj.ToString())
	}

	formatted := vjson.Format(`{"b":2,"a":1}`)
	if !strings.Contains(formatted, "\n") {
		t.Fatalf("Format() = %q, want pretty JSON", formatted)
	}
	if got := vjson.GetByPath(obj, "name"); got != "go-knifer" {
		t.Fatalf("GetByPath(name) = %v", got)
	}
	if got := vjson.GetByPathOr(obj, "missing", "default"); got != "default" {
		t.Fatalf("GetByPathOr(missing) = %v", got)
	}
	if got := vjson.Quote("a\"b"); got != `"a\"b"` {
		t.Fatalf("Quote() = %q", got)
	}
}

func TestFacadeJSONOptions(t *testing.T) {
	compact, err := vjson.ToStr(map[string]any{"name": "go", "empty": nil}, vjson.WithIgnoreNullValue(true))
	if err != nil {
		t.Fatalf("ToStr with options: %v", err)
	}
	if strings.Contains(compact, "empty") {
		t.Fatalf("ToStr WithIgnoreNullValue should omit null field: %s", compact)
	}
	formatted := vjson.FormatWithOptions(`{"a":1,"b":{"c":2}}`, vjson.WithFormatIndentWidth(2), vjson.WithFormatSpaceAfterKey(false))
	if !strings.Contains(formatted, "\n  \"a\":1") || strings.Contains(formatted, "\": 1") {
		t.Fatalf("FormatWithOptions = %q", formatted)
	}
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	obj, err := vjson.ParseObjWithOptions(map[string]any{"name": "go", "empty": nil}, vjson.WithParseConfig(cfg))
	if err != nil {
		t.Fatalf("ParseObjWithOptions: %v", err)
	}
	if got := obj.ToString(); strings.Contains(got, "empty") {
		t.Fatalf("ParseObjWithOptions should apply config: %s", got)
	}
}

func TestFacadeConversionNamesWithoutJSONPrefix(t *testing.T) {
	type user struct {
		Name string `json:"name"`
	}

	var u user
	if err := vjson.ToBean(`{"name":"go-knifer"}`, &u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "go-knifer" {
		t.Fatalf("ToBean() name = %q", u.Name)
	}

	var list []user
	if err := vjson.ToList(`[{"name":"go"},{"name":"tool"}]`, &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[1].Name != "tool" {
		t.Fatalf("ToList() = %#v", list)
	}

	xmlObj, err := vjson.XMLToJSON(`<user><name>go-knifer</name></user>`)
	if err != nil {
		t.Fatal(err)
	}
	if got := objectString(xmlObj, "user", "name"); got != "go-knifer" {
		t.Fatalf("XMLToJSON() user.name = %q", got)
	}

	xmlStr, err := vjson.ToXML(vjson.NewObject().Set("name", "go-knifer"), "user")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xmlStr, "<name>go-knifer</name>") {
		t.Fatalf("ToXML() = %q", xmlStr)
	}
}

func TestFacadeErrorNameWithoutJSONPrefix(t *testing.T) {
	_, err := vjson.ParseObj(`[1,2]`)
	var jsonErr *vjson.Error
	if !errors.As(err, &jsonErr) {
		t.Fatalf("ParseObj() error type = %T, want *vjson.Error", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}

	_, err = vjson.XMLToJSON(`<root><unclosed></root>`)
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("XMLToJSON malformed XML code = %v, want invalid input", err)
	}
}

func arrayString(arr *vjson.Array, index int) string {
	return arr.GetString(index)
}

func objectString(obj *vjson.Object, objectKey, valueKey string) string {
	return obj.GetJSONObject(objectKey).GetString(valueKey)
}
