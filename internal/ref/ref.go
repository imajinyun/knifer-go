package ref

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

// FieldFilter filters struct fields.
type FieldFilter func(reflect.StructField) bool

// MethodFilter filters methods.
type MethodFilter func(reflect.Method) bool

type fieldAccessConfig struct {
	unsafeAccess bool
}

// FieldAccessOption customizes field value access and mutation.
type FieldAccessOption func(*fieldAccessConfig)

// WithUnsafeAccess controls whether unexported addressable fields may be accessed via unsafe.
// It is disabled by default; enable it only for trusted in-process values.
func WithUnsafeAccess(enabled bool) FieldAccessOption {
	return func(c *fieldAccessConfig) { c.unsafeAccess = enabled }
}

// WithAllowUnexported controls whether unexported addressable fields may be accessed via unsafe.
// It is disabled by default; enable it only for trusted in-process values.
func WithAllowUnexported(enabled bool) FieldAccessOption { return WithUnsafeAccess(enabled) }

func applyFieldAccessOptions(opts []FieldAccessOption) fieldAccessConfig {
	cfg := fieldAccessConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// TypeOf returns the non-nil reflection type of object.
func TypeOf(object any) reflect.Type {
	if IsNil(object) {
		return nil
	}
	return reflect.TypeOf(object)
}

// IndirectType unwraps pointers from typ.
func IndirectType(typ reflect.Type) reflect.Type {
	for typ != nil && typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

// ValueOf returns the reflection value of object.
func ValueOf(object any) reflect.Value {
	if object == nil {
		return reflect.Value{}
	}
	return reflect.ValueOf(object)
}

// IndirectValue unwraps pointers and interfaces from value.
func IndirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

// IsNil reports whether object is nil, including typed nil values.
func IsNil(object any) bool {
	if object == nil {
		return true
	}
	return IsNilValue(reflect.ValueOf(object))
}

// IsNilValue reports whether value is invalid or holds a nil-able nil value.
func IsNilValue(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return value.IsNil()
	default:
		return false
	}
}

// IsFunction reports whether in is a func.
func IsFunction(in any) bool { return isType(in, IsFuncType) }

// IsIteratee reports whether in can be ranged over.
func IsIteratee(in any) bool { return isType(in, IsRangeableType) }

// IsCollection reports whether in is a collection (slice or array).
func IsCollection(in any) bool { return isType(in, IsCollectionType) }

// IsSlice reports whether in is a slice.
func IsSlice(in any) bool { return isType(in, IsSliceType) }

// IsArray reports whether in is an array.
func IsArray(in any) bool { return isType(in, IsArrayType) }

// IsMap reports whether in is a map.
func IsMap(in any) bool { return isType(in, IsMapType) }

// IsFuncType reports whether typ is a function type.
func IsFuncType(typ reflect.Type) bool { return typ != nil && typ.Kind() == reflect.Func }

// IsRangeableType reports whether typ can be ranged over.
func IsRangeableType(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	switch typ.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}

// IsCollectionType reports whether typ is an array or slice type.
func IsCollectionType(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	switch typ.Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

// IsSliceType reports whether typ is a slice type.
func IsSliceType(typ reflect.Type) bool { return typ != nil && typ.Kind() == reflect.Slice }

// IsArrayType reports whether typ is an array type.
func IsArrayType(typ reflect.Type) bool { return typ != nil && typ.Kind() == reflect.Array }

// IsMapType reports whether typ is a map type.
func IsMapType(typ reflect.Type) bool { return typ != nil && typ.Kind() == reflect.Map }

// isType reports whether in is non-nil and its reflect.Type satisfies pred.
func isType(in any, pred func(reflect.Type) bool) bool {
	if in == nil {
		return false
	}
	return pred(reflect.TypeOf(in))
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// ImplementsError reports whether typ implements error.
func ImplementsError(typ reflect.Type) bool { return typ != nil && typ.Implements(errorType) }

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// ImplementsContext reports whether typ implements context.Context.
func ImplementsContext(typ reflect.Type) bool { return typ != nil && typ.Implements(contextType) }

// GetConstructor returns a constructor function when target itself is a function.
func GetConstructor(target any) reflect.Value {
	v := reflect.ValueOf(target)
	if v.IsValid() && v.Kind() == reflect.Func {
		return v
	}
	return reflect.Value{}
}

// GetConstructors returns a constructor function list when target itself is a function.
func GetConstructors(target any) []reflect.Value {
	if ctor := GetConstructor(target); ctor.IsValid() {
		return []reflect.Value{ctor}
	}
	return nil
}

// GetConstructorsDirectly is an alias of GetConstructors.
func GetConstructorsDirectly(target any) []reflect.Value { return GetConstructors(target) }

// HasField reports whether target struct type has a field by Go name or common tag alias.
func HasField(target any, name string) bool { return GetField(target, name).Name != "" }

// GetFieldName returns field alias from ref/json/xml tag, or the Go field name.
func GetFieldName(field reflect.StructField) string {
	for _, tagName := range []string{"ref", "json", "xml"} {
		if tag := field.Tag.Get(tagName); tag != "" {
			name := strings.Split(tag, ",")[0]
			if name != "" && name != "-" {
				return name
			}
		}
	}
	return field.Name
}

// GetField returns the first field matched by Go name or common tag alias.
func GetField(target any, name string) reflect.StructField {
	for _, field := range GetFields(target) {
		if field.Name == name || GetFieldName(field) == name {
			return field
		}
	}
	return reflect.StructField{}
}

// GetFieldMap returns a field name to StructField map.
func GetFieldMap(target any) map[string]reflect.StructField {
	fields := GetFields(target)
	out := make(map[string]reflect.StructField, len(fields))
	for _, field := range fields {
		out[field.Name] = field
	}
	return out
}

// GetFields returns all fields from a struct type and embedded anonymous structs.
func GetFields(target any, filters ...FieldFilter) []reflect.StructField {
	return filterFields(GetFieldsDirectly(target, true), filters...)
}

// GetFieldsDirectly returns struct fields. When withEmbeddedFields is true, anonymous struct fields are expanded.
func GetFieldsDirectly(target any, withEmbeddedFields bool) []reflect.StructField {
	t := typeFrom(target)
	t = IndirectType(t)
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	out := make([]reflect.StructField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		out = append(out, field)
		if withEmbeddedFields && field.Anonymous {
			out = append(out, GetFieldsDirectly(field.Type, true)...)
		}
	}
	return out
}

// GetPublicFieldNames returns exported field names from a struct type.
func GetPublicFieldNames(target any) []string {
	fields := GetFieldsDirectly(target, false)
	if len(fields) == 0 {
		return nil
	}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.IsExported() {
			out = append(out, field.Name)
		}
	}
	return out
}

// GetFieldValue returns a field value by name. Missing or inaccessible fields return nil.
func GetFieldValue(obj any, fieldName string) any {
	return GetFieldValueWithOptions(obj, fieldName)
}

// GetFieldValueWithOptions returns a field value by name using per-call access options.
func GetFieldValueWithOptions(obj any, fieldName string, opts ...FieldAccessOption) any {
	v := fieldValue(obj, fieldName)
	if !v.IsValid() {
		return nil
	}
	value, ok := valueInterface(v, applyFieldAccessOptions(opts))
	if !ok {
		return nil
	}
	return value
}

// GetStaticFieldValue returns the value represented by value.
func GetStaticFieldValue(value any) any { return value }

// GetFieldsValue returns values of all matched fields.
func GetFieldsValue(obj any, filters ...FieldFilter) []any {
	return GetFieldsValueWithOptions(obj, nil, filters...)
}

// GetFieldsValueWithOptions returns values of all matched fields using per-call access options.
func GetFieldsValueWithOptions(obj any, opts []FieldAccessOption, filters ...FieldFilter) []any {
	v := IndirectValue(reflect.ValueOf(obj))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return nil
	}
	cfg := applyFieldAccessOptions(opts)
	fields := GetFields(v.Type(), filters...)
	out := make([]any, 0, len(fields))
	for _, field := range fields {
		fv := fieldByIndex(v, field.Index)
		if fv.IsValid() {
			if value, ok := valueInterface(fv, cfg); ok {
				out = append(out, value)
			}
		}
	}
	return out
}

// SetFieldValue sets a field by Go name or common tag alias.
func SetFieldValue(obj any, fieldName string, value any) error {
	return SetFieldValueWithOptions(obj, fieldName, value)
}

// SetFieldValueWithOptions sets a field by Go name or common tag alias using per-call access options.
func SetFieldValueWithOptions(obj any, fieldName string, value any, opts ...FieldAccessOption) error {
	v := fieldValue(obj, fieldName)
	if !v.IsValid() {
		return fmt.Errorf("field %q not found", fieldName)
	}
	return setValue(v, value, applyFieldAccessOptions(opts))
}

// IsOuterClassField reports false in Go; it is kept as a compatibility guard.
func IsOuterClassField(reflect.StructField) bool { return false }

// GetPublicMethodNames returns exported method names.
func GetPublicMethodNames(target any) []string {
	methods := GetPublicMethods(target)
	out := make([]string, 0, len(methods))
	for _, method := range methods {
		out = append(out, method.Name)
	}
	sort.Strings(out)
	return out
}

// GetPublicMethods returns exported methods after filtering.
func GetPublicMethods(target any, filters ...MethodFilter) []reflect.Method {
	return filterMethods(getMethods(target, true), filters...)
}

// GetPublicMethod returns an exported method by name and optional parameter types.
func GetPublicMethod(target any, methodName string, paramTypes ...reflect.Type) (reflect.Method, bool) {
	return GetMethod(target, false, methodName, paramTypes...)
}

// GetMethodOfObj returns a method by inferring parameter types from args.
func GetMethodOfObj(obj any, methodName string, args ...any) (reflect.Method, bool) {
	paramTypes := make([]reflect.Type, 0, len(args))
	for _, arg := range args {
		if arg == nil {
			paramTypes = append(paramTypes, nil)
		} else {
			paramTypes = append(paramTypes, reflect.TypeOf(arg))
		}
	}
	return GetMethod(obj, false, methodName, paramTypes...)
}

// GetMethodIgnoreCase returns a method by case-insensitive name and optional parameter types.
func GetMethodIgnoreCase(target any, methodName string, paramTypes ...reflect.Type) (reflect.Method, bool) {
	return GetMethod(target, true, methodName, paramTypes...)
}

// GetMethod returns a method by name and optional parameter types.
func GetMethod(target any, ignoreCase bool, methodName string, paramTypes ...reflect.Type) (reflect.Method, bool) {
	for _, method := range GetMethods(target) {
		matched := method.Name == methodName
		if ignoreCase {
			matched = strings.EqualFold(method.Name, methodName)
		}
		if matched && methodParamsAssignable(method.Type, paramTypes) {
			return method, true
		}
	}
	return reflect.Method{}, false
}

// GetMethodByName returns the first method with the provided name.
func GetMethodByName(target any, methodName string) (reflect.Method, bool) {
	return getMethodByName(target, false, methodName)
}

// GetMethodByNameIgnoreCase returns the first method with the provided name, ignoring case.
func GetMethodByNameIgnoreCase(target any, methodName string) (reflect.Method, bool) {
	return getMethodByName(target, true, methodName)
}

// GetMethodNames returns all method names.
func GetMethodNames(target any) []string {
	methods := GetMethods(target)
	out := make([]string, 0, len(methods))
	for _, method := range methods {
		out = append(out, method.Name)
	}
	sort.Strings(out)
	return out
}

// GetMethods returns all exported methods after filtering.
func GetMethods(target any, filters ...MethodFilter) []reflect.Method {
	return filterMethods(getMethods(target, true), filters...)
}

// GetMethodsDirectly returns methods on target. Go reflection exposes exported methods only.
func GetMethodsDirectly(target any, _ bool, _ bool) []reflect.Method { return getMethods(target, true) }

// IsEqualsMethod reports whether method name is Equal or Equals.
func IsEqualsMethod(method reflect.Method) bool {
	return method.Name == "Equal" || method.Name == "Equals"
}

// IsHashCodeMethod reports whether method name is HashCode.
func IsHashCodeMethod(method reflect.Method) bool { return method.Name == "HashCode" }

// IsToStringMethod reports whether method name is String or ToString.
func IsToStringMethod(method reflect.Method) bool {
	return method.Name == "String" || method.Name == "ToString"
}

// IsEmptyParam reports whether method has no non-receiver parameters.
func IsEmptyParam(method reflect.Method) bool { return method.Type.NumIn() <= 1 }

// IsGetterOrSetterIgnoreCase reports whether method name looks like a getter or setter.
func IsGetterOrSetterIgnoreCase(method reflect.Method) bool { return IsGetterOrSetter(method, true) }

// IsGetterOrSetter reports whether method name looks like a getter or setter.
func IsGetterOrSetter(method reflect.Method, ignoreCase bool) bool {
	name := method.Name
	if ignoreCase {
		name = strings.ToLower(name)
		return strings.HasPrefix(name, "get") || strings.HasPrefix(name, "set") || strings.HasPrefix(name, "is")
	}
	return strings.HasPrefix(name, "Get") || strings.HasPrefix(name, "Set") || strings.HasPrefix(name, "Is")
}

// NewInstance creates a new value for target. If target is a constructor function, it is invoked.
func NewInstance(target any, params ...any) (any, error) {
	if ctor := GetConstructor(target); ctor.IsValid() {
		return InvokeFunc(ctor.Interface(), params...)
	}
	t := typeFrom(target)
	if t == nil {
		return nil, errors.New("nil type")
	}
	if t.Kind() == reflect.Pointer {
		return reflect.New(t.Elem()).Interface(), nil
	}
	return reflect.New(t).Elem().Interface(), nil
}

// NewInstanceIfPossible creates a useful zero value when possible.
func NewInstanceIfPossible(target any) any {
	v, err := NewInstance(target)
	if err == nil {
		return v
	}
	t := typeFrom(target)
	if t == nil {
		return nil
	}
	return reflect.Zero(t).Interface()
}

// InvokeStatic invokes a function-like method value.
func InvokeStatic(fn any, args ...any) (any, error) { return InvokeFunc(fn, args...) }

// InvokeWithCheck invokes a method value with argument conversion.
func InvokeWithCheck(obj any, method reflect.Method, args ...any) (any, error) {
	if !method.Func.IsValid() {
		return nil, errors.New("invalid method")
	}
	values := append([]reflect.Value{reflect.ValueOf(obj)}, valuesForCall(method.Type, 1, args)...)
	return call(method.Func, values)
}

// InvokeMethod invokes a reflect method.
func InvokeMethod(obj any, method reflect.Method, args ...any) (any, error) {
	return InvokeWithCheck(obj, method, args...)
}

// InvokeRaw invokes fn without name lookup.
func InvokeRaw(fn any, args ...any) (any, error) { return InvokeFunc(fn, args...) }

// Invoke invokes a method by name on obj.
func Invoke(obj any, methodName string, args ...any) (any, error) {
	method, ok := GetMethodOfObj(obj, methodName, args...)
	if !ok {
		return nil, fmt.Errorf("method %q not found", methodName)
	}
	return InvokeWithCheck(obj, method, args...)
}

// InvokeFunc invokes a function with best-effort argument conversion.
func InvokeFunc(fn any, args ...any) (any, error) {
	v := reflect.ValueOf(fn)
	if !v.IsValid() || v.Kind() != reflect.Func {
		return nil, errors.New("target is not a function")
	}
	return call(v, valuesForCall(v.Type(), 0, args))
}

// SetAccessible returns object unchanged. Go does not expose Java-style access flags.
func SetAccessible[T any](object T) T { return object }

// RemoveFinalModify is a no-op compatibility hook.
func RemoveFinalModify(any) {}

func typeFrom(target any) reflect.Type {
	switch t := target.(type) {
	case nil:
		return nil
	case reflect.Type:
		return t
	case reflect.Value:
		return t.Type()
	default:
		return reflect.TypeOf(target)
	}
}

func filterFields(fields []reflect.StructField, filters ...FieldFilter) []reflect.StructField {
	if len(filters) == 0 || filters[0] == nil {
		return fields
	}
	out := make([]reflect.StructField, 0, len(fields))
	for _, field := range fields {
		if filters[0](field) {
			out = append(out, field)
		}
	}
	return out
}

func filterMethods(methods []reflect.Method, filters ...MethodFilter) []reflect.Method {
	if len(filters) == 0 || filters[0] == nil {
		return methods
	}
	out := make([]reflect.Method, 0, len(methods))
	for _, method := range methods {
		if filters[0](method) {
			out = append(out, method)
		}
	}
	return out
}

func fieldValue(obj any, fieldName string) reflect.Value {
	v := IndirectValue(reflect.ValueOf(obj))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	field := GetField(v.Type(), fieldName)
	if field.Name == "" {
		return reflect.Value{}
	}
	return fieldByIndex(v, field.Index)
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if !v.IsValid() || v.Kind() != reflect.Struct || i >= v.NumField() {
			return reflect.Value{}
		}
		v = v.Field(i)
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
	}
	return v
}

func valueInterface(v reflect.Value, cfg fieldAccessConfig) (any, bool) {
	if !v.IsValid() {
		return nil, false
	}
	if v.CanInterface() {
		return v.Interface(), true
	}
	if !cfg.unsafeAccess || !v.CanAddr() {
		return nil, false
	}
	// #nosec G103 -- reflection helpers intentionally access unexported addressable fields.
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface(), true
}

func setValue(dst reflect.Value, value any, cfg fieldAccessConfig) error {
	if !dst.CanSet() {
		if !cfg.unsafeAccess || !dst.CanAddr() {
			return errors.New("field cannot be set")
		}
		// #nosec G103 -- setter must address unexported fields when caller provides addressable values.
		dst = reflect.NewAt(dst.Type(), unsafe.Pointer(dst.UnsafeAddr())).Elem()
	}
	if value == nil {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	src := reflect.ValueOf(value)
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return nil
	}
	if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return nil
	}
	return fmt.Errorf("cannot assign %s to %s", src.Type(), dst.Type())
}

func getMethods(target any, includePointer bool) []reflect.Method {
	t := typeFrom(target)
	if t == nil {
		return nil
	}
	if includePointer && t.Kind() != reflect.Pointer {
		t = reflect.PointerTo(t)
	}
	out := make([]reflect.Method, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		out = append(out, t.Method(i))
	}
	return out
}

func getMethodByName(target any, ignoreCase bool, methodName string) (reflect.Method, bool) {
	for _, method := range GetMethods(target) {
		if method.Name == methodName || (ignoreCase && strings.EqualFold(method.Name, methodName)) {
			return method, true
		}
	}
	return reflect.Method{}, false
}

func methodParamsAssignable(methodType reflect.Type, paramTypes []reflect.Type) bool {
	if len(paramTypes) == 0 {
		return true
	}
	if methodType.NumIn()-1 != len(paramTypes) {
		return false
	}
	for i, paramType := range paramTypes {
		if paramType == nil {
			continue
		}
		expected := methodType.In(i + 1)
		if !paramType.AssignableTo(expected) && !paramType.ConvertibleTo(expected) {
			return false
		}
	}
	return true
}

func valuesForCall(fnType reflect.Type, offset int, args []any) []reflect.Value {
	values := make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		expected := fnType.In(i + offset)
		if arg == nil {
			values = append(values, reflect.Zero(expected))
			continue
		}
		v := reflect.ValueOf(arg)
		switch {
		case v.Type().AssignableTo(expected):
			values = append(values, v)
		case v.Type().ConvertibleTo(expected):
			values = append(values, v.Convert(expected))
		default:
			values = append(values, reflect.Zero(expected))
		}
	}
	return values
}

func call(fn reflect.Value, args []reflect.Value) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invoke failed: %v", r)
		}
	}()
	outs := fn.Call(args)
	if len(outs) == 0 {
		return nil, nil
	}
	if len(outs) == 1 {
		value, _ := valueInterface(outs[0], fieldAccessConfig{unsafeAccess: true})
		return value, nil
	}
	values := make([]any, len(outs))
	for i, out := range outs {
		values[i], _ = valueInterface(out, fieldAccessConfig{unsafeAccess: true})
	}
	return values, nil
}
