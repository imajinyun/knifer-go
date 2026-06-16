package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"
)

func TestNewClientRejectsBadConfig(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		opts []ClientOption
	}{
		{name: "empty host", host: "", port: 587},
		{name: "empty port", host: "smtp.example.com", port: 0},
		{name: "bad local name", host: "smtp.example.com", port: 587, opts: []ClientOption{WithLocalName("ok\nBAD")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.host, tt.port, tt.opts...)
			if err == nil {
				t.Fatal("NewClient() error = nil, want error")
			}
		})
	}
}

func TestClientOptionsAndProviderErrors(t *testing.T) {
	auth := testSMTPAuth{mechanism: "CUSTOM"}
	tlsConfig := &tls.Config{ServerName: "custom.example.com", MinVersion: tls.VersionTLS13}
	client, err := NewClient("smtp.example.com", 587,
		WithAuth("user", "pass"),
		WithSMTPAuth(auth),
		WithTLSConfig(tlsConfig),
		WithTLSPolicy(TLSPolicyUnknown),
		WithAllowPlainAuth(true),
		WithTimeout(time.Second),
		WithLocalName("mail.local"),
		WithDialContext(func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("dial blocked") }),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	tlsConfig.ServerName = "mutated.example.com"
	if client.config.Username != "user" || client.config.Password != "pass" || !client.config.AllowPlainAuth {
		t.Fatalf("auth/plain config = %#v", client.config)
	}
	if client.config.Auth == nil {
		t.Fatal("custom auth was not configured")
	}
	if client.config.TLSPolicy != TLSMandatoryStartTLS {
		t.Fatalf("TLSPolicy = %v, want %v", client.config.TLSPolicy, TLSMandatoryStartTLS)
	}
	if client.config.TLSConfig.ServerName != "custom.example.com" || client.config.TLSConfig.MinVersion != tls.VersionTLS13 {
		t.Fatalf("TLSConfig was not cloned: %#v", client.config.TLSConfig)
	}

	if _, err := NewClient("smtp.example.com", 587, WithTLSConfig(nil)); err != nil {
		t.Fatalf("NewClient(WithTLSConfig(nil)) error = %v", err)
	}
	if _, err := NewClient("smtp.example.com", 587, WithDialContext(nil)); err == nil {
		t.Fatal("NewClient(WithDialContext(nil)) error = nil, want error")
	}
	if _, err := NewClient("smtp.example.com", 587, WithSenderProvider(nil)); err == nil {
		t.Fatal("NewClient(WithSenderProvider(nil)) error = nil, want error")
	}

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	providerErr := errors.New("provider failed")
	client, err = NewClient("smtp.example.com", 587, WithSenderProvider(func(Config) (Sender, error) {
		return nil, providerErr
	}))
	if err != nil {
		t.Fatalf("NewClient(provider) error = %v", err)
	}
	var nilCtx context.Context
	if err := client.Send(nilCtx, msg); !errors.Is(err, providerErr) {
		t.Fatalf("Send(provider error) = %v, want %v", err, providerErr)
	}
}
