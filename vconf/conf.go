package vconf

import confimpl "github.com/imajinyun/go-knifer/internal/conf"

// Conf stores grouped key-value configuration.
type Conf = confimpl.Conf

// New creates an empty Conf.
func New() *Conf { return confimpl.New() }

// Load 读取并解析 setting/properties 配置文件。Load reads and parses a setting/properties file.
func Load(path string) (*Conf, error) { return confimpl.Load(path) }

// Parse 解析 setting/properties 文本内容。Parse parses setting/properties content.
func Parse(content string) (*Conf, error) { return confimpl.Parse(content) }

// ParseBytes 解析 setting/properties 字节内容。ParseBytes parses setting/properties content.
func ParseBytes(content []byte) (*Conf, error) { return confimpl.ParseBytes(content) }

// ParseYAML 将简单 YAML 子集解析为分组配置。ParseYAML parses a small YAML subset into grouped configuration.
func ParseYAML(content string) (*Conf, error) { return confimpl.ParseYAML(content) }
