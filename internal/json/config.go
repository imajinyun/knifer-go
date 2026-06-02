package json

// Config controls JSON serialization behavior.
type Config struct {
	// IgnoreNullValue 序列化时忽略 null。
	IgnoreNullValue bool
	// IgnoreCase 键不区分大小写（仅在 JSONObject 上生效，写入时按首次出现的大小写存储）。
	IgnoreCase bool
	// IgnoreError 在转换失败时忽略错误。
	IgnoreError bool
	// DateFormat 日期格式（time.Time 的 layout），为空时输出毫秒数。
	DateFormat string
	// IndentFactor pretty 输出时缩进字符数。
	IndentFactor int
}

// NewConfig 创建一个默认配置。
func NewConfig() *Config {
	return &Config{IndentFactor: 4}
}

// CreateConfig creates a default JSON config.
func CreateConfig() *Config { return NewConfig() }

// Clone 拷贝配置。
func (c *Config) Clone() *Config {
	if c == nil {
		return NewConfig()
	}
	cp := *c
	return &cp
}
