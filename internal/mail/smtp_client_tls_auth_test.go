package mail

import (
	"context"
	"crypto/tls"
	"strings"
	"testing"
	"time"
)

func TestSMTPClientUsesCustomAuth(t *testing.T) {
	server, err := newFakeSMTPServer(t, withFakeSMTPAuth("CUSTOM", "token", true))
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(),
		WithTLSPolicy(TLSNone),
		WithAllowPlainAuth(true),
		WithSMTPAuth(testSMTPAuth{mechanism: "CUSTOM", initial: []byte("token")}),
		WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
	if !server.Authenticated() {
		t.Fatal("server did not receive successful custom AUTH")
	}
}

func TestSMTPClientReturnsAuthFailure(t *testing.T) {
	server, err := newFakeSMTPServer(t, withFakeSMTPAuth("CUSTOM", "token", false))
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(),
		WithTLSPolicy(TLSNone),
		WithAllowPlainAuth(true),
		WithSMTPAuth(testSMTPAuth{mechanism: "CUSTOM", initial: []byte("token")}),
		WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); err == nil || !strings.Contains(err.Error(), "smtp auth") {
		t.Fatalf("Send() error = %v, want smtp auth error", err)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
}

func TestSMTPClientStartTLS(t *testing.T) {
	cert := newTestCertificate(t)
	server, err := newFakeSMTPServer(t, withFakeSMTPStartTLS(cert))
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("secure body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(),
		WithTLSConfig(&tls.Config{RootCAs: cert.pool, ServerName: "localhost", MinVersion: tls.VersionTLS12}),
		WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
	if !server.TLSActive() || !strings.Contains(server.Data(), "secure body") {
		t.Fatalf("TLSActive=%v DATA=%q", server.TLSActive(), server.Data())
	}
}

func TestSMTPClientImplicitTLS(t *testing.T) {
	cert := newTestCertificate(t)
	server, err := newFakeSMTPServer(t, withFakeSMTPImplicitTLS(cert))
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("implicit body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(),
		WithTLSPolicy(TLSImplicit),
		WithTLSConfig(&tls.Config{RootCAs: cert.pool, ServerName: "localhost", MinVersion: tls.VersionTLS12}),
		WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
	if !server.TLSActive() || !strings.Contains(server.Data(), "implicit body") {
		t.Fatalf("TLSActive=%v DATA=%q", server.TLSActive(), server.Data())
	}
}
