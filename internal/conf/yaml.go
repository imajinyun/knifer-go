package conf

import (
	"bufio"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func defaultYAMLUnmarshal(data []byte, out any) error {
	return yaml.Unmarshal(data, out)
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
			return nil, invalidInputf("invalid yaml line %d: %s", lineNo, line)
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
		return nil, wrapConfigParse("scan yaml content", err)
	}
	return s, nil
}

// ParseYAMLFull parses YAML using yaml.v3 and flattens nested objects into grouped keys.
func ParseYAMLFull(content string) (*Conf, error) {
	return ParseYAMLFullWithOptions(content)
}

// ParseYAMLFullWithOptions parses YAML using a configurable unmarshal provider and flattens nested objects into grouped keys.
func ParseYAMLFullWithOptions(content string, opts ...ParseOption) (*Conf, error) {
	cfg := applyParseOptions(opts)
	var root any
	if err := cfg.yamlUnmarshal([]byte(content), &root); err != nil {
		return nil, wrapConfigParse("parse yaml content", err)
	}
	c := New()
	flattenYAML(c, defaultGroup, "", root)
	return c, nil
}

func flattenYAML(c *Conf, group, prefix string, value any) {
	switch v := value.(type) {
	case map[string]any:
		for k, child := range v {
			if group == defaultGroup && prefix == "" && k == "profile" {
				flattenYAMLProfiles(c, child)
				continue
			}
			if prefix == "" {
				if _, ok := child.(map[string]any); ok {
					flattenYAML(c, k, "", child)
					continue
				}
			}
			next := k
			if prefix != "" {
				next = prefix + "." + k
			}
			flattenYAML(c, group, next, child)
		}
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, scalarString(item))
		}
		c.SetByGroup(group, prefix, strings.Join(parts, ","))
	default:
		if prefix != "" {
			c.SetByGroup(group, prefix, scalarString(v))
		}
	}
}

func flattenYAMLProfiles(c *Conf, value any) {
	profiles, ok := value.(map[string]any)
	if !ok {
		flattenYAML(c, "profile", "", value)
		return
	}
	for profile, child := range profiles {
		group := "profile." + profile
		flattenYAMLProfileGroup(c, group, "", child)
	}
}

func flattenYAMLProfileGroup(c *Conf, group, prefix string, value any) {
	switch v := value.(type) {
	case map[string]any:
		for k, child := range v {
			next := k
			if prefix != "" {
				next = prefix + "." + k
			}
			if prefix == "" {
				if _, ok := child.(map[string]any); ok {
					flattenYAMLProfileGroup(c, group+"."+k, "", child)
					continue
				}
			}
			flattenYAMLProfileGroup(c, group, next, child)
		}
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, scalarString(item))
		}
		c.SetByGroup(group, prefix, strings.Join(parts, ","))
	default:
		if prefix != "" {
			c.SetByGroup(group, prefix, scalarString(v))
		}
	}
}

func scalarString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	default:
		return fmt.Sprint(x)
	}
}
