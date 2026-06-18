package vref

import (
	"reflect"
	"testing"
)

func TestFacadeReflectionHelpers(t *testing.T) {
	s := &facadeSample{Name: "alice", hidden: "secret"}
	if !HasField(s, "name") || GetFieldValue(s, "name") != "alice" {
		t.Fatal("field facade failed")
	}
	if got := GetFieldValue(s, "hidden"); got != nil {
		t.Fatalf("hidden field should require explicit unsafe opt-in, got %v", got)
	}
	if got := GetFieldValueWithOptions(s, "hidden", WithUnsafeAccess(true)); got != "secret" {
		t.Fatalf("hidden field with unsafe opt-in = %v", got)
	}
	if err := SetFieldValue(s, "Name", "bob"); err != nil || s.Name != "bob" {
		t.Fatalf("SetFieldValue facade = %v name=%s", err, s.Name)
	}
	if err := SetFieldValueWithOptions(s, "hidden", "changed", WithUnsafeAccess(true)); err != nil || s.hidden != "changed" {
		t.Fatalf("SetFieldValue hidden with unsafe opt-in = %v hidden=%s", err, s.hidden)
	}
	if _, ok := GetMethod(s, false, "Add", reflect.TypeOf(1), reflect.TypeOf(2)); !ok {
		t.Fatal("method facade failed")
	}
	got, err := Invoke(s, "Add", 2, 3)
	if err != nil || got != 5 {
		t.Fatalf("Invoke facade = %v, %v", got, err)
	}
	if TypeOf(s).Kind() != reflect.Pointer || IndirectType(TypeOf(s)).Name() != "facadeSample" {
		t.Fatal("type facade failed")
	}
}

func TestFacadeFieldHelpers(t *testing.T) {
	field, ok := reflect.TypeOf(facadeExtendedSample{}).FieldByName("Alias")
	if !ok {
		t.Fatal("Alias field missing")
	}
	if got := GetFieldName(field); got != "alias" {
		t.Fatalf("GetFieldName = %q", got)
	}
	target := &facadeExtendedSample{facadeEmbedded: facadeEmbedded{Code: "E"}, Alias: "named", Count: 3}
	if !HasField(target, "code") || !HasField(target, "alias") {
		t.Fatal("HasField did not match tag aliases")
	}
	if got := GetField(target, "alias"); got.Name != "Alias" {
		t.Fatalf("GetField(alias) = %#v", got)
	}
	if got := GetFieldMap(target); got["Alias"].Name != "Alias" || got["facadeEmbedded"].Name != "facadeEmbedded" {
		t.Fatalf("GetFieldMap = %#v", got)
	}
	filteredFields := GetFields(target, func(field reflect.StructField) bool { return field.Name == "Count" })
	if len(filteredFields) != 1 || filteredFields[0].Name != "Count" {
		t.Fatalf("GetFields filtered = %#v", filteredFields)
	}
	if got := GetFieldsDirectly(target, true); len(got) < 4 {
		t.Fatalf("GetFieldsDirectly embedded len = %d", len(got))
	}
	if got := GetPublicFieldNames(target); !reflect.DeepEqual(got, []string{"Alias", "Count"}) {
		t.Fatalf("GetPublicFieldNames = %#v", got)
	}
	if got := GetFieldsValueWithOptions(target, []FieldAccessOption{WithAllowUnexported(true)}, func(field reflect.StructField) bool { return field.Name == "Count" }); len(got) != 1 || got[0] != 3 {
		t.Fatalf("GetFieldsValueWithOptions = %#v", got)
	}
	if got := GetStaticFieldValue("static"); got != "static" {
		t.Fatalf("GetStaticFieldValue = %v", got)
	}
	if IsOuterClassField(field) {
		t.Fatal("IsOuterClassField = true, want false")
	}
}

func TestFacadeGetFieldsValue(t *testing.T) {
	s := &facadeSample{Name: "alice"}
	vals := GetFieldsValue(s)
	if len(vals) == 0 || vals[0] != "alice" {
		t.Fatalf("GetFieldsValue = %#v", vals)
	}
}

func TestFacadeGetMethods(t *testing.T) {
	methods := GetMethods(&facadeSample{})
	if len(methods) == 0 {
		t.Fatal("GetMethods returned empty")
	}
}
