package conf

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultMaxBytes is the default local/remote configuration read limit.
const DefaultMaxBytes int64 = 16 << 20

// DecryptFunc decrypts encrypted configuration values.
type DecryptFunc func(cipherText string) (string, error)

// LoadOptions controls file/remote loading behavior.
type LoadOptions struct {
	// Context controls cancellation for remote loading. Timeout is applied on top when set.
	Context context.Context
	// AllowInclude enables include/import keys in loaded configs.
	AllowInclude bool
	// IncludeKeys are keys used to discover included files. Defaults to include/import.
	IncludeKeys []string
	// Decrypt decrypts ENC(...) values after loading and merging.
	Decrypt DecryptFunc
	// RemoteClient is used by LoadRemote. Defaults to http.DefaultClient.
	RemoteClient *http.Client
	// Headers are added to remote config requests.
	Headers http.Header
	// RequestFactory optionally builds remote config requests. When set, Headers are applied after factory creation.
	RequestFactory func(ctx context.Context, rawURL string) (*http.Request, error)
	// RemoteAllowedHosts restricts remote config HTTP(S) hosts when non-empty.
	RemoteAllowedHosts []string
	// RejectPrivateRemoteHosts rejects localhost, loopback, private, and link-local HTTP(S) hosts unless allowed explicitly.
	RejectPrivateRemoteHosts bool
	// CheckRemoteRedirect validates redirect targets with the same remote URL policy.
	CheckRemoteRedirect bool
	// Timeout bounds remote loading when RemoteClient has no timeout.
	Timeout time.Duration
	// MaxBytes limits local and remote config bytes. Zero uses DefaultMaxBytes; negative disables the limit explicitly.
	MaxBytes int64
	// ReadFile optionally reads a local config file. Defaults to os.Open plus MaxBytes limiting.
	ReadFile func(path string, maxBytes int64) ([]byte, error)
	// ParseOptions customize parsing after local or remote content is read.
	ParseOptions []ParseOption
}

// LoadWithOptions reads and parses a configuration file with advanced options.
func LoadWithOptions(path string, opts LoadOptions) (*Conf, error) {
	opts = normalizeLoadOptions(opts)
	return loadFile(path, opts, map[string]bool{})
}

// LoadFiles loads multiple configuration files and merges them in order.
func LoadFiles(paths ...string) (*Conf, error) { return LoadFilesWithOptions(LoadOptions{}, paths...) }

// LoadFilesWithOptions loads multiple configuration files and merges them in order.
func LoadFilesWithOptions(opts LoadOptions, paths ...string) (*Conf, error) {
	configs := make([]*Conf, 0, len(paths))
	for _, path := range paths {
		c, err := LoadWithOptions(path, opts)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return Merge(configs...), nil
}

// LoadRemote loads configuration from an HTTP(S) URL.
func LoadRemote(rawURL string) (*Conf, error) { return LoadRemoteWithOptions(rawURL, LoadOptions{}) }

// LoadRemoteWithOptions loads configuration from an HTTP(S) URL with options.
func LoadRemoteWithOptions(rawURL string, opts LoadOptions) (*Conf, error) {
	opts = normalizeLoadOptions(opts)
	return loadRemote(rawURL, opts)
}

// LoadRemoteSafe loads configuration from an HTTP(S) URL with SSRF-oriented safety checks enabled.
func LoadRemoteSafe(rawURL string) (*Conf, error) {
	return LoadRemoteSafeWithOptions(rawURL, LoadOptions{})
}

// LoadRemoteSafeWithOptions loads configuration from an HTTP(S) URL with SSRF-oriented safety checks enabled.
func LoadRemoteSafeWithOptions(rawURL string, opts LoadOptions) (*Conf, error) {
	opts.RejectPrivateRemoteHosts = true
	opts.CheckRemoteRedirect = true
	opts = normalizeLoadOptions(opts)
	return loadRemote(rawURL, opts)
}

func normalizeLoadOptions(opts LoadOptions) LoadOptions {
	if opts.MaxBytes == 0 {
		opts.MaxBytes = DefaultMaxBytes
	}
	return opts
}

// Merge merges configurations in order. Later configurations override earlier ones.
func Merge(configs ...*Conf) *Conf {
	out := New()
	for _, c := range configs {
		out.Merge(c)
	}
	return out
}

// Merge merges other into s. Existing keys are overwritten by other.
func (s *Conf) Merge(other *Conf) *Conf {
	if s == nil {
		return Merge(other)
	}
	if other == nil || other.data == nil {
		return s
	}
	for group, m := range other.data {
		for key, value := range m {
			s.SetByGroup(group, key, value)
		}
	}
	return s
}

func loadFile(path string, opts LoadOptions, seen map[string]bool) (*Conf, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, wrapConfigIO("resolve config file "+path, err)
	}
	if seen[abs] {
		return nil, invalidInputf("circular config include: %s", path)
	}
	seen[abs] = true
	defer delete(seen, abs)

	b, err := readFileWithOptions(path, opts) // #nosec G304 G703 -- configuration loader intentionally reads caller-provided paths.
	if err != nil {
		return nil, wrapConfigIO("read config file "+path, err)
	}
	current, err := ParseByExtWithOptions(path, b, opts.ParseOptions...)
	if err != nil {
		return nil, err
	}
	if !opts.AllowInclude {
		return current.DecryptValues(opts.Decrypt)
	}

	includes := current.includePaths(includeKeys(opts))
	current.removeIncludeKeys(includeKeys(opts))
	if len(includes) == 0 {
		return current.DecryptValues(opts.Decrypt)
	}
	baseDir := filepath.Dir(path)
	merged := New()
	for _, include := range includes {
		include = strings.TrimSpace(include)
		if include == "" {
			continue
		}
		if !filepath.IsAbs(include) {
			include = filepath.Join(baseDir, include)
		}
		c, err := loadFile(include, opts, seen)
		if err != nil {
			return nil, err
		}
		merged.Merge(c)
	}
	merged.Merge(current)
	return merged.DecryptValues(opts.Decrypt)
}

func loadRemote(rawURL string, opts LoadOptions) (*Conf, error) {
	if err := validateRemoteConfigURL(rawURL, opts); err != nil {
		return nil, err
	}
	client := opts.RemoteClient
	if client == nil {
		client = http.DefaultClient
	}
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var req *http.Request
	var err error
	if opts.RequestFactory != nil {
		req, err = opts.RequestFactory(ctx, rawURL)
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	}
	if err != nil {
		return nil, invalidInputf("invalid remote config url %s: %s", rawURL, err.Error())
	}
	if req == nil {
		return nil, invalidInputf("invalid remote config url %s: request factory returned nil", rawURL)
	}
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}
	if req.URL == nil {
		return nil, invalidInputf("invalid remote config url %s: request url is nil", rawURL)
	}
	if err := validateRemoteConfigURL(req.URL.String(), opts); err != nil {
		return nil, err
	}
	for key, values := range opts.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if opts.CheckRemoteRedirect {
		clone := *client
		clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if req.URL == nil {
				return invalidInputf("invalid remote config redirect: request url is nil")
			}
			return validateRemoteConfigURL(req.URL.String(), opts)
		}
		client = &clone
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, wrapConfigIO("fetch remote config "+rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, invalidInputf("fetch remote config %s: unexpected status %d", rawURL, resp.StatusCode)
	}
	b, err := readAllLimit(resp.Body, opts.MaxBytes)
	if err != nil {
		return nil, wrapConfigIO("read remote config "+rawURL, err)
	}
	parsePath := rawURL
	if u, err := url.Parse(rawURL); err == nil && u.Path != "" {
		parsePath = u.Path
	}
	c, err := ParseByExtWithOptions(parsePath, b, opts.ParseOptions...)
	if err != nil {
		return nil, err
	}
	return c.DecryptValues(opts.Decrypt)
}

func validateRemoteConfigURL(rawURL string, opts LoadOptions) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return invalidInputf("invalid remote config url %s: %s", rawURL, err.Error())
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return invalidInputf("remote config url scheme %q is not allowed", scheme)
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return invalidInputf("remote config url host is blank")
	}
	if len(opts.RemoteAllowedHosts) > 0 && !containsFold(opts.RemoteAllowedHosts, host) {
		return invalidInputf("remote config host %q is not allowed", host)
	}
	if opts.RejectPrivateRemoteHosts && !containsFold(opts.RemoteAllowedHosts, host) {
		private, err := isPrivateHost(host)
		if err != nil {
			return invalidInputf("resolve remote config host %q: %s", host, err.Error())
		}
		if private {
			return invalidInputf("remote config host %q resolves to a private address", host)
		}
	}
	return nil
}

func containsFold(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}
	return false
}

func isPrivateHost(host string) (bool, error) {
	if strings.EqualFold(host, "localhost") {
		return true, nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return isPrivateIP(ip), nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return false, err
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return true, nil
		}
	}
	return false, nil
}

func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func readFileLimit(path string, maxBytes int64) ([]byte, error) {
	f, err := os.Open(path) // #nosec G304 -- configuration loader intentionally reads caller-provided paths.
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return readAllLimit(f, maxBytes)
}

func readFileWithOptions(path string, opts LoadOptions) ([]byte, error) {
	if opts.ReadFile != nil {
		b, err := opts.ReadFile(path, opts.MaxBytes)
		if err != nil {
			return nil, err
		}
		return enforceMaxBytes(b, opts.MaxBytes)
	}
	return readFileLimit(path, opts.MaxBytes)
}

func enforceMaxBytes(b []byte, maxBytes int64) ([]byte, error) {
	if maxBytes > 0 && int64(len(b)) > maxBytes {
		return nil, invalidInputf("config exceeds max bytes: %d", maxBytes)
	}
	return b, nil
}

func readAllLimit(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(r)
	}
	limited := &io.LimitedReader{R: r, N: maxBytes + 1}
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxBytes {
		return nil, invalidInputf("config exceeds max bytes: %d", maxBytes)
	}
	return b, nil
}

func includeKeys(opts LoadOptions) []string {
	if len(opts.IncludeKeys) > 0 {
		return opts.IncludeKeys
	}
	return []string{"include", "import"}
}

func (s *Conf) includePaths(keys []string) []string {
	var out []string
	for _, key := range keys {
		if value, ok := s.Lookup(defaultGroup, key); ok {
			out = append(out, splitList(value)...)
		}
	}
	return out
}

func (s *Conf) removeIncludeKeys(keys []string) {
	for _, key := range keys {
		s.Delete(key)
	}
}
