package shared

import (
	"encoding/base64"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// CharsetPattern matches charset in Content-Type.
	CharsetPattern = regexp.MustCompile(`(?i)charset\s*=\s*([a-z0-9-]+)`)
	// MetaCharsetPattern matches charset in HTML meta tags.
	MetaCharsetPattern = regexp.MustCompile(`(?i)<meta[^>]*?charset\s*=\s*['"]?([a-z0-9-]+)`)
)

// BuildBasicAuth builds a Basic Auth string.
func BuildBasicAuth(user, pass string) string {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + token
}

// GetCharsetFromContentType extracts charset from Content-Type.
func GetCharsetFromContentType(ct string) string {
	m := CharsetPattern.FindStringSubmatch(ct)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// GetCharsetFromHTML extracts charset from HTML meta tags.
func GetCharsetFromHTML(html string) string {
	m := MetaCharsetPattern.FindStringSubmatch(html)
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

// FilenameFromContentDisposition extracts a filename from a Content-Disposition header.
func FilenameFromContentDisposition(cd string) string {
	if cd == "" {
		return ""
	}
	if i := strings.Index(strings.ToLower(cd), "filename="); i >= 0 {
		name := strings.TrimSpace(cd[i+len("filename="):])
		name = strings.Trim(name, `"`)
		if idx := strings.Index(name, ";"); idx >= 0 {
			name = name[:idx]
		}
		return strings.TrimSpace(name)
	}
	return ""
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
