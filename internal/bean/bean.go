package bean

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// Option customizes bean mapping behavior.
type Option func(*Options)

// Options controls struct/map property mapping.
type Options struct {
	// TagNames are checked in order to resolve field names and aliases.
	TagNames []string
	// WeaklyTyped allows string/numeric/bool conversions and recursive map/struct assignment.
	WeaklyTyped bool
	// CaseInsensitive matches field names and aliases case-insensitively.
	CaseInsensitive bool
	// IgnoreEmpty skips empty source values when copying to maps or structs.
	IgnoreEmpty bool
	// IgnoreZero skips zero source values when copying to maps or structs.
	IgnoreZero bool
	// StrictUnused reports unmatched source keys or fields as errors after assignment.
	StrictUnused bool
	// ParseBool parses strings during weak bool conversion. nil keeps default semantics.
	ParseBool func(string) (bool, error)
	// ParseInt parses strings during weak signed integer conversion. nil means strconv.ParseInt.
	ParseInt func(string, int, int) (int64, error)
	// ParseUint parses strings during weak unsigned integer conversion. nil means strconv.ParseUint.
	ParseUint func(string, int, int) (uint64, error)
	// ParseFloat parses strings during weak floating-point conversion. nil means strconv.ParseFloat.
	ParseFloat func(string, int) (float64, error)
}

// Result reports which source properties were consumed, skipped, or left unused.
type Result struct {
	Matched []string
	Skipped []string
	Unused  []string
}

func (r *Result) sort() {
	slices.Sort(r.Matched)
	slices.Sort(r.Skipped)
	slices.Sort(r.Unused)
}

// NewOptions returns default mapping options.
func NewOptions() Options {
	return Options{
		TagNames:        []string{"bean", "json", "xml", "ref"},
		WeaklyTyped:     true,
		CaseInsensitive: true,
		ParseBool:       defaultBoolParser,
		ParseInt:        strconv.ParseInt,
		ParseUint:       strconv.ParseUint,
		ParseFloat:      strconv.ParseFloat,
	}
}

// WithTagNames sets tag names used to resolve field names and aliases.
func WithTagNames(names ...string) Option {
	return func(o *Options) {
		o.TagNames = slices.Clone(names)
	}
}

// WithWeaklyTyped controls whether weak type conversion is enabled.
func WithWeaklyTyped(enable bool) Option { return func(o *Options) { o.WeaklyTyped = enable } }

// WithCaseInsensitive controls case-insensitive name matching.
func WithCaseInsensitive(enable bool) Option { return func(o *Options) { o.CaseInsensitive = enable } }

// WithIgnoreEmpty skips empty source values.
func WithIgnoreEmpty(enable bool) Option { return func(o *Options) { o.IgnoreEmpty = enable } }

// WithIgnoreZero skips zero source values.
func WithIgnoreZero(enable bool) Option { return func(o *Options) { o.IgnoreZero = enable } }

// WithStrictUnused reports unmatched source keys or fields as errors after assignment.
func WithStrictUnused(enable bool) Option { return func(o *Options) { o.StrictUnused = enable } }

// WithBoolParser sets the parser used during weak string-to-bool conversion.
func WithBoolParser(parser func(string) (bool, error)) Option {
	return func(o *Options) {
		if parser != nil {
			o.ParseBool = parser
		}
	}
}

// WithIntParser sets the parser used during weak string-to-signed-integer conversion.
func WithIntParser(parser func(string, int, int) (int64, error)) Option {
	return func(o *Options) {
		if parser != nil {
			o.ParseInt = parser
		}
	}
}

// WithUintParser sets the parser used during weak string-to-unsigned-integer conversion.
func WithUintParser(parser func(string, int, int) (uint64, error)) Option {
	return func(o *Options) {
		if parser != nil {
			o.ParseUint = parser
		}
	}
}

// WithFloatParser sets the parser used during weak string-to-floating-point conversion.
func WithFloatParser(parser func(string, int) (float64, error)) Option {
	return func(o *Options) {
		if parser != nil {
			o.ParseFloat = parser
		}
	}
}

// ToMap converts a struct or map to map[string]any using field tags and aliases.
func ToMap(src any, opts ...Option) (map[string]any, error) {
	out := map[string]any{}
	if err := FillMap(src, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// FillMap copies properties from src into dst.
func FillMap(src any, dst map[string]any, opts ...Option) error {
	cfg := applyOptions(opts...)
	_, err := fillMapResult(src, dst, cfg)
	return err
}

func fillMapResult(src any, dst map[string]any, cfg Options) (Result, error) {
	if dst == nil {
		return Result{}, invalidBeanInputf("bean: dst map is nil")
	}
	props, err := collectProperties(src, cfg)
	if err != nil {
		return Result{}, err
	}
	result := Result{}
	for _, prop := range props {
		if shouldSkip(prop.value, cfg) {
			result.Skipped = append(result.Skipped, prop.name)
			continue
		}
		dst[prop.name] = prop.value.Interface()
		result.Matched = append(result.Matched, prop.name)
	}
	result.sort()
	return result, nil
}

// ToStruct copies properties from src into dst, which must be a pointer to struct.
func ToStruct(src any, dst any, opts ...Option) error { return CopyProperties(src, dst, opts...) }

// Decode converts matching properties from src into dst using the configured weak conversion rules.
func Decode(src any, dst any, opts ...Option) error {
	_, err := DecodeResult(src, dst, opts...)
	return err
}

// DecodeResult converts matching properties from src into dst and reports mapping metadata.
func DecodeResult(src any, dst any, opts ...Option) (Result, error) {
	return copyPropertiesResult(src, dst, opts...)
}

// Merge copies one or more sources into dst from left to right.
// Later sources override earlier sources.
func Merge(dst any, sources ...any) error {
	_, err := MergeResult(dst, sources...)
	return err
}

// MergeResult copies one or more sources into dst and reports aggregate mapping metadata.
func MergeResult(dst any, sources ...any) (Result, error) {
	return mergeResult(dst, nil, sources...)
}

// MergeWithOptions copies sources into dst from left to right using options.
func MergeWithOptions(dst any, sources []any, opts ...Option) error {
	_, err := MergeResultWithOptions(dst, sources, opts...)
	return err
}

// MergeResultWithOptions copies sources into dst from left to right using options and reports metadata.
func MergeResultWithOptions(dst any, sources []any, opts ...Option) (Result, error) {
	return mergeResult(dst, opts, sources...)
}

func mergeResult(dst any, opts []Option, sources ...any) (Result, error) {
	result := Result{}
	for _, src := range sources {
		partial, err := copyPropertiesResult(src, dst, opts...)
		if err != nil {
			return result, err
		}
		result.Matched = append(result.Matched, partial.Matched...)
		result.Skipped = append(result.Skipped, partial.Skipped...)
		result.Unused = append(result.Unused, partial.Unused...)
	}
	result.sort()
	return result, nil
}

// Copy is an alias of CopyProperties.
func Copy(src any, dst any, opts ...Option) error { return CopyProperties(src, dst, opts...) }

// CopyProperties copies matching properties between struct/map values.
func CopyProperties(src any, dst any, opts ...Option) error {
	_, err := copyPropertiesResult(src, dst, opts...)
	return err
}

func copyPropertiesResult(src any, dst any, opts ...Option) (Result, error) {
	cfg := applyOptions(opts...)
	if dst == nil {
		return Result{}, invalidBeanInputf("bean: dst is nil")
	}
	if m, ok := dst.(map[string]any); ok {
		return fillMapResult(src, m, cfg)
	}
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Pointer || dv.IsNil() {
		return Result{}, invalidBeanInputf("bean: dst must be a non-nil pointer or map[string]any")
	}
	dv = indirect(dv)
	if !dv.IsValid() || dv.Kind() != reflect.Struct {
		return Result{}, invalidBeanInputf("bean: dst must point to struct")
	}
	props, err := collectProperties(src, cfg)
	if err != nil {
		return Result{}, err
	}
	index := propertyIndex(props, cfg)
	used := map[string]struct{}{}
	result := Result{}
	for _, field := range structFields(dv.Type(), cfg) {
		fv := fieldByIndex(dv, field.index)
		if !fv.IsValid() || !fv.CanSet() {
			continue
		}
		prop, ok := lookupProperty(index, field.aliases, cfg)
		if !ok {
			continue
		}
		name := prop.name
		used[normalize(name, cfg)] = struct{}{}
		if shouldSkip(prop.value, cfg) {
			result.Skipped = append(result.Skipped, name)
			continue
		}
		if err := assignValue(fv, prop.value, cfg); err != nil {
			return Result{}, wrapBeanInput("bean: set field "+field.goName, err)
		}
		result.Matched = append(result.Matched, name)
	}
	result.Unused = unusedProperties(props, used, cfg)
	result.sort()
	if cfg.StrictUnused && len(result.Unused) > 0 {
		return result, invalidBeanInputf("bean: unused source properties: %s", strings.Join(result.Unused, ", "))
	}
	return result, nil
}

func unusedProperties(props []property, used map[string]struct{}, cfg Options) []string {
	out := make([]string, 0)
	seen := map[string]struct{}{}
	for _, prop := range props {
		key := normalize(prop.name, cfg)
		if _, ok := used[key]; ok {
			continue
		}
		if _, ok := seen[prop.name]; ok {
			continue
		}
		seen[prop.name] = struct{}{}
		out = append(out, prop.name)
	}
	slices.Sort(out)
	return out
}

type property struct {
	name    string
	aliases []string
	value   reflect.Value
}

type fieldInfo struct {
	goName  string
	name    string
	aliases []string
	index   []int
}

func applyOptions(opts ...Option) Options {
	cfg := NewOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.ParseBool == nil {
		cfg.ParseBool = defaultBoolParser
	}
	if cfg.ParseInt == nil {
		cfg.ParseInt = strconv.ParseInt
	}
	if cfg.ParseUint == nil {
		cfg.ParseUint = strconv.ParseUint
	}
	if cfg.ParseFloat == nil {
		cfg.ParseFloat = strconv.ParseFloat
	}
	return cfg
}

func optionsFromConfig(cfg Options) []Option {
	return []Option{
		WithTagNames(cfg.TagNames...),
		WithWeaklyTyped(cfg.WeaklyTyped),
		WithCaseInsensitive(cfg.CaseInsensitive),
		WithIgnoreEmpty(cfg.IgnoreEmpty),
		WithIgnoreZero(cfg.IgnoreZero),
		WithStrictUnused(cfg.StrictUnused),
		WithBoolParser(cfg.ParseBool),
		WithIntParser(cfg.ParseInt),
		WithUintParser(cfg.ParseUint),
		WithFloatParser(cfg.ParseFloat),
	}
}

func collectProperties(src any, cfg Options) ([]property, error) {
	if src == nil {
		return nil, invalidBeanInputf("bean: src is nil")
	}
	sv := indirect(reflect.ValueOf(src))
	if !sv.IsValid() {
		return nil, invalidBeanInputf("bean: src is nil")
	}
	switch sv.Kind() {
	case reflect.Map:
		return mapProperties(sv), nil
	case reflect.Struct:
		return structProperties(sv, cfg), nil
	default:
		return nil, invalidBeanInputf("bean: unsupported src kind %s", sv.Kind())
	}
}

func mapProperties(v reflect.Value) []property {
	props := make([]property, 0, v.Len())
	for _, key := range v.MapKeys() {
		name := fmt.Sprint(key.Interface())
		props = append(props, property{name: name, aliases: []string{name}, value: v.MapIndex(key)})
	}
	return props
}

func structProperties(v reflect.Value, cfg Options) []property {
	fields := structFields(v.Type(), cfg)
	props := make([]property, 0, len(fields))
	for _, field := range fields {
		fv := fieldByIndex(v, field.index)
		if !fv.IsValid() || !fv.CanInterface() {
			continue
		}
		props = append(props, property{name: field.name, aliases: field.aliases, value: fv})
	}
	return props
}

func structFields(t reflect.Type, cfg Options) []fieldInfo {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	out := make([]fieldInfo, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		name, aliases, omit, tagged := resolveFieldNames(field, cfg)
		if omit {
			continue
		}
		if field.Anonymous && !tagged {
			for _, nested := range structFields(field.Type, cfg) {
				nested.index = append(slices.Clone(field.Index), nested.index...)
				out = append(out, nested)
			}
			continue
		}
		out = append(out, fieldInfo{goName: field.Name, name: name, aliases: aliases, index: field.Index})
	}
	return out
}

func resolveFieldNames(field reflect.StructField, cfg Options) (string, []string, bool, bool) {
	name := field.Name
	aliases := []string{field.Name}
	tagged := false
	for _, tagName := range cfg.TagNames {
		if tagName == "" {
			continue
		}
		tag := field.Tag.Get(tagName)
		if tag == "" {
			continue
		}
		tagged = true
		parts := strings.Split(tag, ",")
		primary := strings.TrimSpace(parts[0])
		if primary == "-" {
			return "", nil, true, true
		}
		if primary != "" {
			name = primary
			aliases = append(aliases, primary)
		}
		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
			if values, ok := strings.CutPrefix(part, "alias="); ok {
				aliases = append(aliases, splitAliases(values)...)
			}
			if values, ok := strings.CutPrefix(part, "aliases="); ok {
				aliases = append(aliases, splitAliases(values)...)
			}
		}
	}
	return name, uniqueStrings(append([]string{name}, aliases...)), false, tagged
}

func splitAliases(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '|' || r == ';' })
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func propertyIndex(props []property, cfg Options) map[string]property {
	index := make(map[string]property, len(props))
	for _, prop := range props {
		for _, alias := range uniqueStrings(append([]string{prop.name}, prop.aliases...)) {
			index[normalize(alias, cfg)] = prop
		}
	}
	return index
}

func lookupProperty(index map[string]property, aliases []string, cfg Options) (property, bool) {
	for _, alias := range aliases {
		if prop, ok := index[normalize(alias, cfg)]; ok {
			return prop, true
		}
	}
	return property{}, false
}

func normalize(name string, cfg Options) string {
	name = strings.TrimSpace(name)
	if cfg.CaseInsensitive {
		return strings.ToLower(name)
	}
	return name
}

func shouldSkip(v reflect.Value, cfg Options) bool {
	if !v.IsValid() {
		return true
	}
	if cfg.IgnoreEmpty && isEmptyValue(v) {
		return true
	}
	if cfg.IgnoreZero && v.IsZero() {
		return true
	}
	return false
}

func isEmptyValue(v reflect.Value) bool {
	v = indirect(v)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	default:
		return false
	}
}

func assignValue(dst, src reflect.Value, cfg Options) error {
	if !src.IsValid() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	if src.Kind() == reflect.Interface && !src.IsNil() {
		src = src.Elem()
	}
	if dst.Kind() == reflect.Pointer {
		if isNilValue(src) {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assignValue(dst.Elem(), src, cfg)
	}
	if isNilValue(src) {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return nil
	}
	if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return nil
	}
	if !cfg.WeaklyTyped {
		return fmt.Errorf("cannot assign %s to %s", src.Type(), dst.Type())
	}
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(valueString(src))
		return nil
	case reflect.Bool:
		b, err := valueBool(src, cfg)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := valueInt(src, dst.Type().Bits(), cfg)
		if err != nil {
			return err
		}
		dst.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := valueUint(src, dst.Type().Bits(), cfg)
		if err != nil {
			return err
		}
		dst.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		f, err := valueFloat(src, dst.Type().Bits(), cfg)
		if err != nil {
			return err
		}
		dst.SetFloat(f)
		return nil
	case reflect.Slice:
		return assignSlice(dst, src, cfg)
	case reflect.Map:
		return assignMap(dst, src, cfg)
	case reflect.Struct:
		return assignStruct(dst, src, cfg)
	case reflect.Interface:
		dst.Set(src)
		return nil
	default:
		return fmt.Errorf("cannot assign %s to %s", src.Type(), dst.Type())
	}
}

func assignSlice(dst, src reflect.Value, cfg Options) error {
	if dst.Type().Elem().Kind() == reflect.Uint8 && src.Kind() == reflect.String {
		dst.SetBytes([]byte(src.String()))
		return nil
	}
	src = indirect(src)
	if !src.IsValid() || (src.Kind() != reflect.Slice && src.Kind() != reflect.Array) {
		return fmt.Errorf("cannot assign %s to %s", src.Type(), dst.Type())
	}
	out := reflect.MakeSlice(dst.Type(), src.Len(), src.Len())
	for i := 0; i < src.Len(); i++ {
		if err := assignValue(out.Index(i), src.Index(i), cfg); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	dst.Set(out)
	return nil
}

func assignMap(dst, src reflect.Value, cfg Options) error {
	if src.Kind() == reflect.Struct {
		m, err := ToMap(src.Interface(), optionsFromConfig(cfg)...)
		if err != nil {
			return err
		}
		src = reflect.ValueOf(m)
	}
	src = indirect(src)
	if !src.IsValid() || src.Kind() != reflect.Map {
		return fmt.Errorf("cannot assign %s to %s", src.Type(), dst.Type())
	}
	out := reflect.MakeMapWithSize(dst.Type(), src.Len())
	for _, key := range src.MapKeys() {
		newKey := reflect.New(dst.Type().Key()).Elem()
		if err := assignValue(newKey, key, cfg); err != nil {
			return fmt.Errorf("map key: %w", err)
		}
		newVal := reflect.New(dst.Type().Elem()).Elem()
		if err := assignValue(newVal, src.MapIndex(key), cfg); err != nil {
			return fmt.Errorf("map value: %w", err)
		}
		out.SetMapIndex(newKey, newVal)
	}
	dst.Set(out)
	return nil
}

func assignStruct(dst, src reflect.Value, cfg Options) error {
	if !src.CanInterface() {
		return fmt.Errorf("cannot assign inaccessible %s", src.Type())
	}
	return CopyProperties(src.Interface(), dst.Addr().Interface(), optionsFromConfig(cfg)...)
}

func valueString(v reflect.Value) string {
	v = indirect(v)
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	if v.CanInterface() {
		return fmt.Sprint(v.Interface())
	}
	return fmt.Sprint(v)
}

func valueBool(v reflect.Value, cfg Options) (bool, error) {
	v = indirect(v)
	if !v.IsValid() {
		return false, nil
	}
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.String:
		return cfg.ParseBool(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %s to bool", v.Type())
	}
}

func valueInt(v reflect.Value, bits int, cfg Options) (int64, error) {
	v = indirect(v)
	if !v.IsValid() {
		return 0, nil
	}
	var n int64
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u := v.Uint()
		if u > ^uint64(0)>>1 {
			return 0, fmt.Errorf("integer overflow")
		}
		n = int64(u)
	case reflect.Float32, reflect.Float64:
		n = int64(v.Float())
	case reflect.Bool:
		if v.Bool() {
			n = 1
		}
	case reflect.String:
		s := strings.TrimSpace(v.String())
		if s == "" {
			return 0, nil
		}
		parsed, err := cfg.ParseInt(s, 10, bits)
		if err == nil {
			return parsed, nil
		}
		f, ferr := cfg.ParseFloat(s, 64)
		if ferr != nil {
			return 0, err
		}
		n = int64(f)
	default:
		return 0, fmt.Errorf("cannot convert %s to int", v.Type())
	}
	if bits > 0 && bits < 64 {
		min := -(int64(1) << (bits - 1))
		max := int64(1)<<(bits-1) - 1
		if n < min || n > max {
			return 0, fmt.Errorf("integer overflow")
		}
	}
	return n, nil
}

func valueUint(v reflect.Value, bits int, cfg Options) (uint64, error) {
	v = indirect(v)
	if !v.IsValid() {
		return 0, nil
	}
	var n uint64
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n = v.Uint()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := v.Int()
		if i < 0 {
			return 0, fmt.Errorf("negative value %d", i)
		}
		n = uint64(i)
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if f < 0 {
			return 0, fmt.Errorf("negative value %v", f)
		}
		n = uint64(f)
	case reflect.Bool:
		if v.Bool() {
			n = 1
		}
	case reflect.String:
		s := strings.TrimSpace(v.String())
		if s == "" {
			return 0, nil
		}
		parsed, err := cfg.ParseUint(s, 10, bits)
		if err == nil {
			return parsed, nil
		}
		f, ferr := cfg.ParseFloat(s, 64)
		if ferr != nil || f < 0 {
			return 0, err
		}
		n = uint64(f)
	default:
		return 0, fmt.Errorf("cannot convert %s to uint", v.Type())
	}
	if bits > 0 && bits < 64 && n > uint64(1<<bits-1) {
		return 0, fmt.Errorf("unsigned integer overflow")
	}
	return n, nil
}

func valueFloat(v reflect.Value, bits int, cfg Options) (float64, error) {
	v = indirect(v)
	if !v.IsValid() {
		return 0, nil
	}
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), nil
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.String:
		s := strings.TrimSpace(v.String())
		if s == "" {
			return 0, nil
		}
		return cfg.ParseFloat(s, bits)
	default:
		return 0, fmt.Errorf("cannot convert %s to float", v.Type())
	}
}

func defaultBoolParser(s string) (bool, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "yes", "y", "ok", "1", "on":
		return true, nil
	case "false", "no", "n", "0", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse bool %q", s)
	}
}

func indirect(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func isNilValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct || i < 0 || i >= v.NumField() {
			return reflect.Value{}
		}
		v = v.Field(i)
	}
	return v
}
