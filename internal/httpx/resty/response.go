package resty

import (
	"bytes"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/imajinyun/go-knifer/internal/httpx/internal/shared"
	grestry "resty.dev/v3"
)

// HTTPResponse wraps resty.Response and provides convenient readers.
type HTTPResponse struct {
	resp *grestry.Response
	err  error
}

// Cookie contains a response cookie name and value.
type Cookie struct {
	Name  string
	Value string
}

func wrapResponse(r *grestry.Response) *HTTPResponse { return &HTTPResponse{resp: r} }

func readAllWithLimit(r io.Reader, maxBytes int64, readAll func(io.Reader) ([]byte, error)) ([]byte, error) {
	if readAll == nil {
		readAll = io.ReadAll
	}
	if maxBytes <= 0 {
		return readAll(r)
	}
	limited := &io.LimitedReader{R: r, N: maxBytes + 1}
	data, err := readAll(limited)
	if int64(len(data)) > maxBytes {
		return nil, HTTPErrorf("response body exceeds max bytes: %d", maxBytes)
	}
	if err != nil {
		return data, err
	}
	return data, nil
}

type saveConfig struct {
	filePerm        fs.FileMode
	dirPerm         fs.FileMode
	overwrite       bool
	createParents   bool
	defaultFilename string
	stat            func(string) (os.FileInfo, error)
	mkdirAll        func(string, fs.FileMode) error
	openFile        func(string, int, fs.FileMode) (io.WriteCloser, error)
}

// SaveOption customizes response file saving.
type SaveOption func(*saveConfig)

func defaultSaveConfig() saveConfig {
	return saveConfig{filePerm: 0o644, dirPerm: 0o750, overwrite: true, createParents: true, defaultFilename: "download.bin", stat: os.Stat, mkdirAll: os.MkdirAll, openFile: defaultOpenWriteFile}
}

func defaultOpenWriteFile(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
	// #nosec G304 -- SaveAs intentionally writes to the caller-provided destination path.
	return os.OpenFile(path, flag, perm)
}

// WithSaveFilePerm sets the file permission used when creating the destination file.
func WithSaveFilePerm(perm fs.FileMode) SaveOption { return func(c *saveConfig) { c.filePerm = perm } }

// WithSaveDirPerm sets the directory permission used when creating parent directories.
func WithSaveDirPerm(perm fs.FileMode) SaveOption { return func(c *saveConfig) { c.dirPerm = perm } }

// WithSaveOverwrite controls whether an existing destination file may be replaced.
func WithSaveOverwrite(overwrite bool) SaveOption {
	return func(c *saveConfig) { c.overwrite = overwrite }
}

// WithSaveCreateParents controls whether parent directories are created automatically.
func WithSaveCreateParents(create bool) SaveOption {
	return func(c *saveConfig) { c.createParents = create }
}

// WithSaveDefaultFilename sets the fallback file name used when dest is a directory.
func WithSaveDefaultFilename(name string) SaveOption {
	return func(c *saveConfig) {
		if name != "" {
			c.defaultFilename = name
		}
	}
}

// WithSaveStat sets the stat provider used to resolve directory destinations.
func WithSaveStat(stat func(string) (os.FileInfo, error)) SaveOption {
	return func(c *saveConfig) { c.stat = stat }
}

// WithSaveMkdirAll sets the directory creator used when saving responses.
func WithSaveMkdirAll(mkdirAll func(string, fs.FileMode) error) SaveOption {
	return func(c *saveConfig) { c.mkdirAll = mkdirAll }
}

// WithSaveOpenFile sets the file opener used when saving responses.
func WithSaveOpenFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) SaveOption {
	return func(c *saveConfig) { c.openFile = openFile }
}

func applySaveOptions(opts []SaveOption) saveConfig {
	cfg := defaultSaveConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.stat == nil {
		cfg.stat = os.Stat
	}
	if cfg.mkdirAll == nil {
		cfg.mkdirAll = os.MkdirAll
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenWriteFile
	}
	return cfg
}

// Err returns the error raised during execution.
func (r *HTTPResponse) Err() error { return r.err }

// Status returns the HTTP status code, or 0 on error.
func (r *HTTPResponse) Status() int {
	if r.resp == nil {
		return 0
	}
	return r.resp.StatusCode()
}

// IsOK reports whether the status is a 2xx success.
func (r *HTTPResponse) IsOK() bool { return r.Status() >= 200 && r.Status() < 300 }

// Header returns a response header value.
func (r *HTTPResponse) Header(name string) string {
	if r.resp == nil {
		return ""
	}
	return r.resp.Header().Get(name)
}

// Headers returns all response headers.
func (r *HTTPResponse) Headers() HeaderValues {
	if r.resp == nil {
		return nil
	}
	out := HeaderValues{}
	for k, values := range r.resp.Header() {
		out[k] = append([]string(nil), values...)
	}
	return out
}

// Cookies returns cookies from the response.
func (r *HTTPResponse) Cookies() []Cookie {
	if r.resp == nil {
		return nil
	}
	values := r.resp.Header().Values(string(HeaderSetCookie))
	cookies := make([]Cookie, 0, len(values))
	for _, value := range values {
		name, rest, ok := strings.Cut(value, "=")
		if !ok || name == "" {
			continue
		}
		val, _, _ := strings.Cut(rest, ";")
		cookies = append(cookies, Cookie{Name: strings.TrimSpace(name), Value: strings.TrimSpace(val)})
	}
	return cookies
}

// ContentType returns the response Content-Type.
func (r *HTTPResponse) ContentType() string { return r.Header(string(HeaderContentType)) }

// ContentLength returns the response Content-Length.
func (r *HTTPResponse) ContentLength() int64 {
	if r.resp == nil {
		return -1
	}
	return r.resp.Size()
}

// Bytes reads and returns the response body bytes.
func (r *HTTPResponse) Bytes() []byte {
	if r.err != nil {
		return nil
	}
	if r.resp == nil {
		return nil
	}
	return r.resp.Bytes()
}

// Body reads the response body and returns it as a string.
func (r *HTTPResponse) Body() string { return string(r.Bytes()) }

// WriteTo writes the response body to the writer and returns the number of bytes written.
func (r *HTTPResponse) WriteTo(w io.Writer) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	n, err := w.Write(r.Bytes())
	return int64(n), err
}

// SaveAs saves the response body to a file and returns the number of bytes written.
//
// When dest is a directory, the file name is extracted from URL or Content-Disposition automatically.
func (r *HTTPResponse) SaveAs(dest string, opts ...SaveOption) (n int64, err error) {
	if r.resp == nil {
		return 0, HTTPErrorf("no response")
	}
	cfg := applySaveOptions(opts)
	target := dest
	if info, err := cfg.stat(dest); err == nil && info.IsDir() {
		fileName, err := shared.SafeDownloadedFilename(r.fileName())
		if err != nil {
			return 0, err
		}
		if fileName == "" {
			fileName = cfg.defaultFilename
		}
		target, err = shared.SafeJoinDownloadPath(dest, fileName)
		if err != nil {
			return 0, err
		}
	}
	if cfg.createParents {
		if err := cfg.mkdirAll(filepath.Dir(target), cfg.dirPerm); err != nil {
			return 0, NewHTTPError("create parent directory failed", err)
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	f, err := cfg.openFile(target, flag, cfg.filePerm) // #nosec G304 -- SaveAs intentionally writes to a caller-provided destination.
	if err != nil {
		return 0, NewHTTPError("create file failed", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = NewHTTPError("close file failed", closeErr)
		}
	}()
	return r.WriteTo(f)
}

// Close closes the underlying response body when available.
func (r *HTTPResponse) Close() error {
	if r.resp != nil && r.resp.Body != nil {
		return r.resp.Body.Close()
	}
	return nil
}

// RestyRaw returns the original resty response.
func (r *HTTPResponse) RestyRaw() *grestry.Response { return r.resp }

// Result returns the automatically decoded response result, if one was registered on the request.
func (r *HTTPResponse) Result() any {
	if r.resp == nil {
		return nil
	}
	return r.resp.Result()
}

// ErrorResult returns the automatically decoded error response result, if one was registered on the request.
func (r *HTTPResponse) ErrorResult() any {
	if r.resp == nil {
		return nil
	}
	return r.resp.ResultError()
}

func (r *HTTPResponse) fileName() string {
	if name := shared.FilenameFromContentDisposition(r.Header(string(HeaderContentDisposition))); name != "" {
		return name
	}
	if r.resp != nil && r.resp.Request != nil {
		requestURL := r.resp.Request.URL
		if parsed, err := url.Parse(requestURL); err == nil {
			_, name := filepath.Split(parsed.Path)
			return name
		}
		_, name := filepath.Split(requestURL)
		return name
	}
	return ""
}

func bytesReader(data []byte) io.Reader { return bytes.NewReader(data) }
