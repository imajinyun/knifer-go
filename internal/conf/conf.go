package conf

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

const defaultGroup = ""

// Conf stores grouped key-value configuration.
type Conf struct {
	data map[string]map[string]string
}

// New creates an empty Conf.
func New() *Conf {
	return &Conf{data: map[string]map[string]string{defaultGroup: {}}}
}

// Load 读取并解析 setting/properties 配置文件。Load reads and parses a setting/properties file.
func Load(path string) (*Conf, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(b)
}

// Parse 解析 setting/properties 文本内容。Parse parses setting/properties content.
func Parse(content string) (*Conf, error) { return ParseBytes([]byte(content)) }

// ParseBytes 解析 setting/properties 字节内容。ParseBytes parses setting/properties content.
func ParseBytes(content []byte) (*Conf, error) {
	s := New()
	group := defaultGroup
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			group = strings.TrimSpace(line[1 : len(line)-1])
			s.ensureGroup(group)
			continue
		}
		idx := strings.IndexAny(line, "=:")
		if idx < 0 {
			return nil, fmt.Errorf("invalid setting line %d: %s", lineNo, line)
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key == "" {
			return nil, fmt.Errorf("empty setting key at line %d", lineNo)
		}
		s.SetByGroup(group, key, unquote(value))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return s, nil
}

// ParseYAML 将简单 YAML 子集解析为分组配置。ParseYAML parses a small YAML subset into grouped configuration.
func ParseYAML(content string) (*Conf, error) {
	s := New()
	group := defaultGroup
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " \t"))
		idx := strings.Index(line, ":")
		if idx < 0 {
			return nil, fmt.Errorf("invalid yaml line %d: %s", lineNo, line)
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if value == "" && indent == 0 {
			group = key
			s.ensureGroup(group)
			continue
		}
		if indent == 0 {
			group = defaultGroup
		}
		s.SetByGroup(group, key, unquote(value))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return s, nil
}

// Get 从默认分组获取配置值。Get returns a value from the default group.
func (s *Conf) Get(key string) string { return s.GetByGroup(defaultGroup, key) }

// GetOrDefault 从默认分组获取配置值，不存在时返回 def。GetOrDefault returns a value from the default group or def when absent.
func (s *Conf) GetOrDefault(key, def string) string {
	if v, ok := s.Lookup(defaultGroup, key); ok {
		return v
	}
	return def
}

// GetByGroup 获取指定分组中的配置值。GetByGroup returns a grouped value.
func (s *Conf) GetByGroup(group, key string) string {
	v, _ := s.Lookup(group, key)
	return v
}

// Lookup 获取指定分组中的配置值并返回是否存在。Lookup returns a grouped value and whether it exists.
func (s *Conf) Lookup(group, key string) (string, bool) {
	if s == nil || s.data == nil {
		return "", false
	}
	m, ok := s.data[group]
	if !ok {
		return "", false
	}
	v, ok := m[key]
	return v, ok
}

// GetInt 从默认分组获取 int 值，不存在或格式非法时返回 def。GetInt returns an int value from the default group or def when absent/invalid.
func (s *Conf) GetInt(key string, def int) int {
	v, ok := s.Lookup(defaultGroup, key)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// GetBool 从默认分组获取 bool 值，不存在或格式非法时返回 def。GetBool returns a bool value from the default group or def when absent/invalid.
func (s *Conf) GetBool(key string, def bool) bool {
	v, ok := s.Lookup(defaultGroup, key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// Set 将配置值写入默认分组。Set stores a value in the default group.
func (s *Conf) Set(key, value string) { s.SetByGroup(defaultGroup, key, value) }

// SetByGroup 将配置值写入指定分组。SetByGroup stores a grouped value.
func (s *Conf) SetByGroup(group, key, value string) {
	s.ensureGroup(group)
	s.data[group][key] = value
}

// Groups 返回全部分组名称。Groups returns all group names.
func (s *Conf) Groups() []string {
	if s == nil || s.data == nil {
		return []string{}
	}
	groups := make([]string, 0, len(s.data))
	for g := range s.data {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return groups
}

// Keys 返回指定分组中的全部键。Keys returns keys from group.
func (s *Conf) Keys(group string) []string {
	if s == nil || s.data == nil {
		return []string{}
	}
	m := s.data[group]
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ToMap 返回所有分组配置的深拷贝。ToMap returns a deep copy of all groups.
func (s *Conf) ToMap() map[string]map[string]string {
	if s == nil || s.data == nil {
		return map[string]map[string]string{}
	}
	out := make(map[string]map[string]string, len(s.data))
	for g, m := range s.data {
		out[g] = make(map[string]string, len(m))
		for k, v := range m {
			out[g][k] = v
		}
	}
	return out
}

func (s *Conf) ensureGroup(group string) {
	if s.data == nil {
		s.data = map[string]map[string]string{}
	}
	if _, ok := s.data[group]; !ok {
		s.data[group] = map[string]string{}
	}
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
