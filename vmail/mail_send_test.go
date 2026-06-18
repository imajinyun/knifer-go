package vmail

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"
)

func TestSendTextUsesInjectedProvider(t *testing.T) {
	var got *Message
	err := SendText(
		context.Background(),
		"smtp.example.com",
		587,
		"from@example.com",
		[]string{"to@example.com"},
		"subject",
		"body",
		WithSenderProvider(func(config Config) (Sender, error) {
			if config.Host != "smtp.example.com" || config.Port != 587 {
				t.Fatalf("Config = %#v", config)
			}
			return SenderFunc(func(ctx context.Context, message *Message) error {
				got = message
				return nil
			}), nil
		}),
	)
	if err != nil {
		t.Fatalf("SendText() error = %v", err)
	}
	if got == nil || got.Subject != "subject" || got.Text != "body" {
		t.Fatalf("sent message = %#v", got)
	}
}

func TestAccountQuickSendFacade(t *testing.T) {
	account := Account{
		Host:           "smtp.example.com",
		Port:           587,
		Username:       "user@example.com",
		Password:       "secret",
		From:           "from@example.com",
		FromName:       "Facade Sender",
		TLSPolicy:      TLSNone,
		AllowPlainAuth: true,
		Timeout:        time.Second,
	}

	var got *Message
	provider := func(config Config) (Sender, error) {
		if config.Host != "smtp.example.com" || config.Port != 587 {
			t.Fatalf("Config address = %#v", config)
		}
		if config.Username != "user@example.com" || config.Password != "secret" || !config.AllowPlainAuth {
			t.Fatalf("Config auth = %#v", config)
		}
		return SenderFunc(func(ctx context.Context, message *Message) error {
			got = message
			return nil
		}), nil
	}

	if err := SendAccountHTML(
		context.Background(),
		account,
		[]string{"to@example.com"},
		"subject",
		"<p>html</p>",
		WithQuickMessageOptions(WithHeader("X-Facade-Quick", "yes")),
		WithQuickClientOptions(WithSenderProvider(provider)),
	); err != nil {
		t.Fatalf("SendAccountHTML() error = %v", err)
	}
	if got == nil || got.From.Name != "Facade Sender" || got.HTML != "<p>html</p>" {
		t.Fatalf("sent message = %#v", got)
	}
	if values := got.Headers.Values("X-Facade-Quick"); len(values) != 1 || values[0] != "yes" {
		t.Fatalf("X-Facade-Quick = %v, want yes", values)
	}

	got = nil
	if err := QuickSend(
		context.Background(),
		account,
		WithQuickMessageOptions(WithTo("to@example.com"), WithSubject("quick"), WithText("body")),
		WithQuickClientOptions(WithSenderProvider(provider)),
	); err != nil {
		t.Fatalf("QuickSend() error = %v", err)
	}
	if got == nil || got.Subject != "quick" || got.Text != "body" {
		t.Fatalf("QuickSend message = %#v", got)
	}
}

func TestFacadeSendAndClientOptions(t *testing.T) {
	message, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	var got *Message
	auth := facadeSMTPAuth{mechanism: "CUSTOM"}
	provider := func(config Config) (Sender, error) {
		if config.Username != "user" || config.Password != "pass" || !config.AllowPlainAuth {
			t.Fatalf("Config auth = %#v", config)
		}
		if config.Auth == nil {
			t.Fatal("Config Auth is nil")
		}
		if config.TLSPolicy != TLSNone || config.LocalName != "mail.local" || config.Timeout != time.Second {
			t.Fatalf("Config transport = %#v", config)
		}
		if config.TLSConfig == nil || config.TLSConfig.ServerName != "smtp.example.com" {
			t.Fatalf("Config TLS = %#v", config.TLSConfig)
		}
		return SenderFunc(func(ctx context.Context, message *Message) error {
			got = message
			return nil
		}), nil
	}
	if err := Send(context.Background(), "smtp.example.com", 587, message,
		WithAuth("user", "pass"),
		WithSMTPAuth(auth),
		WithTLSConfig(&tls.Config{ServerName: "smtp.example.com", MinVersion: tls.VersionTLS12}),
		WithTLSPolicy(TLSNone),
		WithAllowPlainAuth(true),
		WithTimeout(time.Second),
		WithLocalName("mail.local"),
		WithDialContext(func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("unused") }),
		WithSenderProvider(provider),
	); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got != message {
		t.Fatalf("sent message = %p, want %p", got, message)
	}

	got = nil
	if err := SendHTML(context.Background(), "smtp.example.com", 587, "from@example.com", []string{"to@example.com"}, "subject", "<p>html</p>", WithSenderProvider(func(Config) (Sender, error) {
		return SenderFunc(func(ctx context.Context, message *Message) error {
			got = message
			return nil
		}), nil
	})); err != nil {
		t.Fatalf("SendHTML() error = %v", err)
	}
	if got == nil || got.HTML != "<p>html</p>" {
		t.Fatalf("SendHTML message = %#v", got)
	}

	client, err := NewClient("smtp.example.com", 587, WithSenderProvider(provider))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}
}

func TestFacadeSendAccountText(t *testing.T) {
	account := Account{
		Host:           "smtp.example.com",
		Port:           587,
		Username:       "user@example.com",
		Password:       "secret",
		From:           "from@example.com",
		FromName:       "Facade Sender",
		TLSPolicy:      TLSNone,
		AllowPlainAuth: true,
		Timeout:        time.Second,
	}

	var got *Message
	provider := func(config Config) (Sender, error) {
		return SenderFunc(func(ctx context.Context, message *Message) error {
			got = message
			return nil
		}), nil
	}

	if err := SendAccountText(
		context.Background(),
		account,
		[]string{"to@example.com"},
		"subject",
		"body",
		WithQuickClientOptions(WithSenderProvider(provider)),
	); err != nil {
		t.Fatalf("SendAccountText() error = %v", err)
	}
	if got == nil || got.Subject != "subject" || got.Text != "body" {
		t.Fatalf("sent message = %#v", got)
	}
	if got.From.Name != "Facade Sender" || got.From.Email != "from@example.com" {
		t.Fatalf("from = %#v", got.From)
	}
}
