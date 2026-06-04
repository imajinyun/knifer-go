package http

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/imajinyun/go-knifer/internal/httpx/internal/shared"
)

// HTTPResponse wraps http.Response and provides convenient readers, aligned with the utility toolkit-http HttpResponse.
type HTTPResponse struct {
	resp     *http.Response
	body     []byte
	bodyRead bool
	once     sync.Once
	err      error
}

func wrapResponse(r *http.Response) *HTTPResponse { return &HTTPResponse{resp: r} }

type saveConfig struct {
	filePerm        fs.FileMode
	dirPerm         fs.FileMode
	overwrite       bool
	createParents   bool
	defaultFilename string
}

// SaveOption customizes response file saving.
type SaveOption func(*saveConfig)

func defaultSaveConfig() saveConfig {
	return saveConfig{filePerm: 0o644, dirPerm: 0o750, overwrite: true, createParents: true, defaultFilename: "download.bin"}
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

func applySaveOptions(opts []SaveOption) saveConfig {
	cfg := defaultSaveConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
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
	return r.resp.StatusCode
}

// IsOK reports whether the status is a 2xx success.
func (r *HTTPResponse) IsOK() bool {
	return r.Status() >= 200 && r.Status() < 300
}

// Header returns a response header value.
func (r *HTTPResponse) Header(name string) string {
	if r.resp == nil {
		return ""
	}
	return r.resp.Header.Get(name)
}

// Headers returns all response headers.
func (r *HTTPResponse) Headers() http.Header {
	if r.resp == nil {
		return nil
	}
	return r.resp.Header
}

// Cookies returns cookies from the response.
func (r *HTTPResponse) Cookies() []*http.Cookie {
	if r.resp == nil {
		return nil
	}
	return r.resp.Cookies()
}

// ContentType returns the response Content-Type.
func (r *HTTPResponse) ContentType() string { return r.Header(string(HeaderContentType)) }

// ContentLength returns the response Content-Length.
func (r *HTTPResponse) ContentLength() int64 {
	if r.resp == nil {
		return -1
	}
	return r.resp.ContentLength
}

// Charset parses the charset from Content-Type and returns UTF-8 when unspecified.
func (r *HTTPResponse) Charset() string {
	if cs := charsetFromContentType(r.ContentType()); cs != "" {
		return cs
	}
	return "UTF-8"
}

// Bytes reads and returns the response body bytes.
func (r *HTTPResponse) Bytes() []byte {
	if r.bodyRead {
		return r.body
	}
	r.once.Do(func() {
		if r.resp == nil || r.resp.Body == nil {
			return
		}
		defer func() {
			if err := r.resp.Body.Close(); err != nil && r.err == nil {
				r.err = NewHTTPError("close response body failed", err)
			}
		}()
		reader, err := decodedBody(r.resp)
		if err != nil {
			r.err = err
			return
		}
		data, err := io.ReadAll(reader)
		if err != nil && (!IsIgnoreEOFError() || err != io.ErrUnexpectedEOF) {
			r.err = NewHTTPError("read response body failed", err)
			return
		}
		r.body = data
		r.bodyRead = true
	})
	return r.body
}

// Body reads the response body and returns it as a string.
func (r *HTTPResponse) Body() string { return string(r.Bytes()) }

// WriteTo writes the response body to the writer and returns the number of bytes written.
func (r *HTTPResponse) WriteTo(w io.Writer) (int64, error) {
	data := r.Bytes()
	if r.err != nil {
		return 0, r.err
	}
	n, err := w.Write(data)
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
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		fileName := r.fileName()
		if fileName == "" {
			fileName = cfg.defaultFilename
		}
		target = filepath.Join(dest, fileName)
	}
	if cfg.createParents {
		if err := os.MkdirAll(filepath.Dir(target), cfg.dirPerm); err != nil {
			return 0, NewHTTPError("create parent directory failed", err)
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	// #nosec G304 -- SaveAs intentionally writes to the caller-provided path after
	// optional directory resolution; callers control the download destination.
	f, err := os.OpenFile(target, flag, cfg.filePerm)
	if err != nil {
		return 0, NewHTTPError("create file failed", err)
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	return r.writeBodyTo(f)
}

// Close closes the underlying response body; this is only needed when the body has not been read.
func (r *HTTPResponse) Close() error {
	if r.resp != nil && r.resp.Body != nil {
		return r.resp.Body.Close()
	}
	return nil
}

// Raw returns the original *http.Response for streaming; remember to close Body manually.
func (r *HTTPResponse) Raw() *http.Response { return r.resp }

func (r *HTTPResponse) writeBodyTo(w io.Writer) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	if r.bodyRead {
		return io.Copy(w, bytes.NewReader(r.body))
	}
	if r.resp == nil || r.resp.Body == nil {
		return 0, nil
	}
	defer func() {
		if err := r.resp.Body.Close(); err != nil && r.err == nil {
			r.err = NewHTTPError("close response body failed", err)
		}
	}()
	reader, err := decodedBody(r.resp)
	if err != nil {
		r.err = err
		return 0, err
	}
	if closer, ok := reader.(io.Closer); ok && closer != r.resp.Body {
		defer func() {
			if closeErr := closer.Close(); closeErr != nil && r.err == nil {
				r.err = NewHTTPError("close decoded body failed", closeErr)
			}
		}()
	}
	n, err := io.Copy(w, reader)
	if err != nil && (!IsIgnoreEOFError() || err != io.ErrUnexpectedEOF) {
		r.err = NewHTTPError("read response body failed", err)
		return n, r.err
	}
	r.bodyRead = true
	return n, nil
}

func (r *HTTPResponse) fileName() string {
	if name := shared.FilenameFromContentDisposition(r.Header(string(HeaderContentDisposition))); name != "" {
		return name
	}
	if r.resp != nil && r.resp.Request != nil && r.resp.Request.URL != nil {
		_, name := filepath.Split(r.resp.Request.URL.Path)
		return name
	}
	return ""
}

func decodedBody(resp *http.Response) (io.Reader, error) {
	enc := strings.ToLower(resp.Header.Get(string(HeaderContentEncoding)))
	switch enc {
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			// Some servers may declare gzip without compressing; try to fall back.
			if err == io.EOF {
				return bytes.NewReader(nil), nil
			}
			return nil, NewHTTPError("gzip reader init failed", err)
		}
		return gr, nil
	case "deflate":
		zr, err := zlib.NewReader(resp.Body)
		if err != nil {
			return nil, NewHTTPError("deflate reader init failed", err)
		}
		return zr, nil
	default:
		return resp.Body, nil
	}
}

var charsetRegex = regexp.MustCompile(`(?i)charset\s*=\s*([a-z0-9-]+)`)

// charsetFromContentType extracts the charset from Content-Type.
func charsetFromContentType(ct string) string {
	if ct == "" {
		return ""
	}
	m := charsetRegex.FindStringSubmatch(ct)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
