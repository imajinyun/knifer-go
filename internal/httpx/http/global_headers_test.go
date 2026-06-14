package http

import (
	"strings"
	"testing"
)

func TestGlobalHeadersDefault(t *testing.T) {
	headers := CloneGlobalHeaders()
	if headers.Get("User-Agent") == "" {
		t.Fatal("default UA missing")
	}
	if headers.Get("Accept") == "" {
		t.Fatal("default Accept missing")
	}
	if got := headers.Get("Accept-Encoding"); strings.Contains(got, "br") {
		t.Fatalf("default Accept-Encoding = %q should not advertise br without brotli decoding support", got)
	}
}

func TestGlobalHeadersSetAndRemove(t *testing.T) {
	SetGlobalHeader("X-Test", "v1")
	defer RemoveGlobalHeader("X-Test")

	headers := CloneGlobalHeaders()
	if headers.Get("X-Test") != "v1" {
		t.Fatalf("X-Test: %q", headers.Get("X-Test"))
	}

	RemoveGlobalHeader("X-Test")
	if got := CloneGlobalHeaders().Get("X-Test"); got != "" {
		t.Fatalf("after remove: %q", got)
	}
}
