package shared

// Header defines common HTTP header names.
type Header string

const (
	// General headers.
	HeaderAuthorization      Header = "Authorization"
	HeaderProxyAuthorization Header = "Proxy-Authorization"
	HeaderDate               Header = "Date"
	HeaderConnection         Header = "Connection"
	HeaderMimeVersion        Header = "MIME-Version"
	HeaderTrailer            Header = "Trailer"
	HeaderTransferEncoding   Header = "Transfer-Encoding"
	HeaderUpgrade            Header = "Upgrade"
	HeaderVia                Header = "Via"
	HeaderCacheControl       Header = "Cache-Control"
	HeaderPragma             Header = "Pragma"
	HeaderContentType        Header = "Content-Type"

	// Request headers.
	HeaderHost           Header = "Host"
	HeaderReferer        Header = "Referer"
	HeaderOrigin         Header = "Origin"
	HeaderUserAgent      Header = "User-Agent"
	HeaderAccept         Header = "Accept"
	HeaderAcceptLanguage Header = "Accept-Language"
	HeaderAcceptEncoding Header = "Accept-Encoding"
	HeaderAcceptCharset  Header = "Accept-Charset"
	HeaderCookie         Header = "Cookie"
	HeaderContentLength  Header = "Content-Length"

	// Response headers.
	HeaderWWWAuthenticate    Header = "WWW-Authenticate"
	HeaderSetCookie          Header = "Set-Cookie"
	HeaderContentEncoding    Header = "Content-Encoding"
	HeaderContentDisposition Header = "Content-Disposition"
	HeaderETag               Header = "ETag"
	HeaderLocation           Header = "Location"
)

// String returns the header name.
func (h Header) String() string { return string(h) }
