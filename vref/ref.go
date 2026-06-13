package vref

import (
	"reflect"

	refimpl "github.com/imajinyun/go-knifer/internal/ref"
)

type (
	FieldFilter       = refimpl.FieldFilter
	MethodFilter      = refimpl.MethodFilter
	FieldAccessOption = refimpl.FieldAccessOption
)

// WithUnsafeAccess controls whether unexported addressable fields may be accessed via unsafe.
func WithUnsafeAccess(enabled bool) FieldAccessOption { return refimpl.WithUnsafeAccess(enabled) }

// WithAllowUnexported controls whether unexported addressable fields may be accessed via unsafe.
func WithAllowUnexported(enabled bool) FieldAccessOption { return refimpl.WithAllowUnexported(enabled) }

func TypeOf(object any) reflect.Type                  { return refimpl.TypeOf(object) }
func IndirectType(typ reflect.Type) reflect.Type      { return refimpl.IndirectType(typ) }
func ValueOf(object any) reflect.Value                { return refimpl.ValueOf(object) }
func IndirectValue(value reflect.Value) reflect.Value { return refimpl.IndirectValue(value) }
func IsNil(object any) bool                           { return refimpl.IsNil(object) }
func IsNilValue(value reflect.Value) bool             { return refimpl.IsNilValue(value) }
func IsFuncType(typ reflect.Type) bool                { return refimpl.IsFuncType(typ) }
func IsRangeableType(typ reflect.Type) bool           { return refimpl.IsRangeableType(typ) }
func IsCollectionType(typ reflect.Type) bool          { return refimpl.IsCollectionType(typ) }
func IsSliceType(typ reflect.Type) bool               { return refimpl.IsSliceType(typ) }
func IsArrayType(typ reflect.Type) bool               { return refimpl.IsArrayType(typ) }
func IsMapType(typ reflect.Type) bool                 { return refimpl.IsMapType(typ) }
func ImplementsError(typ reflect.Type) bool           { return refimpl.ImplementsError(typ) }
func ImplementsContext(typ reflect.Type) bool         { return refimpl.ImplementsContext(typ) }
func GetConstructor(target any) reflect.Value         { return refimpl.GetConstructor(target) }
func GetConstructors(target any) []reflect.Value      { return refimpl.GetConstructors(target) }
func GetConstructorsDirectly(target any) []reflect.Value {
	return refimpl.GetConstructorsDirectly(target)
}
func HasField(target any, name string) bool                 { return refimpl.HasField(target, name) }
func GetFieldName(field reflect.StructField) string         { return refimpl.GetFieldName(field) }
func GetField(target any, name string) reflect.StructField  { return refimpl.GetField(target, name) }
func GetFieldMap(target any) map[string]reflect.StructField { return refimpl.GetFieldMap(target) }
func GetFields(target any, filters ...FieldFilter) []reflect.StructField {
	return refimpl.GetFields(target, filters...)
}

func GetPublicFieldNames(target any) []string { return refimpl.GetPublicFieldNames(target) }

func GetFieldsDirectly(target any, withEmbeddedFields bool) []reflect.StructField {
	return refimpl.GetFieldsDirectly(target, withEmbeddedFields)
}
func GetFieldValue(obj any, fieldName string) any { return GetFieldValueWithOptions(obj, fieldName) }

func GetFieldValueWithOptions(obj any, fieldName string, opts ...FieldAccessOption) any {
	return refimpl.GetFieldValueWithOptions(obj, fieldName, opts...)
}

func GetStaticFieldValue(value any) any { return refimpl.GetStaticFieldValue(value) }
func GetFieldsValue(obj any, filters ...FieldFilter) []any {
	return GetFieldsValueWithOptions(obj, nil, filters...)
}

func GetFieldsValueWithOptions(obj any, opts []FieldAccessOption, filters ...FieldFilter) []any {
	return refimpl.GetFieldsValueWithOptions(obj, opts, filters...)
}

func SetFieldValue(obj any, fieldName string, value any) error {
	return SetFieldValueWithOptions(obj, fieldName, value)
}

func SetFieldValueWithOptions(obj any, fieldName string, value any, opts ...FieldAccessOption) error {
	return refimpl.SetFieldValueWithOptions(obj, fieldName, value, opts...)
}
func IsOuterClassField(field reflect.StructField) bool { return refimpl.IsOuterClassField(field) }
func GetPublicMethodNames(target any) []string         { return refimpl.GetPublicMethodNames(target) }
func GetPublicMethods(target any, filters ...MethodFilter) []reflect.Method {
	return refimpl.GetPublicMethods(target, filters...)
}

func GetPublicMethod(target any, methodName string, paramTypes ...reflect.Type) (reflect.Method, bool) {
	return refimpl.GetPublicMethod(target, methodName, paramTypes...)
}

func GetMethodOfObj(obj any, methodName string, args ...any) (reflect.Method, bool) {
	return refimpl.GetMethodOfObj(obj, methodName, args...)
}

func GetMethodIgnoreCase(target any, methodName string, paramTypes ...reflect.Type) (reflect.Method, bool) {
	return refimpl.GetMethodIgnoreCase(target, methodName, paramTypes...)
}

func GetMethod(target any, ignoreCase bool, methodName string, paramTypes ...reflect.Type) (reflect.Method, bool) {
	return refimpl.GetMethod(target, ignoreCase, methodName, paramTypes...)
}

func GetMethodByName(target any, methodName string) (reflect.Method, bool) {
	return refimpl.GetMethodByName(target, methodName)
}

func GetMethodByNameIgnoreCase(target any, methodName string) (reflect.Method, bool) {
	return refimpl.GetMethodByNameIgnoreCase(target, methodName)
}
func GetMethodNames(target any) []string { return refimpl.GetMethodNames(target) }
func GetMethods(target any, filters ...MethodFilter) []reflect.Method {
	return refimpl.GetMethods(target, filters...)
}

func GetMethodsDirectly(target any, withSupers, withMethodFromObject bool) []reflect.Method {
	return refimpl.GetMethodsDirectly(target, withSupers, withMethodFromObject)
}
func IsEqualsMethod(method reflect.Method) bool   { return refimpl.IsEqualsMethod(method) }
func IsHashCodeMethod(method reflect.Method) bool { return refimpl.IsHashCodeMethod(method) }
func IsToStringMethod(method reflect.Method) bool { return refimpl.IsToStringMethod(method) }
func IsEmptyParam(method reflect.Method) bool     { return refimpl.IsEmptyParam(method) }
func IsGetterOrSetterIgnoreCase(method reflect.Method) bool {
	return refimpl.IsGetterOrSetterIgnoreCase(method)
}

func IsGetterOrSetter(method reflect.Method, ignoreCase bool) bool {
	return refimpl.IsGetterOrSetter(method, ignoreCase)
}

func NewInstance(target any, params ...any) (any, error) {
	return refimpl.NewInstance(target, params...)
}
func NewInstanceIfPossible(target any) any          { return refimpl.NewInstanceIfPossible(target) }
func InvokeStatic(fn any, args ...any) (any, error) { return refimpl.InvokeStatic(fn, args...) }
func InvokeWithCheck(obj any, method reflect.Method, args ...any) (any, error) {
	return refimpl.InvokeWithCheck(obj, method, args...)
}

func InvokeMethod(obj any, method reflect.Method, args ...any) (any, error) {
	return refimpl.InvokeMethod(obj, method, args...)
}
func InvokeRaw(fn any, args ...any) (any, error) { return refimpl.InvokeRaw(fn, args...) }
func Invoke(obj any, methodName string, args ...any) (any, error) {
	return refimpl.Invoke(obj, methodName, args...)
}
func InvokeFunc(fn any, args ...any) (any, error) { return refimpl.InvokeFunc(fn, args...) }
func SetAccessible[T any](object T) T             { return refimpl.SetAccessible(object) }
func RemoveFinalModify(object any)                { refimpl.RemoveFinalModify(object) }
