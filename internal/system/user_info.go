package system

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type userInfoConfig struct {
	currentUser func() (*user.User, error)
	getenv      func(string) string
	getwd       func() (string, error)
	tempDir     func() string
}

// UserInfoOption customizes user information collection per call.
type UserInfoOption func(*userInfoConfig)

// WithCurrentUserFunc sets the function used to discover the current OS user.
func WithCurrentUserFunc(fn func() (*user.User, error)) UserInfoOption {
	return func(c *userInfoConfig) {
		if fn != nil {
			c.currentUser = fn
		}
	}
}

// WithUserEnvLookup sets the environment lookup function used by NewUserInfoWithOptions.
func WithUserEnvLookup(lookup func(string) string) UserInfoOption {
	return func(c *userInfoConfig) {
		if lookup != nil {
			c.getenv = lookup
		}
	}
}

// WithWorkingDirFunc sets the function used to discover the current working directory.
func WithWorkingDirFunc(fn func() (string, error)) UserInfoOption {
	return func(c *userInfoConfig) {
		if fn != nil {
			c.getwd = fn
		}
	}
}

// WithTempDirFunc sets the function used to discover the temporary directory.
func WithTempDirFunc(fn func() string) UserInfoOption {
	return func(c *userInfoConfig) {
		if fn != nil {
			c.tempDir = fn
		}
	}
}

func applyUserInfoOptions(opts []UserInfoOption) userInfoConfig {
	cfg := userInfoConfig{
		currentUser: user.Current,
		getenv:      os.Getenv,
		getwd:       os.Getwd,
		tempDir:     os.TempDir,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.currentUser == nil {
		cfg.currentUser = user.Current
	}
	if cfg.getenv == nil {
		cfg.getenv = os.Getenv
	}
	if cfg.getwd == nil {
		cfg.getwd = os.Getwd
	}
	if cfg.tempDir == nil {
		cfg.tempDir = os.TempDir
	}
	return cfg
}

// UserInfo describes current logged-in user information.
type UserInfo struct {
	Name       string
	HomeDir    string
	CurrentDir string
	TempDir    string
	Language   string
	Country    string
}

// NewUserInfo creates current user information.
func NewUserInfo() *UserInfo {
	return NewUserInfoWithOptions()
}

// NewUserInfoWithOptions creates current user information with per-call options.
func NewUserInfoWithOptions(opts ...UserInfoOption) *UserInfo {
	cfg := applyUserInfoOptions(opts)
	u := &UserInfo{}

	if cur, err := cfg.currentUser(); err == nil && cur != nil {
		u.Name = cur.Username
		u.HomeDir = fixPath(cur.HomeDir)
	} else {
		u.Name = cfg.getenv("USER")
		if u.Name == "" {
			u.Name = cfg.getenv("USERNAME")
		}
		u.HomeDir = fixPath(cfg.getenv("HOME"))
	}

	if dir, err := cfg.getwd(); err == nil {
		u.CurrentDir = fixPath(dir)
	}
	u.TempDir = fixPath(cfg.tempDir())

	lang, country := parseLocale(cfg.getenv("LANG"))
	if lang == "" {
		lang, country = parseLocale(cfg.getenv("LC_ALL"))
	}
	u.Language = lang
	u.Country = country
	return u
}

// GetName returns the user name.
func (u *UserInfo) GetName() string { return u.Name }

// GetHomeDir returns the home directory.
func (u *UserInfo) GetHomeDir() string { return u.HomeDir }

// GetCurrentDir returns the current working directory.
func (u *UserInfo) GetCurrentDir() string { return u.CurrentDir }

// GetTempDir returns the temporary directory.
func (u *UserInfo) GetTempDir() string { return u.TempDir }

// GetLanguage returns the language, such as zh.
func (u *UserInfo) GetLanguage() string { return u.Language }

// GetCountry returns the country or region, such as CN.
func (u *UserInfo) GetCountry() string { return u.Country }

// String implements fmt.Stringer.
func (u *UserInfo) String() string {
	var b strings.Builder
	appendLine(&b, "User Name:        ", u.Name)
	appendLine(&b, "User Home Dir:    ", u.HomeDir)
	appendLine(&b, "User Current Dir: ", u.CurrentDir)
	appendLine(&b, "User Temp Dir:    ", u.TempDir)
	appendLine(&b, "User Language:    ", u.Language)
	appendLine(&b, "User Country:     ", u.Country)
	return b.String()
}

// fixPath appends a trailing path separator.
func fixPath(p string) string {
	if p == "" {
		return p
	}
	return addSuffixIfNot(p, string(filepath.Separator))
}

// parseLocale parses a LANG string such as "zh_CN.UTF-8" and returns language and country.
func parseLocale(locale string) (lang, country string) {
	if locale == "" {
		return "", ""
	}
	if i := strings.IndexByte(locale, '.'); i >= 0 {
		locale = locale[:i]
	}
	parts := strings.Split(locale, "_")
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], parts[1]
	}
}
