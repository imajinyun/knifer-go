package http

import (
	"net/http"
	"testing"
)

func TestGlobalCookieJar(t *testing.T) {
	jar := GetCookieJar()
	if jar == nil {
		t.Fatal("default jar should not be nil")
	}
	CloseCookie()
	if GetCookieJar() != nil {
		t.Fatal("after close should be nil")
	}
	// Restore the default jar.
	SetCookieJar(jar)
	if GetCookieJar() == nil {
		t.Fatal("restored jar nil")
	}

	// Customize the jar.
	var custom http.CookieJar
	SetCookieJar(custom)
	if GetCookieJar() != nil {
		t.Fatal("custom nil jar")
	}
	SetCookieJar(jar)
}
