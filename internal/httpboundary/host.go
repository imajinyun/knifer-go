// Package httpboundary centralizes HTTP trust-boundary host classification.
package httpboundary

import (
	"context"
	"errors"
	"net"
	"strings"
)

var (
	ErrPrivateHost = errors.New("httpboundary: private host")
	ErrNoAddresses = errors.New("httpboundary: no addresses")
)

type LookupIPFunc func(context.Context, string) ([]net.IP, error)

func DefaultLookupIP(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

func IsPrivateIP(ip net.IP) bool {
	return ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate()
}

func IsPrivateHost(ctx context.Context, lookupIP LookupIPFunc, host string) (bool, error) {
	if strings.EqualFold(host, "localhost") {
		return true, nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return IsPrivateIP(ip), nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if lookupIP == nil {
		lookupIP = DefaultLookupIP
	}
	ips, err := lookupIP(ctx, host)
	if err != nil {
		return false, err
	}
	for _, ip := range ips {
		if IsPrivateIP(ip) {
			return true, nil
		}
	}
	return false, nil
}

func PublicHostIPs(ctx context.Context, lookupIP LookupIPFunc, host string) ([]net.IP, error) {
	if strings.EqualFold(host, "localhost") {
		return nil, ErrPrivateHost
	}
	if ip := net.ParseIP(host); ip != nil {
		if IsPrivateIP(ip) {
			return nil, ErrPrivateHost
		}
		return []net.IP{ip}, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if lookupIP == nil {
		lookupIP = DefaultLookupIP
	}
	ips, err := lookupIP(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, ErrNoAddresses
	}
	public := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if IsPrivateIP(ip) {
			return nil, ErrPrivateHost
		}
		public = append(public, ip)
	}
	return public, nil
}
