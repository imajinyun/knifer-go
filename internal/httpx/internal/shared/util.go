package shared

import (
	"encoding/base64"
	"path/filepath"
	"regexp"
	"strings"

	knifer "github.com/imajinyun/knifer-go"
)

var (
	// CharsetPattern matches charset in Content-Type.
	CharsetPattern = regexp.MustCompile(`(?i)charset\s*=\s*([a-z0-9-]+)`)
	// MetaCharsetPattern matches charset in HTML meta tags.
	MetaCharsetPattern = regexp.MustCompile(`(?i)<meta[^>]*?charset\s*=\s*['"]?([a-z0-9-]+)`)
)

type charsetConfig struct {
	charsetRe *regexp.Regexp
	metaRe    *regexp.Regexp
}

// CharsetOption customizes charset extraction helpers per call.
type CharsetOption func(*charsetConfig)

// WithCharsetRegexp sets the regexp used by GetCharsetFromContentTypeWithOptions.
func WithCharsetRegexp(re *regexp.Regexp) CharsetOption {
	return func(c *charsetConfig) { c.charsetRe = re }
}

// WithMetaCharsetRegexp sets the regexp used by GetCharsetFromHTMLWithOptions.
func WithMetaCharsetRegexp(re *regexp.Regexp) CharsetOption {
	return func(c *charsetConfig) { c.metaRe = re }
}

func applyCharsetOptions(opts []CharsetOption) charsetConfig {
	cfg := charsetConfig{charsetRe: CharsetPattern, metaRe: MetaCharsetPattern}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.charsetRe == nil {
		cfg.charsetRe = CharsetPattern
	}
	if cfg.metaRe == nil {
		cfg.metaRe = MetaCharsetPattern
	}
	return cfg
}

// BuildBasicAuth builds a Basic Auth string.
func BuildBasicAuth(user, pass string) string {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + token
}

// GetCharsetFromContentType extracts charset from Content-Type.
func GetCharsetFromContentType(ct string) string {
	return GetCharsetFromContentTypeWithOptions(ct)
}

// GetCharsetFromContentTypeWithOptions extracts charset from Content-Type with options.
func GetCharsetFromContentTypeWithOptions(ct string, opts ...CharsetOption) string {
	m := applyCharsetOptions(opts).charsetRe.FindStringSubmatch(ct)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// GetCharsetFromHTML extracts charset from HTML meta tags.
func GetCharsetFromHTML(html string) string {
	return GetCharsetFromHTMLWithOptions(html)
}

// GetCharsetFromHTMLWithOptions extracts charset from HTML meta tags with options.
func GetCharsetFromHTMLWithOptions(html string, opts ...CharsetOption) string {
	m := applyCharsetOptions(opts).metaRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// GetMimeType returns the MIME type by file extension, or an empty string when unknown.
func GetMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return ""
	}
	return mimeTypes[ext]
}

// NormalizeEncoding normalizes HTTP content-encoding tokens.
func NormalizeEncoding(encoding string) string {
	return strings.ToLower(strings.TrimSpace(encoding))
}

// FilenameFromContentDisposition extracts a filename from a Content-Disposition header.
func FilenameFromContentDisposition(cd string) string {
	if cd == "" {
		return ""
	}
	if i := strings.Index(strings.ToLower(cd), "filename="); i >= 0 {
		name := strings.TrimSpace(cd[i+len("filename="):])
		if strings.HasPrefix(name, `"`) {
			name = strings.TrimPrefix(name, `"`)
			if idx := strings.Index(name, `"`); idx >= 0 {
				return strings.TrimSpace(name[:idx])
			}
		}
		if idx := strings.Index(name, ";"); idx >= 0 {
			name = name[:idx]
		}
		return strings.Trim(strings.TrimSpace(name), `"`)
	}
	return ""
}

// SafeDownloadedFilename validates an automatically discovered download file name.
// It rejects absolute paths, path separators, and parent-directory references so
// server-provided Content-Disposition values cannot escape the caller's target directory.
func SafeDownloadedFilename(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil
	}
	if filepath.IsAbs(name) || strings.Contains(name, "/") || strings.Contains(name, `\`) {
		return "", HTTPErrorfWithCode(knifer.ErrCodeInvalidInput, "unsafe download filename: %q", name)
	}
	clean := filepath.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", HTTPErrorfWithCode(knifer.ErrCodeInvalidInput, "unsafe download filename: %q", name)
	}
	base := filepath.Base(clean)
	if base != clean || base == "." || base == ".." {
		return "", HTTPErrorfWithCode(knifer.ErrCodeInvalidInput, "unsafe download filename: %q", name)
	}
	return base, nil
}

// SafeJoinDownloadPath joins a sanitized download file name under dir and verifies
// the resulting absolute path remains inside dir.
func SafeJoinDownloadPath(dir, fileName string) (string, error) {
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return "", NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "resolve download directory failed", err)
	}
	target := filepath.Join(dirAbs, fileName)
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "resolve download target failed", err)
	}
	rel, err := filepath.Rel(dirAbs, targetAbs)
	if err != nil {
		return "", NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "validate download target failed", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", HTTPErrorfWithCode(knifer.ErrCodeInvalidInput, "download target escapes destination directory: %q", fileName)
	}
	return targetAbs, nil
}

var mimeTypes = map[string]string{
	".html": "text/html",
	".htm":  "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".pdf":  "application/pdf",
	".zip":  "application/zip",
	".gz":   "application/gzip",
	".tar":  "application/x-tar",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".mp4":  "video/mp4",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
}
