package vnet

import (
	"context"
	stdnet "net"
	"net/http"
	"time"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func IDNToASCII(unicode string) (string, error) { return netimpl.IDNToASCII(unicode) }

func WithResolveContext(ctx context.Context) ResolveOption { return netimpl.WithResolveContext(ctx) }

func WithResolveTimeout(timeout time.Duration) ResolveOption {
	return netimpl.WithResolveTimeout(timeout)
}

func WithResolveNetwork(network string) ResolveOption { return netimpl.WithResolveNetwork(network) }

func WithResolver(resolver *stdnet.Resolver) ResolveOption { return netimpl.WithResolver(resolver) }

func WithDNSTypes(attrNames ...string) ResolveOption { return netimpl.WithDNSTypes(attrNames...) }

func GetIPByHostWithOptions(hostName string, opts ...ResolveOption) ([]string, error) {
	return netimpl.GetIPByHostWithOptions(hostName, opts...)
}

func GetMultistageReverseProxyIP(ip string) string { return netimpl.GetMultistageReverseProxyIP(ip) }

func IsUnknown(checkString string) bool { return netimpl.IsUnknown(checkString) }

func ParseCookies(cookieStr string) []*http.Cookie { return netimpl.ParseCookies(cookieStr) }

// GetDNSInfo returns DNS records for hostName, optionally limited to attrNames record types.
func GetDNSInfo(hostName string, attrNames ...string) ([]string, error) {
	return GetDNSInfoWithOptions(hostName, WithDNSTypes(attrNames...))
}

func GetDNSInfoWithOptions(hostName string, opts ...ResolveOption) ([]string, error) {
	return netimpl.GetDNSInfoWithOptions(hostName, opts...)
}
