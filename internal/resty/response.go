package resty

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
func (r *HTTPResponse) SaveAs(dest string) (n int64, err error) {
	if r.resp == nil {
		return 0, HTTPErrorf("no response")
	}
	target := dest
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		fileName := r.fileName()
		if fileName == "" {
			fileName = "download.bin"
		}
		target = filepath.Join(dest, fileName)
	}
	f, err := os.Create(target) // #nosec G304 -- SaveAs intentionally writes to a caller-provided destination.
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

func (r *HTTPResponse) fileName() string {
	if cd := r.Header(string(HeaderContentDisposition)); cd != "" {
		if i := strings.Index(strings.ToLower(cd), "filename="); i >= 0 {
			name := strings.TrimSpace(cd[i+len("filename="):])
			name = strings.Trim(name, `"`)
			if idx := strings.Index(name, ";"); idx >= 0 {
				name = name[:idx]
			}
			if name != "" {
				return name
			}
		}
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
