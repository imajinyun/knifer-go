package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"
)

func TestSMTPClientContextCancelClosesConnection(t *testing.T) {
	server, err := newFakeSMTPServer(t, withFakeSMTPHangOnData())
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(), WithTLSPolicy(TLSNone), WithTimeout(0))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	canceled := make(chan error, 1)
	go func() { canceled <- client.Send(ctx, msg) }()
	server.WaitForDataCommand(t)
	cancel()
	select {
	case err := <-canceled:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Send() error = %v, want %v", err, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Fatal("Send() did not return after context cancellation")
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
}

func TestClientDialReusesConnectionWithReset(t *testing.T) {
	server, err := newFakeSMTPServer(t)
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	client, err := NewClient(server.Host(), server.Port(), WithTLSPolicy(TLSNone))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	sendCloser, err := client.Dial(context.Background())
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}

	first, err := NewMessage(WithFrom("from@example.com"), WithTo("first@example.com"), WithSubject("first"), WithText("first body"))
	if err != nil {
		t.Fatalf("NewMessage(first) error = %v", err)
	}
	second, err := NewMessage(WithFrom("from@example.com"), WithTo("second@example.com"), WithSubject("second"), WithText("second body"))
	if err != nil {
		t.Fatalf("NewMessage(second) error = %v", err)
	}
	if err := sendCloser.Send(context.Background(), first); err != nil {
		t.Fatalf("Send(first) error = %v", err)
	}
	if got := server.RSETCount(); got != 0 {
		t.Fatalf("RSET count after first send = %d, want 0", got)
	}
	if err := sendCloser.Send(context.Background(), second); err != nil {
		t.Fatalf("Send(second) error = %v", err)
	}
	if got := server.RSETCount(); got != 1 {
		t.Fatalf("RSET count after second send = %d, want 1", got)
	}
	if err := sendCloser.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := sendCloser.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if err := sendCloser.Send(context.Background(), second); err == nil {
		t.Fatal("Send() after Close() error = nil")
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
}

func TestSMTPHelpers(t *testing.T) {
	ctx, cancel := withClientTimeout(context.Background(), 0)
	if _, ok := ctx.Deadline(); ok {
		t.Fatal("withClientTimeout(0) unexpectedly set a deadline")
	}
	cancel()

	deadlineCtx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	ctx, cancel = withClientTimeout(deadlineCtx, time.Second)
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("withClientTimeout(existing deadline) removed deadline")
	}
	existingDeadline, _ := deadlineCtx.Deadline()
	if !deadline.Equal(existingDeadline) {
		t.Fatalf("deadline = %v, want %v", deadline, existingDeadline)
	}

	sender := smtpSender{config: Config{Host: "smtp.example.com"}}
	config := sender.tlsConfig()
	if config.ServerName != "smtp.example.com" || config.MinVersion != tls.VersionTLS12 {
		t.Fatalf("default tlsConfig = %#v", config)
	}

	original := &tls.Config{MinVersion: tls.VersionTLS13}
	sender.config.TLSConfig = original
	config = sender.tlsConfig()
	if config.ServerName != "smtp.example.com" || config.MinVersion != tls.VersionTLS13 {
		t.Fatalf("cloned tlsConfig = %#v", config)
	}
	config.ServerName = "mutated.example.com"
	if original.ServerName != "" {
		t.Fatalf("tlsConfig mutated original: %#v", original)
	}

	dialErr := errors.New("dial failed")
	sender.config.DialContext = func(context.Context, string, string) (net.Conn, error) { return nil, dialErr }
	if _, err := sender.dial(context.Background(), "smtp.example.com:587", config); !errors.Is(err, dialErr) {
		t.Fatalf("dial() error = %v, want %v", err, dialErr)
	}
}
