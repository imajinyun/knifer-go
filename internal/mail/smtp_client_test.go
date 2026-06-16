package mail

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSMTPClientSendAgainstFakeServer(t *testing.T) {
	server, err := newFakeSMTPServer(t)
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithSubject("hello"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(), WithTLSPolicy(TLSNone), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
	if !strings.Contains(server.Data(), "Subject: hello") || !strings.Contains(server.Data(), "body") {
		t.Fatalf("SMTP DATA = %q", server.Data())
	}
}

func TestSMTPClientUsesEnvelopeSenderAndDedupedRecipients(t *testing.T) {
	server, err := newFakeSMTPServer(t)
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(
		WithFrom("header@example.com"),
		WithEnvelopeFrom("bounce@example.com"),
		WithTo("to@example.com", "TO@example.com"),
		WithCc("cc@example.com", "to@example.com"),
		WithBcc("hidden@example.com", "cc@example.com"),
		WithText("body"),
	)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(), WithTLSPolicy(TLSNone), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
	if got := server.MailFrom(); got != "<bounce@example.com>" {
		t.Fatalf("MAIL FROM = %q, want bounce envelope", got)
	}
	wantRecipients := []string{"<to@example.com>", "<cc@example.com>", "<hidden@example.com>"}
	if got := server.RcptTo(); strings.Join(got, ",") != strings.Join(wantRecipients, ",") {
		t.Fatalf("RCPT TO = %v, want %v", got, wantRecipients)
	}
	if recipients := msg.Recipients(); strings.Join(recipients, ",") != "to@example.com,cc@example.com,hidden@example.com" {
		t.Fatalf("Message.Recipients() = %v", recipients)
	}
}

func TestSMTPClientRequiresStartTLS(t *testing.T) {
	server, err := newFakeSMTPServer(t)
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); !errors.Is(err, ErrTLSRequired) {
		t.Fatalf("Send() error = %v, want %v", err, ErrTLSRequired)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
}

func TestSMTPClientRejectsPlainAuthWithoutTLS(t *testing.T) {
	server, err := newFakeSMTPServer(t)
	if err != nil {
		t.Fatalf("newFakeSMTPServer() error = %v", err)
	}
	defer server.Close()

	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	client, err := NewClient(server.Host(), server.Port(), WithTLSPolicy(TLSNone), WithAuth("user", "pass"), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Send(context.Background(), msg); !errors.Is(err, ErrPlainAuth) {
		t.Fatalf("Send() error = %v, want %v", err, ErrPlainAuth)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("fake SMTP server error = %v", err)
	}
}
