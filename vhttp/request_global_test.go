package vhttp_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeRequestGlobalAccessors(t *testing.T) {
	previous := vhttp.SnapshotGlobalConfig()
	defer vhttp.ConfigureGlobalConfig(previous)
	vhttp.SetGlobalMaxRedirects(3)
	vhttp.SetGlobalMaxResponseBytes(99)
	vhttp.SetGlobalFollowRedirects(false)
	vhttp.SetGlobalUserAgent("vhttp-extra/1.0")
	vhttp.SetIgnoreEOFError(true)
	vhttp.SetGlobalBoundary("boundary-extra")
	vhttp.SetGlobalDecodeURL(true)
	if vhttp.GetGlobalMaxRedirects() != 3 || vhttp.GetGlobalMaxResponseBytes() != 99 || vhttp.GetGlobalFollowRedirects() || vhttp.GetGlobalUserAgent() != "vhttp-extra/1.0" || !vhttp.IsIgnoreEOFError() || vhttp.GetGlobalBoundary() != "boundary-extra" || !vhttp.IsGlobalDecodeURL() {
		t.Fatalf("global accessors snapshot = %#v", vhttp.SnapshotGlobalConfig())
	}
}
