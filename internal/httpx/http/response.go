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
	"sync"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/internal/httpx/internal/shared"
)

// HTTPResponse wraps http.Response and provides convenient readers, aligned with the utility toolkit-http HttpResponse.
type HTTPResponse struct {
	resp         *http.Response
	mu           sync.Mutex
	body         []byte
	bodyRead     bool
	bodyConsumed bool
	err          error
	decodeConfig responseDecodeConfig
}

func wrapResponse(r *http.Response, cfg responseDecodeConfig) *HTTPResponse {
	return &HTTPResponse{resp: r, decodeConfig: cfg.normalized()}
}

// ContentDecoder decodes a response body for a Content-Encoding value.
type ContentDecoder func(io.Reader) (io.ReadCloser, error)

type responseDecodeConfig struct {
	autoDecode     bool
	decoders       map[string]ContentDecoder
	maxBytes       int64
	readAll        func(io.Reader) ([]byte, error)
	ignoreEOFError bool
}

func defaultResponseDecodeConfig() responseDecodeConfig {
	return responseDecodeConfig{
		autoDecode:     true,
		ignoreEOFError: true,
		decoders: map[string]ContentDecoder{
			"gzip":    gzipDecoder,
			"deflate": deflateDecoder,
		},
	}
}

func responseDecodeConfigFromGlobal(cfg GlobalConfig) responseDecodeConfig {
	decodeConfig := defaultResponseDecodeConfig()
	decodeConfig.maxBytes = cfg.MaxResponseBytes
	decodeConfig.ignoreEOFError = cfg.IgnoreEOFError
	return decodeConfig
}

func (c responseDecodeConfig) normalized() responseDecodeConfig {
	if c.readAll == nil {
		c.readAll = io.ReadAll
	}
	if c.decoders == nil {
		c.decoders = defaultResponseDecodeConfig().decoders
		return c
	}
	cloned := make(map[string]ContentDecoder, len(c.decoders))
	for enc, decoder := range c.decoders {
		if decoder != nil {
			cloned[enc] = decoder
		}
	}
	c.decoders = cloned
	return c
}

func (c *responseDecodeConfig) setDecoder(encoding string, decoder ContentDecoder) {
	encoding = normalizeEncoding(encoding)
	if encoding == "" {
		return
	}
	if c.decoders == nil {
		c.decoders = defaultResponseDecodeConfig().decoders
	}
	cloned := make(map[string]ContentDecoder, len(c.decoders)+1)
	for enc, existing := range c.decoders {
		cloned[enc] = existing
	}
	if decoder == nil {
		delete(cloned, encoding)
	} else {
		cloned[encoding] = decoder
	}
	c.decoders = cloned
}

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
		return nil, HTTPErrorfWithCode(knifer.ErrCodeUnsupported, "response body exceeds max bytes: %d", maxBytes)
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
func (r *HTTPResponse) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}

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
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.bodyRead {
		return r.body
	}
	if r.bodyConsumed {
		if r.err == nil {
			r.err = HTTPErrorfWithCode(knifer.ErrCodeUnsupported, "response body has already been consumed")
		}
		return nil
	}
	if r.resp == nil || r.resp.Body == nil {
		return nil
	}
	defer func() {
		if err := r.resp.Body.Close(); err != nil && r.err == nil {
			r.err = NewHTTPError("close response body failed", err)
		}
	}()
	reader, err := r.decodedBody()
	if err != nil {
		r.err = err
		return nil
	}
	data, err := readAllWithLimit(reader, r.decodeConfig.maxBytes, r.decodeConfig.readAll)
	if err != nil && (!r.decodeConfig.ignoreEOFError || err != io.ErrUnexpectedEOF) {
		r.err = NewHTTPError("read response body failed", err)
		return nil
	}
	r.body = data
	r.bodyRead = true
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
	// #nosec G304 -- SaveAs intentionally writes to the caller-provided path after
	// optional directory resolution; callers control the download destination.
	f, err := cfg.openFile(target, flag, cfg.filePerm)
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
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return 0, r.err
	}
	if r.bodyRead {
		return io.Copy(w, bytes.NewReader(r.body))
	}
	if r.bodyConsumed {
		r.err = HTTPErrorfWithCode(knifer.ErrCodeUnsupported, "response body has already been consumed")
		return 0, r.err
	}
	if r.resp == nil || r.resp.Body == nil {
		return 0, nil
	}
	defer func() {
		if err := r.resp.Body.Close(); err != nil && r.err == nil {
			r.err = NewHTTPError("close response body failed", err)
		}
	}()
	reader, err := r.decodedBody()
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
	if err != nil && (!r.decodeConfig.ignoreEOFError || err != io.ErrUnexpectedEOF) {
		r.err = NewHTTPError("read response body failed", err)
		return n, r.err
	}
	r.bodyConsumed = true
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

func (r *HTTPResponse) decodedBody() (io.Reader, error) {
	if r.resp == nil || r.resp.Body == nil || !r.decodeConfig.autoDecode {
		return r.resp.Body, nil
	}
	enc := normalizeEncoding(r.resp.Header.Get(string(HeaderContentEncoding)))
	decoder := r.decodeConfig.decoders[enc]
	if decoder == nil {
		return r.resp.Body, nil
	}
	reader, err := decoder(r.resp.Body)
	if err != nil {
		if enc == "gzip" && err == io.EOF {
			// Some servers may declare gzip without compressing; try to fall back.
			return bytes.NewReader(nil), nil
		}
		return nil, NewHTTPError(enc+" reader init failed", err)
	}
	return reader, nil
}

func gzipDecoder(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) }

func deflateDecoder(r io.Reader) (io.ReadCloser, error) { return zlib.NewReader(r) }

func normalizeEncoding(encoding string) string {
	return shared.NormalizeEncoding(encoding)
}

// charsetFromContentType extracts the charset from Content-Type.
func charsetFromContentType(ct string) string {
	return shared.GetCharsetFromContentType(ct)
}
