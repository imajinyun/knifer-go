package vjson_test

import (
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vjson"
)

func TestFacadeHelperNamesWithoutJSONPrefix(t *testing.T) {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	obj := vjson.NewObjectWithConfig(cfg).
		Set("name", "knifer-go").
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
	if got := vjson.GetByPath(obj, "name"); got != "knifer-go" {
		t.Fatalf("GetByPath(name) = %v", got)
	}
	if got := vjson.GetByPathOr(obj, "missing", "default"); got != "default" {
		t.Fatalf("GetByPathOr(missing) = %v", got)
	}
	if got := vjson.Quote("a\"b"); got != `"a\"b"` {
		t.Fatalf("Quote() = %q", got)
	}
}
