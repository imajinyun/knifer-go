package ref

import (
	"context"
	"reflect"
	"testing"
	"unsafe"
)

type embeddedSample struct {
	Base string `json:"base"`
}

type sample struct {
	embeddedSample
	Name   string `json:"name"`
	Age    int
	hidden string
}

func (s sample) GetName() string         { return s.Name }
func (s *sample) SetName(name string)    { s.Name = name }
func (s sample) Add(a int, b int) int    { return a + b }
func (s sample) String() string          { return s.Name }
func (s sample) Equal(other sample) bool { return s.Name == other.Name }
func (s sample) HashCode() int           { return len(s.Name) }

func newSample(name string, age int) sample { return sample{Name: name, Age: age} }

type sampleError struct{}

func (sampleError) Error() string { return "sample" }

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

func TestAdditionalTypeClassificationHelpers(t *testing.T) {
	var nilSlice []string
	var nilUnsafePointer unsafe.Pointer
	if !IsNilValue(reflect.Value{}) || !IsNilValue(reflect.ValueOf(nilSlice)) || !IsNilValue(reflect.ValueOf(nilUnsafePointer)) {
		t.Fatal("IsNilValue did not treat invalid or nil-able nil values as nil")
	}
	if IsNilValue(reflect.ValueOf(1)) {
		t.Fatal("IsNilValue returned true for non-nil int")
	}

	tests := []struct {
		name       string
		typ        reflect.Type
		funcType   bool
		rangeable  bool
		collection bool
		sliceType  bool
		arrayType  bool
		mapType    bool
	}{
		{name: "nil"},
		{name: "function", typ: reflect.TypeOf(func() {}), funcType: true},
		{name: "slice", typ: reflect.TypeOf([]int{}), rangeable: true, collection: true, sliceType: true},
		{name: "array", typ: reflect.TypeOf([1]int{}), rangeable: true, collection: true, arrayType: true},
		{name: "map", typ: reflect.TypeOf(map[string]int{}), rangeable: true, mapType: true},
		{name: "string", typ: reflect.TypeOf("value")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsFuncType(tt.typ) != tt.funcType {
				t.Fatalf("IsFuncType(%v) = %v", tt.typ, !tt.funcType)
			}
			if IsRangeableType(tt.typ) != tt.rangeable {
				t.Fatalf("IsRangeableType(%v) = %v", tt.typ, !tt.rangeable)
			}
			if IsCollectionType(tt.typ) != tt.collection {
				t.Fatalf("IsCollectionType(%v) = %v", tt.typ, !tt.collection)
			}
			if IsSliceType(tt.typ) != tt.sliceType {
				t.Fatalf("IsSliceType(%v) = %v", tt.typ, !tt.sliceType)
			}
			if IsArrayType(tt.typ) != tt.arrayType {
				t.Fatalf("IsArrayType(%v) = %v", tt.typ, !tt.arrayType)
			}
			if IsMapType(tt.typ) != tt.mapType {
				t.Fatalf("IsMapType(%v) = %v", tt.typ, !tt.mapType)
			}
		})
	}

	if !ImplementsError(reflect.TypeOf(sampleError{})) || ImplementsError(nil) || ImplementsError(reflect.TypeOf("value")) {
		t.Fatal("ImplementsError returned unexpected result")
	}
	if !ImplementsContext(reflect.TypeOf(context.Background())) || ImplementsContext(nil) || ImplementsContext(reflect.TypeOf(sampleError{})) {
		t.Fatal("ImplementsContext returned unexpected result")
	}
	if got := GetPublicFieldNames(sample{}); !reflect.DeepEqual(got, []string{"Name", "Age"}) {
		t.Fatalf("GetPublicFieldNames = %#v", got)
	}
	if got := GetPublicFieldNames((*sample)(nil)); !reflect.DeepEqual(got, []string{"Name", "Age"}) {
		t.Fatalf("GetPublicFieldNames pointer = %#v", got)
	}
	if got := GetPublicFieldNames(123); got != nil {
		t.Fatalf("GetPublicFieldNames non-struct = %#v", got)
	}
}

func TestMethodHelpersAndInvoke(t *testing.T) {
	s := &sample{Name: "alice"}
	if names := GetPublicMethodNames(s); !containsString(names, "GetName") || !containsString(names, "SetName") {
		t.Fatalf("method names = %v", names)
	}
	if methods := GetPublicMethods(s, func(m reflect.Method) bool { return m.Name == "Add" }); len(methods) != 1 {
		t.Fatalf("filtered methods = %v", methods)
	}
	if _, ok := GetPublicMethod(s, "Add", reflect.TypeOf(1), reflect.TypeOf(2)); !ok {
		t.Fatal("GetPublicMethod failed")
	}
	if _, ok := GetMethodOfObj(s, "Add", 1, 2); !ok {
		t.Fatal("GetMethodOfObj failed")
	}
	if _, ok := GetMethodIgnoreCase(s, "getname"); !ok {
		t.Fatal("GetMethodIgnoreCase failed")
	}
	if _, ok := GetMethodByName(s, "String"); !ok {
		t.Fatal("GetMethodByName failed")
	}
	if _, ok := GetMethodByNameIgnoreCase(s, "string"); !ok {
		t.Fatal("GetMethodByNameIgnoreCase failed")
	}
	if len(GetMethods(s)) == 0 || len(GetMethodsDirectly(s, true, true)) == 0 {
		t.Fatal("GetMethods failed")
	}
	stringMethod, _ := GetMethodByName(s, "String")
	equalMethod, _ := GetMethodByName(s, "Equal")
	hashMethod, _ := GetMethodByName(s, "HashCode")
	getMethod, _ := GetMethodByName(s, "GetName")
	setMethod, _ := GetMethodByName(s, "SetName")
	if !IsToStringMethod(stringMethod) || !IsEqualsMethod(equalMethod) || !IsHashCodeMethod(hashMethod) || !IsEmptyParam(getMethod) {
		t.Fatal("method classification failed")
	}
	if !IsGetterOrSetter(getMethod, false) || !IsGetterOrSetterIgnoreCase(setMethod) {
		t.Fatal("getter/setter classification failed")
	}
	got, err := Invoke(s, "Add", int8(1), int8(2))
	if err != nil || got != 3 {
		t.Fatalf("Invoke Add = %v, %v", got, err)
	}
	if _, err := Invoke(s, "Missing"); err == nil {
		t.Fatal("Invoke missing should fail")
	}
	if got, err := InvokeStatic(func(a int, b int) int { return a * b }, 2, 3); err != nil || got != 6 {
		t.Fatalf("InvokeStatic = %v, %v", got, err)
	}
	if got, err := InvokeFunc(func() (int, string) { return 1, "a" }); err != nil || !reflect.DeepEqual(got, []any{1, "a"}) {
		t.Fatalf("InvokeFunc multi return = %#v, %v", got, err)
	}
	if _, err := InvokeRaw(123); err == nil {
		t.Fatal("InvokeRaw non-func should fail")
	}
}

func TestNewInstanceAndConstructorHelpers(t *testing.T) {
	ctor := GetConstructor(newSample)
	if !ctor.IsValid() || len(GetConstructors(newSample)) != 1 || len(GetConstructorsDirectly(newSample)) != 1 {
		t.Fatal("constructor helpers failed")
	}
	created, err := NewInstance(newSample, "alice", 20)
	if err != nil || created.(sample).Name != "alice" || created.(sample).Age != 20 {
		t.Fatalf("NewInstance constructor = %#v, %v", created, err)
	}
	zero, err := NewInstance(reflect.TypeOf(sample{}))
	if err != nil || zero.(sample).Name != "" {
		t.Fatalf("NewInstance type = %#v, %v", zero, err)
	}
	ptr, err := NewInstance(reflect.TypeOf(&sample{}))
	if err != nil || reflect.TypeOf(ptr).String() != "*ref.sample" {
		t.Fatalf("NewInstance pointer type = %#v, %v", ptr, err)
	}
	if NewInstanceIfPossible(nil) != nil {
		t.Fatal("NewInstanceIfPossible nil should be nil")
	}
	if SetAccessible(1) != 1 {
		t.Fatal("SetAccessible should return input")
	}
	RemoveFinalModify(nil)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
