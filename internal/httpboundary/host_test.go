package httpboundary

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestIsPrivateHost(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		{name: "localhost", host: "localhost", want: true},
		{name: "loopback", host: "127.0.0.1", want: true},
		{name: "private", host: "10.0.0.1", want: true},
		{name: "unspecified", host: "0.0.0.0", want: true},
		{name: "multicast", host: "224.0.0.1", want: true},
		{name: "public", host: "93.184.216.34", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPrivateHost(context.Background(), nil, tt.host)
			if err != nil {
				t.Fatalf("IsPrivateHost() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsPrivateHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestPublicHostIPs(t *testing.T) {
	public, err := PublicHostIPs(context.Background(), nil, "93.184.216.34")
	if err != nil || len(public) != 1 || public[0].String() != "93.184.216.34" {
		t.Fatalf("PublicHostIPs direct public = %#v, %v", public, err)
	}
	if _, err := PublicHostIPs(context.Background(), nil, "127.0.0.1"); !errors.Is(err, ErrPrivateHost) {
		t.Fatalf("PublicHostIPs private error = %v, want ErrPrivateHost", err)
	}
	if _, err := PublicHostIPs(context.Background(), func(context.Context, string) ([]net.IP, error) {
		return nil, nil
	}, "example.com"); !errors.Is(err, ErrNoAddresses) {
		t.Fatalf("PublicHostIPs no addresses error = %v, want ErrNoAddresses", err)
	}
	if _, err := PublicHostIPs(context.Background(), func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}, "example.com"); err != nil {
		t.Fatalf("PublicHostIPs lookup public error = %v", err)
	}
	if _, err := PublicHostIPs(context.Background(), func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.1")}, nil
	}, "example.com"); !errors.Is(err, ErrPrivateHost) {
		t.Fatalf("PublicHostIPs lookup private error = %v, want ErrPrivateHost", err)
	}
}
