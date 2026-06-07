package json

import "strings"

// JSONArray 对应 the utility JSONArray，是有序的 JSON 值列表。
type JSONArray struct {
	cfg    *Config
	values []any
}

// NewJSONArray 创建空数组。
func NewJSONArray() *JSONArray { return NewJSONArrayWithConfig(nil) }

// NewJSONArrayWithConfig 使用配置创建。
func NewJSONArrayWithConfig(cfg *Config) *JSONArray {
	if cfg == nil {
		cfg = NewConfig()
	}
	return &JSONArray{cfg: cfg}
}

// Config 返回配置。
func (a *JSONArray) Config() *Config { return a.cfg }

// Len 元素数量。
func (a *JSONArray) Len() int { return len(a.values) }

// Get 索引访问，越界返回 nil,false。
func (a *JSONArray) Get(i int) (any, bool) {
	if i < 0 || i >= len(a.values) {
		return nil, false
	}
	return a.values[i], true
}

// GetOrDefault 越界返回默认。
func (a *JSONArray) GetOrDefault(i int, def any) any {
	if v, ok := a.Get(i); ok {
		return v
	}
	return def
}

// IsNull 索引处是否为 JSON null。
func (a *JSONArray) IsNull(i int) bool {
	v, ok := a.Get(i)
	if !ok {
		return false
	}
	return IsNull(v)
}

// Add 追加元素。
func (a *JSONArray) Add(value any) *JSONArray {
	v := wrap(value, a.cfg)
	if a.cfg.IgnoreNullValue && IsNull(v) {
		return a
	}
	a.values = append(a.values, v)
	return a
}

// AddAll 批量追加。
func (a *JSONArray) AddAll(values ...any) *JSONArray {
	for _, v := range values {
		a.Add(v)
	}
	return a
}

// Set 写入指定下标，越界自动 nil 填充。
func (a *JSONArray) Set(i int, value any) *JSONArray {
	v := wrap(value, a.cfg)
	for len(a.values) <= i {
		a.values = append(a.values, Null)
	}
	a.values[i] = v
	return a
}

// Insert 在 i 处插入。
func (a *JSONArray) Insert(i int, value any) *JSONArray {
	v := wrap(value, a.cfg)
	if i < 0 {
		i = 0
	}
	if i >= len(a.values) {
		a.values = append(a.values, v)
		return a
	}
	a.values = append(a.values, nil)
	copy(a.values[i+1:], a.values[i:])
	a.values[i] = v
	return a
}

// Remove 删除指定下标元素，越界返回 false。
func (a *JSONArray) Remove(i int) bool {
	if i < 0 || i >= len(a.values) {
		return false
	}
	a.values = append(a.values[:i], a.values[i+1:]...)
	return true
}

// Range 顺序遍历元素。
func (a *JSONArray) Range(fn func(i int, v any) bool) {
	for i, v := range a.values {
		if !fn(i, v) {
			return
		}
	}
}

// ToSlice 转为 []any。
func (a *JSONArray) ToSlice() []any {
	out := make([]any, len(a.values))
	copy(out, a.values)
	return out
}

// Join 将所有元素以分隔符连接（值通过 toString 转换）。
func (a *JSONArray) Join(sep string) string {
	if len(a.values) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range a.values {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(toString(v, "", a.cfg))
	}
	return b.String()
}

// 类型化 getter。

// GetString 索引取字符串。
func (a *JSONArray) GetString(i int) string { return a.GetStringOr(i, "") }

// GetStringOr 索引取字符串或默认。
func (a *JSONArray) GetStringOr(i int, def string) string {
	v, ok := a.Get(i)
	if !ok {
		return def
	}
	return toString(v, def, a.cfg)
}

// GetInt 索引取 int。
func (a *JSONArray) GetInt(i int) int { return int(a.GetInt64Or(i, 0)) }

// GetInt64 索引取 int64。
func (a *JSONArray) GetInt64(i int) int64 { return a.GetInt64Or(i, 0) }

// GetInt64Or 索引取 int64 或默认。
func (a *JSONArray) GetInt64Or(i int, def int64) int64 {
	v, ok := a.Get(i)
	if !ok {
		return def
	}
	return toInt64(v, def, a.cfg)
}

// GetFloat64 索引取 float64。
func (a *JSONArray) GetFloat64(i int) float64 { return a.GetFloat64Or(i, 0) }

// GetFloat64Or 索引取 float64 或默认。
func (a *JSONArray) GetFloat64Or(i int, def float64) float64 {
	v, ok := a.Get(i)
	if !ok {
		return def
	}
	return toFloat64(v, def, a.cfg)
}

// GetBool 索引取 bool。
func (a *JSONArray) GetBool(i int) bool { return a.GetBoolOr(i, false) }

// GetBoolOr 索引取 bool 或默认。
func (a *JSONArray) GetBoolOr(i int, def bool) bool {
	v, ok := a.Get(i)
	if !ok {
		return def
	}
	return toBool(v, def, a.cfg)
}

// GetJSONObject 索引取 JSONObject。
func (a *JSONArray) GetJSONObject(i int) *JSONObject {
	v, ok := a.Get(i)
	if !ok {
		return nil
	}
	if obj, ok := v.(*JSONObject); ok {
		return obj
	}
	return nil
}

// GetJSONArray 索引取 JSONArray。
func (a *JSONArray) GetJSONArray(i int) *JSONArray {
	v, ok := a.Get(i)
	if !ok {
		return nil
	}
	if arr, ok := v.(*JSONArray); ok {
		return arr
	}
	return nil
}

// String 紧凑输出。
func (a *JSONArray) String() string {
	s, _ := writeValue(a, 0)
	return s
}

// ToString 紧凑输出。
func (a *JSONArray) ToString() string { return a.String() }

// ToStringPretty 4 空格缩进输出。
func (a *JSONArray) ToStringPretty() string {
	s, _ := writeValue(a, defaultIndent(a.cfg))
	return s
}

// MarshalJSON 实现 encoding/json.Marshaler。
func (a *JSONArray) MarshalJSON() ([]byte, error) {
	s, err := writeValue(a, 0)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// UnmarshalJSON 实现 encoding/json.Unmarshaler。
func (a *JSONArray) UnmarshalJSON(b []byte) error {
	v, err := parseBytes(b)
	if err != nil {
		return err
	}
	arr, ok := v.(*JSONArray)
	if !ok {
		return NewJSONError("expect json array, got %T", v)
	}
	a.cfg = arr.cfg
	a.values = arr.values
	return nil
}

// GetByPath 路径读取。
func (a *JSONArray) GetByPath(path string) any { return getByPath(a, path) }

// PutByPath 路径写入。
func (a *JSONArray) PutByPath(path string, value any) error { return putByPath(a, path, value) }
