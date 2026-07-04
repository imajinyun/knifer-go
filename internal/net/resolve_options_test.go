package net

import (
	"context"
	stdnet "net"
	"testing"
	"time"
)

func TestResolveWithOptions(t *testing.T) {
	ips, err := GetIPByHostWithOptions("localhost", WithResolveNetwork("ip4"), WithResolveTimeout(time.Second))
	if err != nil {
		t.Fatalf("GetIPByHostWithOptions: %v", err)
	}
	if len(ips) == 0 {
		t.Fatal("GetIPByHostWithOptions returned no IPs")
	}
	dns, err := GetDNSInfoWithOptions("localhost", WithDNSTypes("A"), WithResolveTimeout(time.Second))
	if err != nil {
		t.Fatalf("GetDNSInfoWithOptions: %v", err)
	}
	if len(dns) == 0 {
		t.Fatal("GetDNSInfoWithOptions returned no A records")
	}
}

func TestNilResolverOptionDoesNotClearPreviousResolver(t *testing.T) {
	resolver := &stdnet.Resolver{
		Dial: func(context.Context, string, string) (stdnet.Conn, error) {
			return nil, stdnet.ErrClosed
		},
	}
	cfg, cancel := applyResolveOptions([]ResolveOption{
		WithResolver(resolver),
		WithResolver(nil),
	})
	defer cancel()
	if cfg.resolver != resolver {
		t.Fatal("nil WithResolver cleared previous resolver")
	}
}
