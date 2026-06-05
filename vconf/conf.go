package vconf

import (
	"time"

	confimpl "github.com/imajinyun/go-knifer/internal/conf"
)

// Conf stores grouped key-value configuration.
type Conf = confimpl.Conf

// Error is the configuration module error type.
type Error = confimpl.ConfError

type (
	// LoadOptions controls local and remote configuration loading behavior.
	LoadOptions = confimpl.LoadOptions
	// DecryptFunc decrypts encrypted configuration values.
	DecryptFunc = confimpl.DecryptFunc
	// FieldRule describes one schema validation rule.
	FieldRule = confimpl.FieldRule
	// Schema describes configuration validation rules.
	Schema = confimpl.Schema
	// WatchOptions controls configuration file polling behavior.
	WatchOptions = confimpl.WatchOptions
	// WatchEvent describes a detected configuration file change.
	WatchEvent = confimpl.WatchEvent
)

const (
	TypeString = confimpl.TypeString
	TypeBool   = confimpl.TypeBool
	TypeInt    = confimpl.TypeInt
	TypeFloat  = confimpl.TypeFloat
	TypeList   = confimpl.TypeList
)

// New creates an empty Conf.
func New() *Conf { return confimpl.New() }

// Load 读取并解析 setting/properties 配置文件。Load reads and parses a setting/properties file.
func Load(path string) (*Conf, error) { return confimpl.Load(path) }

// LoadWithOptions reads and parses a configuration file with advanced options.
func LoadWithOptions(path string, opts LoadOptions) (*Conf, error) {
	return confimpl.LoadWithOptions(path, opts)
}

// LoadFiles loads multiple configuration files and merges them in order.
func LoadFiles(paths ...string) (*Conf, error) { return confimpl.LoadFiles(paths...) }

// LoadFilesWithOptions loads multiple configuration files using opts and merges them in order.
func LoadFilesWithOptions(opts LoadOptions, paths ...string) (*Conf, error) {
	return confimpl.LoadFilesWithOptions(opts, paths...)
}

// LoadRemote loads configuration from an HTTP(S) URL.
func LoadRemote(rawURL string) (*Conf, error) { return confimpl.LoadRemote(rawURL) }

// LoadRemoteWithOptions loads configuration from an HTTP(S) URL with options.
func LoadRemoteWithOptions(rawURL string, opts LoadOptions) (*Conf, error) {
	return confimpl.LoadRemoteWithOptions(rawURL, opts)
}

// Merge merges configurations in order. Later configurations override earlier ones.
func Merge(configs ...*Conf) *Conf { return confimpl.Merge(configs...) }

// Base64Decrypt decodes base64 encrypted configuration values.
func Base64Decrypt(cipherText string) (string, error) { return confimpl.Base64Decrypt(cipherText) }

// SchemaFromStruct builds validation schema rules from conf tags on dst.
func SchemaFromStruct(dst any) (Schema, error) { return confimpl.SchemaFromStruct(dst) }

// LoadProfile loads a configuration file and applies profile-specific overrides.
func LoadProfile(path, profile string) (*Conf, error) { return confimpl.LoadProfile(path, profile) }

// Parse 解析 setting/properties 文本内容。Parse parses setting/properties content.
func Parse(content string) (*Conf, error) { return confimpl.Parse(content) }

// ParseBytes 解析 setting/properties 字节内容。ParseBytes parses setting/properties content.
func ParseBytes(content []byte) (*Conf, error) { return confimpl.ParseBytes(content) }

// ParseByExt parses content according to path extension.
func ParseByExt(path string, content []byte) (*Conf, error) {
	return confimpl.ParseByExt(path, content)
}

// ParseYAML 将简单 YAML 子集解析为分组配置。ParseYAML parses a small YAML subset into grouped configuration.
func ParseYAML(content string) (*Conf, error) { return confimpl.ParseYAML(content) }

// ParseYAMLFull parses YAML using yaml.v3 and flattens nested objects into grouped keys.
func ParseYAMLFull(content string) (*Conf, error) { return confimpl.ParseYAMLFull(content) }

// ParseTOML parses common TOML key-value and section syntax into grouped configuration.
func ParseTOML(content string) (*Conf, error) { return confimpl.ParseTOML(content) }

// Watch polls path and calls onChange after successful reloads.
func Watch(path string, interval time.Duration, onChange func(*Conf, error)) (func(), error) {
	return confimpl.Watch(path, interval, onChange)
}

// WatchWithOptions polls path with options and calls onChange after successful reloads.
func WatchWithOptions(path string, opts WatchOptions, onChange func(*Conf, error)) (func(), error) {
	return confimpl.WatchWithOptions(path, opts, onChange)
}
