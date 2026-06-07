package json

import (
	"strconv"
	"strings"
)

// JSONObject 对应 the utility JSONObject，按插入顺序保留键。
type JSONObject struct {
	cfg    *Config
	keys   []string
	values map[string]any
	// keyMap 在 ignoreCase 时保存小写 → 原始键映射。
	keyMap map[string]string
}

// NewJSONObject 创建空对象。
func NewJSONObject() *JSONObject {
	return NewJSONObjectWithConfig(nil)
}

// NewJSONObjectWithConfig 使用指定配置创建对象。
func NewJSONObjectWithConfig(cfg *Config) *JSONObject {
	if cfg == nil {
		cfg = NewConfig()
	}
	o := &JSONObject{cfg: cfg, values: map[string]any{}}
	if cfg.IgnoreCase {
		o.keyMap = map[string]string{}
	}
	return o
}

// Config 返回配置。
func (o *JSONObject) Config() *Config { return o.cfg }

// Len 键数量。
func (o *JSONObject) Len() int { return len(o.keys) }

// Keys 返回有序键列表的拷贝。
func (o *JSONObject) Keys() []string {
	out := make([]string, len(o.keys))
	copy(out, o.keys)
	return out
}

// canonicalKey 在 ignoreCase 时返回真实键，否则返回原键。
func (o *JSONObject) canonicalKey(key string) (string, bool) {
	if o.cfg.IgnoreCase {
		if real, ok := o.keyMap[strings.ToLower(key)]; ok {
			return real, true
		}
		return key, false
	}
	_, ok := o.values[key]
	return key, ok
}

// Has 是否存在 key。
func (o *JSONObject) Has(key string) bool {
	_, ok := o.canonicalKey(key)
	return ok
}

// Get 获取原始值；不存在时返回 nil,false。
func (o *JSONObject) Get(key string) (any, bool) {
	real, ok := o.canonicalKey(key)
	if !ok {
		return nil, false
	}
	v, ok := o.values[real]
	return v, ok
}

// GetOrDefault 获取值或默认。
func (o *JSONObject) GetOrDefault(key string, def any) any {
	if v, ok := o.Get(key); ok {
		return v
	}
	return def
}

// IsNull 判断 key 是否存在并且值为 JSON null。
func (o *JSONObject) IsNull(key string) bool {
	v, ok := o.Get(key)
	if !ok {
		return false
	}
	return IsNull(v)
}

// Set 写入键值，返回自身以支持链式调用。
func (o *JSONObject) Set(key string, value any) *JSONObject {
	value = wrap(value, o.cfg)
	if o.cfg.IgnoreNullValue && IsNull(value) {
		return o
	}
	if real, ok := o.canonicalKey(key); ok {
		o.values[real] = value
		return o
	}
	o.keys = append(o.keys, key)
	o.values[key] = value
	if o.cfg.IgnoreCase {
		o.keyMap[strings.ToLower(key)] = key
	}
	return o
}

// Put 与 Set 相同（兼容 the utility toolkit 命名）。
func (o *JSONObject) Put(key string, value any) *JSONObject { return o.Set(key, value) }

// Remove 删除键，返回是否删除成功。
func (o *JSONObject) Remove(key string) bool {
	real, ok := o.canonicalKey(key)
	if !ok {
		return false
	}
	delete(o.values, real)
	for i, k := range o.keys {
		if k == real {
			o.keys = append(o.keys[:i], o.keys[i+1:]...)
			break
		}
	}
	if o.cfg.IgnoreCase {
		delete(o.keyMap, strings.ToLower(real))
	}
	return true
}

// ForEach 按插入顺序遍历。
func (o *JSONObject) ForEach(fn func(key string, value any) bool) {
	for _, k := range o.keys {
		if !fn(k, o.values[k]) {
			return
		}
	}
}

// ToMap 转为普通 map（值为原始 JSON 值）。
func (o *JSONObject) ToMap() map[string]any {
	out := make(map[string]any, len(o.keys))
	for _, k := range o.keys {
		out[k] = o.values[k]
	}
	return out
}

// 一系列类型化 getter。

// GetString 取字符串，不存在或类型不符时返回 ""。
func (o *JSONObject) GetString(key string) string { return o.GetStringOr(key, "") }

// GetStringOr 取字符串或默认。
func (o *JSONObject) GetStringOr(key, def string) string {
	v, ok := o.Get(key)
	if !ok {
		return def
	}
	return toString(v, def, o.cfg)
}

// GetInt 取 int。
func (o *JSONObject) GetInt(key string) int { return int(o.GetInt64Or(key, 0)) }

// GetIntOr 取 int 或默认。
func (o *JSONObject) GetIntOr(key string, def int) int {
	return int(o.GetInt64Or(key, int64(def)))
}

// GetInt64 取 int64。
func (o *JSONObject) GetInt64(key string) int64 { return o.GetInt64Or(key, 0) }

// GetInt64Or 取 int64 或默认。
func (o *JSONObject) GetInt64Or(key string, def int64) int64 {
	v, ok := o.Get(key)
	if !ok {
		return def
	}
	return toInt64(v, def, o.cfg)
}

// GetFloat64 取 float64。
func (o *JSONObject) GetFloat64(key string) float64 { return o.GetFloat64Or(key, 0) }

// GetFloat64Or 取 float64 或默认。
func (o *JSONObject) GetFloat64Or(key string, def float64) float64 {
	v, ok := o.Get(key)
	if !ok {
		return def
	}
	return toFloat64(v, def, o.cfg)
}

// GetBool 取 bool。
func (o *JSONObject) GetBool(key string) bool { return o.GetBoolOr(key, false) }

// GetBoolOr 取 bool 或默认。
func (o *JSONObject) GetBoolOr(key string, def bool) bool {
	v, ok := o.Get(key)
	if !ok {
		return def
	}
	return toBool(v, def, o.cfg)
}

// GetJSONObject 取嵌套对象，不存在或非对象返回 nil。
func (o *JSONObject) GetJSONObject(key string) *JSONObject {
	v, ok := o.Get(key)
	if !ok {
		return nil
	}
	if obj, ok := v.(*JSONObject); ok {
		return obj
	}
	return nil
}

// GetJSONArray 取嵌套数组，不存在或非数组返回 nil。
func (o *JSONObject) GetJSONArray(key string) *JSONArray {
	v, ok := o.Get(key)
	if !ok {
		return nil
	}
	if arr, ok := v.(*JSONArray); ok {
		return arr
	}
	return nil
}

// String 输出紧凑 JSON 字符串。
func (o *JSONObject) String() string {
	s, _ := writeValue(o, 0)
	return s
}

// ToString 紧凑输出。
func (o *JSONObject) ToString() string { return o.String() }

// ToStringPretty 4 空格缩进输出。
func (o *JSONObject) ToStringPretty() string {
	s, _ := writeValue(o, defaultIndent(o.cfg))
	return s
}

// MarshalJSON 实现 encoding/json.Marshaler。
func (o *JSONObject) MarshalJSON() ([]byte, error) {
	s, err := writeValue(o, 0)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// UnmarshalJSON 实现 encoding/json.Unmarshaler。
func (o *JSONObject) UnmarshalJSON(b []byte) error {
	v, err := parseBytes(b)
	if err != nil {
		return err
	}
	parsed, ok := v.(*JSONObject)
	if !ok {
		return NewJSONError("expect json object, got %T", v)
	}
	o.cfg = parsed.cfg
	o.keys = parsed.keys
	o.values = parsed.values
	o.keyMap = parsed.keyMap
	return nil
}

// GetByPath 通过路径表达式读取值。
func (o *JSONObject) GetByPath(path string) any { return getByPath(o, path) }

// PutByPath 通过路径表达式写入值。
func (o *JSONObject) PutByPath(path string, value any) error { return putByPath(o, path, value) }

// defaultIndent 返回配置中的缩进，为 0 返回 4。
func defaultIndent(cfg *Config) int {
	if cfg != nil && cfg.IndentFactor > 0 {
		return cfg.IndentFactor
	}
	return 4
}

// indexKey 给数字 key 转 int，便于与数组操作互通。
func parseIndex(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}
