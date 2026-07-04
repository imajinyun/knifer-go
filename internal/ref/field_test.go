package ref

import (
	"reflect"
	"testing"
)

func TestFieldHelpers(t *testing.T) {
	s := &sample{embeddedSample: embeddedSample{Base: "b"}, Name: "alice", Age: 18, hidden: "secret"}
	if TypeOf(s).Kind() != reflect.Pointer || IndirectType(TypeOf(s)).Name() != "sample" {
		t.Fatal("type helpers failed")
	}
	if !IndirectValue(ValueOf(s)).IsValid() || !IsNil((*sample)(nil)) || IsNil(s) {
		t.Fatal("value/nil helpers failed")
	}
	if !HasField(s, "name") || GetField(s, "missing").Name != "" {
		t.Fatal("field lookup failed")
	}
	field := GetField(s, "name")
	if field.Name != "Name" || GetFieldName(field) != "name" {
		t.Fatalf("field alias failed: %#v", field)
	}
	if GetFieldMap(s)["Age"].Name != "Age" {
		t.Fatal("field map failed")
	}
	if len(GetFields(s)) < 4 || len(GetFieldsDirectly(s, false)) != 4 {
		t.Fatal("fields direct/embedded failed")
	}
	if got := GetFieldValue(s, "hidden"); got != nil {
		t.Fatalf("GetFieldValue hidden without opt-in = %v", got)
	}
	if got := GetFieldValueWithOptions(s, "hidden", WithUnsafeAccess(true)); got != "secret" {
		t.Fatalf("GetFieldValue hidden with opt-in = %v", got)
	}
	if err := SetFieldValue(s, "name", "bob"); err != nil || s.Name != "bob" {
		t.Fatalf("SetFieldValue exported = %v name=%s", err, s.Name)
	}
	if err := SetFieldValue(s, "hidden", "changed"); err == nil || s.hidden != "secret" {
		t.Fatalf("SetFieldValue hidden without opt-in err=%v hidden=%s", err, s.hidden)
	}
	if err := SetFieldValueWithOptions(s, "hidden", "changed", WithUnsafeAccess(true)); err != nil || s.hidden != "changed" {
		t.Fatalf("SetFieldValue hidden with opt-in = %v hidden=%s", err, s.hidden)
	}
	values := GetFieldsValue(s, func(f reflect.StructField) bool { return f.Name == "Age" })
	if len(values) != 1 || values[0] != 18 {
		t.Fatalf("GetFieldsValue = %#v", values)
	}
	if GetStaticFieldValue(123) != 123 || IsOuterClassField(field) {
		t.Fatal("static/outer helpers failed")
	}
}

func TestSetFieldValueRejectsUnsafeNumericConversion(t *testing.T) {
	type target struct {
		Count int8
	}
	dst := &target{}
	if err := SetFieldValue(dst, "Count", int16(128)); err == nil {
		t.Fatal("SetFieldValue overflow error = nil")
	}
	if err := SetFieldValue(dst, "Count", int16(127)); err != nil {
		t.Fatalf("SetFieldValue safe conversion error = %v", err)
	}
	if dst.Count != 127 {
		t.Fatalf("Count = %d, want 127", dst.Count)
	}
}
