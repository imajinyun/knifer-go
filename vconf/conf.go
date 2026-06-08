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
	// ExpandOption customizes configuration variable expansion per call.
	ExpandOption = confimpl.ExpandOption
	// ValueOption customizes typed value getters per call.
	ValueOption = confimpl.ValueOption
	// BindOption customizes struct binding per call.
	BindOption = confimpl.BindOption
	// SchemaOption customizes schema validation per call.
	SchemaOption = confimpl.SchemaOption
	// ParseOption customizes ParseByExt and full YAML parsing helpers per call.
	ParseOption = confimpl.ParseOption
	// FieldRule describes one schema validation rule.
	FieldRule = confimpl.FieldRule
	// Schema describes configuration validation rules.
	Schema = confimpl.Schema
	// WatchOptions controls configuration file polling behavior.
	WatchOptions = confimpl.WatchOptions
	// WatchEvent describes a detected configuration file change.
	WatchEvent = confimpl.WatchEvent
	// WatchTicker stops a watch polling ticker created by WatchTickerFactory.
	WatchTicker = confimpl.WatchTicker
	// WatchTickerFactory creates a ticker channel and stopper for WatchWithOptions.
	WatchTickerFactory = confimpl.WatchTickerFactory
)

const (
	TypeString      = confimpl.TypeString
	TypeBool        = confimpl.TypeBool
	TypeInt         = confimpl.TypeInt
	TypeFloat       = confimpl.TypeFloat
	TypeList        = confimpl.TypeList
	DefaultMaxBytes = confimpl.DefaultMaxBytes
)

// New creates an empty Conf.
func New() *Conf { return confimpl.New() }

// Load 读取并解析 setting/properties 配置文件。Load reads and parses a setting/properties file.
func Load(path string) (*Conf, error) { return LoadWithOptions(path, LoadOptions{}) }

// LoadWithOptions reads and parses a configuration file with advanced options.
func LoadWithOptions(path string, opts LoadOptions) (*Conf, error) {
	return confimpl.LoadWithOptions(path, opts)
}

// LoadFiles loads multiple configuration files and merges them in order.
func LoadFiles(paths ...string) (*Conf, error) { return LoadFilesWithOptions(LoadOptions{}, paths...) }

// LoadFilesWithOptions loads multiple configuration files using opts and merges them in order.
func LoadFilesWithOptions(opts LoadOptions, paths ...string) (*Conf, error) {
	return confimpl.LoadFilesWithOptions(opts, paths...)
}

// LoadRemote loads configuration from an HTTP(S) URL.
func LoadRemote(rawURL string) (*Conf, error) { return LoadRemoteWithOptions(rawURL, LoadOptions{}) }

// LoadRemoteWithOptions loads configuration from an HTTP(S) URL with options.
func LoadRemoteWithOptions(rawURL string, opts LoadOptions) (*Conf, error) {
	return confimpl.LoadRemoteWithOptions(rawURL, opts)
}

// LoadRemoteSafe loads configuration from an HTTP(S) URL with SSRF-oriented safety checks enabled.
func LoadRemoteSafe(rawURL string) (*Conf, error) { return confimpl.LoadRemoteSafe(rawURL) }

// LoadRemoteSafeWithOptions loads configuration from an HTTP(S) URL with SSRF-oriented safety checks enabled.
func LoadRemoteSafeWithOptions(rawURL string, opts LoadOptions) (*Conf, error) {
	return confimpl.LoadRemoteSafeWithOptions(rawURL, opts)
}

// Merge merges configurations in order. Later configurations override earlier ones.
func Merge(configs ...*Conf) *Conf { return confimpl.Merge(configs...) }

// WithEnvLookup sets the environment lookup function used for ${ENV:NAME} placeholders.
func WithEnvLookup(lookup func(string) string) ExpandOption { return confimpl.WithEnvLookup(lookup) }

// WithIntParser sets the parser used by Conf.GetIntWithOptions.
func WithIntParser(parser func(string) (int, error)) ValueOption {
	return confimpl.WithIntParser(parser)
}

// WithBoolParser sets the parser used by Conf.GetBoolWithOptions.
func WithBoolParser(parser func(string) (bool, error)) ValueOption {
	return confimpl.WithBoolParser(parser)
}

// WithBindBoolParser sets the bool parser used by Conf.BindWithOptions and Conf.BindGroupWithOptions.
func WithBindBoolParser(parser func(string) (bool, error)) BindOption {
	return confimpl.WithBindBoolParser(parser)
}

// WithBindIntParser sets the signed integer parser used by Conf.BindWithOptions and Conf.BindGroupWithOptions.
func WithBindIntParser(parser func(string, int, int) (int64, error)) BindOption {
	return confimpl.WithBindIntParser(parser)
}

// WithBindUintParser sets the unsigned integer parser used by Conf.BindWithOptions and Conf.BindGroupWithOptions.
func WithBindUintParser(parser func(string, int, int) (uint64, error)) BindOption {
	return confimpl.WithBindUintParser(parser)
}

// WithBindFloatParser sets the floating-point parser used by Conf.BindWithOptions and Conf.BindGroupWithOptions.
func WithBindFloatParser(parser func(string, int) (float64, error)) BindOption {
	return confimpl.WithBindFloatParser(parser)
}

// WithSchemaBoolParser sets the bool parser used by Conf.ValidateSchemaWithOptions and Conf.ValidateStructWithOptions.
func WithSchemaBoolParser(parser func(string) (bool, error)) SchemaOption {
	return confimpl.WithSchemaBoolParser(parser)
}

// WithSchemaIntParser sets the signed integer parser used by Conf.ValidateSchemaWithOptions and Conf.ValidateStructWithOptions.
func WithSchemaIntParser(parser func(string, int, int) (int64, error)) SchemaOption {
	return confimpl.WithSchemaIntParser(parser)
}

// WithSchemaFloatParser sets the floating-point parser used by Conf.ValidateSchemaWithOptions and Conf.ValidateStructWithOptions.
func WithSchemaFloatParser(parser func(string, int) (float64, error)) SchemaOption {
	return confimpl.WithSchemaFloatParser(parser)
}

// Base64Decrypt decodes base64 encrypted configuration values.
func Base64Decrypt(cipherText string) (string, error) { return confimpl.Base64Decrypt(cipherText) }

// SchemaFromStruct builds validation schema rules from conf tags on dst.
func SchemaFromStruct(dst any) (Schema, error) { return confimpl.SchemaFromStruct(dst) }

// LoadProfile loads a configuration file and applies profile-specific overrides.
func LoadProfile(path, profile string) (*Conf, error) {
	return LoadProfileWithOptions(path, profile, LoadOptions{})
}

// LoadProfileWithOptions loads a configuration file with options and applies profile-specific overrides.
func LoadProfileWithOptions(path, profile string, opts LoadOptions) (*Conf, error) {
	return confimpl.LoadProfileWithOptions(path, profile, opts)
}

// Parse 解析 setting/properties 文本内容。Parse parses setting/properties content.
func Parse(content string) (*Conf, error) { return confimpl.Parse(content) }

// ParseBytes 解析 setting/properties 字节内容。ParseBytes parses setting/properties content.
func ParseBytes(content []byte) (*Conf, error) { return confimpl.ParseBytes(content) }

// ParseByExt parses content according to path extension.
func ParseByExt(path string, content []byte) (*Conf, error) {
	return confimpl.ParseByExt(path, content)
}

// ParseByExtWithOptions parses content according to path extension with custom providers.
func ParseByExtWithOptions(path string, content []byte, opts ...ParseOption) (*Conf, error) {
	return confimpl.ParseByExtWithOptions(path, content, opts...)
}

// WithYAMLUnmarshalFunc sets the YAML unmarshal provider used by ParseYAMLFullWithOptions.
func WithYAMLUnmarshalFunc(unmarshal func([]byte, any) error) ParseOption {
	return confimpl.WithYAMLUnmarshalFunc(unmarshal)
}

// WithTOMLUnmarshalFunc sets the TOML unmarshal provider used by ParseTOMLWithOptions.
func WithTOMLUnmarshalFunc(unmarshal func([]byte, any) error) ParseOption {
	return confimpl.WithTOMLUnmarshalFunc(unmarshal)
}

// WithParserForExt sets the parser used by ParseByExtWithOptions for an extension.
func WithParserForExt(ext string, parser func([]byte) (*Conf, error)) ParseOption {
	return confimpl.WithParserForExt(ext, parser)
}

// ParseYAML 将简单 YAML 子集解析为分组配置。ParseYAML parses a small YAML subset into grouped configuration.
func ParseYAML(content string) (*Conf, error) { return confimpl.ParseYAML(content) }

// ParseYAMLFull parses YAML using yaml.v3 and flattens nested objects into grouped keys.
func ParseYAMLFull(content string) (*Conf, error) { return confimpl.ParseYAMLFull(content) }

// ParseYAMLFullWithOptions parses YAML using a configurable unmarshal provider and flattens nested objects into grouped keys.
func ParseYAMLFullWithOptions(content string, opts ...ParseOption) (*Conf, error) {
	return confimpl.ParseYAMLFullWithOptions(content, opts...)
}

// ParseTOML parses common TOML key-value and section syntax into grouped configuration.
func ParseTOML(content string) (*Conf, error) { return confimpl.ParseTOML(content) }

// ParseTOMLWithOptions parses common TOML syntax into grouped configuration with custom providers.
func ParseTOMLWithOptions(content string, opts ...ParseOption) (*Conf, error) {
	return confimpl.ParseTOMLWithOptions(content, opts...)
}

// Watch polls path and calls onChange after successful reloads.
func Watch(path string, interval time.Duration, onChange func(*Conf, error)) (func(), error) {
	return WatchWithOptions(path, WatchOptions{Interval: interval}, onChange)
}

// WatchWithOptions polls path with options and calls onChange after successful reloads.
func WatchWithOptions(path string, opts WatchOptions, onChange func(*Conf, error)) (func(), error) {
	return confimpl.WatchWithOptions(path, opts, onChange)
}
