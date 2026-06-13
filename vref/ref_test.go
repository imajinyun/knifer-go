package vref

import (
	"context"
	"reflect"
	"testing"
	"unsafe"
)

type facadeSample struct {
	Name   string `json:"name"`
	hidden string
}

func (s facadeSample) GetName() string      { return s.Name }
func (s facadeSample) Add(a int, b int) int { return a + b }

type facadeEmbedded struct {
	Code string `ref:"code"`
}

type facadeExtendedSample struct {
	facadeEmbedded
	Alias string `xml:"alias"`
	Count int
}

type facadeMethodSample struct{}

func (facadeMethodSample) Equal(facadeMethodSample) bool { return true }
func (facadeMethodSample) HashCode() int                 { return 7 }
func (facadeMethodSample) String() string                { return "method-sample" }
func (facadeMethodSample) SetName(string)                {}

func newFacadeSample(name string) facadeSample { return facadeSample{Name: name} }

type facadeError struct{}

func (facadeError) Error() string { return "facade" }

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

func TestFacadeAdditionalTypeAndFieldHelpers(t *testing.T) {
	var nilSlice []string
	if !IsNil(nilSlice) || IsNil(1) {
		t.Fatal("IsNil facade returned unexpected result")
	}
	var nilUnsafePointer unsafe.Pointer
	if !IsNilValue(reflect.Value{}) || !IsNilValue(reflect.ValueOf(nilSlice)) || !IsNilValue(reflect.ValueOf(nilUnsafePointer)) {
		t.Fatal("IsNilValue facade returned unexpected result")
	}
	if IsNilValue(reflect.ValueOf(1)) {
		t.Fatal("IsNilValue facade returned true for non-nil int")
	}
	if !IsFuncType(reflect.TypeOf(newFacadeSample)) || IsFuncType(nil) {
		t.Fatal("IsFuncType facade returned unexpected result")
	}
	if !IsRangeableType(reflect.TypeOf(map[string]int{})) || !IsRangeableType(reflect.TypeOf([]int{})) || IsRangeableType(reflect.TypeOf(1)) {
		t.Fatal("IsRangeableType facade returned unexpected result")
	}
	if !IsCollectionType(reflect.TypeOf([1]int{})) || !IsCollectionType(reflect.TypeOf([]int{})) || IsCollectionType(reflect.TypeOf(map[string]int{})) {
		t.Fatal("IsCollectionType facade returned unexpected result")
	}
	if !IsSliceType(reflect.TypeOf([]int{})) || !IsArrayType(reflect.TypeOf([1]int{})) || !IsMapType(reflect.TypeOf(map[string]int{})) {
		t.Fatal("specific type predicate facade returned unexpected result")
	}
	if !ImplementsError(reflect.TypeOf(facadeError{})) || ImplementsError(nil) || !ImplementsContext(reflect.TypeOf(context.Background())) || ImplementsContext(nil) {
		t.Fatal("interface implementation facade returned unexpected result")
	}
	if got := ValueOf(nil); got.IsValid() {
		t.Fatalf("ValueOf(nil).IsValid() = true: %v", got)
	}
	if got := IndirectValue(reflect.ValueOf(&facadeSample{Name: "alice"})); !got.IsValid() || got.FieldByName("Name").String() != "alice" {
		t.Fatalf("IndirectValue = %v", got)
	}

	ctor := GetConstructor(newFacadeSample)
	if !ctor.IsValid() || len(GetConstructors(newFacadeSample)) != 1 || len(GetConstructorsDirectly(newFacadeSample)) != 1 {
		t.Fatal("constructor helpers did not expose function target")
	}
	if got := GetConstructor(facadeSample{}); got.IsValid() {
		t.Fatalf("GetConstructor(non-func).IsValid() = true: %v", got)
	}

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

func TestFacadeMethodLookupAndClassifierHelpers(t *testing.T) {
	target := facadeMethodSample{}
	names := GetPublicMethodNames(target)
	if !reflect.DeepEqual(names, []string{"Equal", "HashCode", "SetName", "String"}) {
		t.Fatalf("GetPublicMethodNames = %#v", names)
	}
	methods := GetPublicMethods(target, func(method reflect.Method) bool { return method.Name == "String" })
	if len(methods) != 1 || methods[0].Name != "String" {
		t.Fatalf("GetPublicMethods filtered = %#v", methods)
	}
	if method, ok := GetPublicMethod(target, "SetName", reflect.TypeOf("name")); !ok || method.Name != "SetName" {
		t.Fatalf("GetPublicMethod SetName = %q, %v", method.Name, ok)
	}
	if method, ok := GetMethodIgnoreCase(target, "hashcode"); !ok || method.Name != "HashCode" {
		t.Fatalf("GetMethodIgnoreCase = %q, %v", method.Name, ok)
	}
	if method, ok := GetMethodByName(target, "String"); !ok || !IsToStringMethod(method) || !IsEmptyParam(method) {
		t.Fatalf("GetMethodByName String = %q, %v", method.Name, ok)
	}
	if method, ok := GetMethodByNameIgnoreCase(target, "equal"); !ok || !IsEqualsMethod(method) {
		t.Fatalf("GetMethodByNameIgnoreCase Equal = %q, %v", method.Name, ok)
	}
	if method, ok := GetMethodOfObj(target, "SetName", "bob"); !ok || !IsGetterOrSetter(method, false) || !IsGetterOrSetterIgnoreCase(method) {
		t.Fatalf("GetMethodOfObj SetName = %q, %v", method.Name, ok)
	}
	if method, ok := GetMethodByName(target, "HashCode"); !ok || !IsHashCodeMethod(method) {
		t.Fatalf("GetMethodByName HashCode = %q, %v", method.Name, ok)
	}
	if got := GetMethodNames(target); !reflect.DeepEqual(got, names) {
		t.Fatalf("GetMethodNames = %#v", got)
	}
	if got := GetMethodsDirectly(target, true, true); len(got) != len(names) {
		t.Fatalf("GetMethodsDirectly len = %d, want %d", len(got), len(names))
	}
}

func TestFacadeInstantiationAndInvocationHelpers(t *testing.T) {
	created, err := NewInstance(newFacadeSample, "constructed")
	if err != nil {
		t.Fatalf("NewInstance constructor: %v", err)
	}
	if got, ok := created.(facadeSample); !ok || got.Name != "constructed" {
		t.Fatalf("NewInstance constructor = %#v", created)
	}
	created, err = NewInstance(facadeSample{})
	if err != nil {
		t.Fatalf("NewInstance struct: %v", err)
	}
	if got, ok := created.(facadeSample); !ok || got.Name != "" {
		t.Fatalf("NewInstance struct = %#v", created)
	}
	if got := NewInstanceIfPossible((*facadeSample)(nil)); reflect.TypeOf(got) != reflect.TypeOf(&facadeSample{}) {
		t.Fatalf("NewInstanceIfPossible pointer type = %#v", got)
	}

	s := &facadeSample{Name: "alice"}
	method, ok := GetMethodOfObj(s, "Add", int32(2), int32(3))
	if !ok {
		t.Fatal("GetMethodOfObj Add with convertible args = false")
	}
	if got, err := InvokeWithCheck(s, method, int32(2), int32(3)); err != nil || got != 5 {
		t.Fatalf("InvokeWithCheck = %v, %v", got, err)
	}
	if got, err := InvokeMethod(s, method, 4, 5); err != nil || got != 9 {
		t.Fatalf("InvokeMethod = %v, %v", got, err)
	}
	if got, err := InvokeRaw(func(a, b int) int { return a + b }, 6, 7); err != nil || got != 13 {
		t.Fatalf("InvokeRaw = %v, %v", got, err)
	}
	if got, err := InvokeStatic(func() string { return "static" }); err != nil || got != "static" {
		t.Fatalf("InvokeStatic = %v, %v", got, err)
	}
	if got, err := InvokeFunc(func(a int) (int, string) { return a, "ok" }, 8); err != nil || !reflect.DeepEqual(got, []any{8, "ok"}) {
		t.Fatalf("InvokeFunc multi = %#v, %v", got, err)
	}
	if _, err := InvokeFunc("not-func"); err == nil {
		t.Fatal("InvokeFunc non-func error = nil")
	}
	if _, err := Invoke(s, "Missing"); err == nil {
		t.Fatal("Invoke missing method error = nil")
	}
	if _, err := InvokeWithCheck(s, reflect.Method{}); err == nil {
		t.Fatal("InvokeWithCheck invalid method error = nil")
	}
	if got := SetAccessible(s); got.Name != s.Name {
		t.Fatalf("SetAccessible = %#v", got)
	}
	RemoveFinalModify(&s)
}
