package vhttp

import (
	"net/http"

	httpx "github.com/imajinyun/knifer-go/internal/httpx/http"
)

// SetCookieJar delegates to the internal httpx implementation.
func SetCookieJar(jar http.CookieJar) {
	httpx.SetCookieJar(jar)
}

// GetCookieJar delegates to the internal httpx implementation.
func GetCookieJar() http.CookieJar {
	return httpx.GetCookieJar()
}

// CloseCookie delegates to the internal httpx implementation.
func CloseCookie() {
	httpx.CloseCookie()
}
