package shared

// Method represents an HTTP request method.
type Method string

const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodHead    Method = "HEAD"
	MethodOptions Method = "OPTIONS"
	MethodPut     Method = "PUT"
	MethodDelete  Method = "DELETE"
	MethodTrace   Method = "TRACE"
	MethodConnect Method = "CONNECT"
	MethodPatch   Method = "PATCH"
)

// String returns the method string.
func (m Method) String() string { return string(m) }
