package vjson_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vjson"
)

func TestFacadeUsesNamesWithoutJSONPrefix(t *testing.T) {
	obj := vjson.NewObject().
		Set("name", "knifer-go").
		Set("tags", []string{"go", "tool"})

	if got := obj.GetString("name"); got != "knifer-go" {
		t.Fatalf("GetString(name) = %q", got)
	}

	var parsed *vjson.Object
	parsed, err := vjson.ParseObj(obj.ToString())
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.GetString("name"); got != "knifer-go" {
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

func arrayString(arr *vjson.Array, index int) string {
	return arr.GetString(index)
}
