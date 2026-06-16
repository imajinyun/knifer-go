package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSendTextAndHTMLConvenienceFunctions(t *testing.T) {
	for _, tt := range []struct {
		name     string
		send     func(context.Context, SenderProvider) error
		wantBody string
	}{
		{
			name: "text",
			send: func(ctx context.Context, provider SenderProvider) error {
				return SendText(ctx, "smtp.example.com", 587, "from@example.com", []string{"to@example.com"}, "subject", "plain", WithSenderProvider(provider))
			},
			wantBody: "plain",
		},
		{
			name: "html",
			send: func(ctx context.Context, provider SenderProvider) error {
				return SendHTML(ctx, "smtp.example.com", 587, "from@example.com", []string{"to@example.com"}, "subject", "<b>html</b>", WithSenderProvider(provider))
			},
			wantBody: "<b>html</b>",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got *Message
			provider := func(config Config) (Sender, error) {
				if config.Host != "smtp.example.com" || config.Port != 587 {
					t.Fatalf("Config = %#v", config)
				}
				return SenderFunc(func(ctx context.Context, message *Message) error {
					got = message
					return nil
				}), nil
			}
			if err := tt.send(context.Background(), provider); err != nil {
				t.Fatalf("send() error = %v", err)
			}
			if got == nil || !strings.Contains(got.Text+got.HTML, tt.wantBody) {
				t.Fatalf("sent message = %#v, want body %q", got, tt.wantBody)
			}
		})
	}
}

func TestAccountQuickSendUsesAccountDefaults(t *testing.T) {
	auth := testSMTPAuth{mechanism: "CUSTOM"}
	tlsConfig := &tls.Config{ServerName: "smtp.example.com", MinVersion: tls.VersionTLS12}
	account := Account{
		Host:           "smtp.example.com",
		Port:           587,
		Username:       "user@example.com",
		Password:       "secret",
		Auth:           auth,
		From:           "from@example.com",
		FromName:       "Sender",
		TLSConfig:      tlsConfig,
		TLSPolicy:      TLSNone,
		AllowPlainAuth: true,
		Timeout:        time.Second,
		LocalName:      "mail.local",
	}

	var got *Message
	provider := func(config Config) (Sender, error) {
		if config.Host != account.Host || config.Port != account.Port {
			t.Fatalf("Config address = %#v", config)
		}
		if config.Username != account.Username || config.Password != account.Password || config.Auth == nil {
			t.Fatalf("Config auth = %#v", config)
		}
		if config.TLSPolicy != TLSNone || !config.AllowPlainAuth || config.Timeout != time.Second || config.LocalName != "mail.local" {
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
	err := SendAccountText(
		context.Background(),
		account,
		[]string{"to@example.com"},
		"subject",
		"plain",
		WithQuickMessageOptions(WithHeader("X-Quick", "yes")),
		WithQuickClientOptions(WithSenderProvider(provider)),
	)
	if err != nil {
		t.Fatalf("SendAccountText() error = %v", err)
	}
	if got == nil || got.From.Email != "from@example.com" || got.From.Name != "Sender" {
		t.Fatalf("sent From = %#v", got)
	}
	if got.Subject != "subject" || got.Text != "plain" || got.Headers.Values("X-Quick")[0] != "yes" {
		t.Fatalf("sent message = %#v", got)
	}

	tlsConfig.ServerName = "mutated.example.com"
	if account.TLSConfig.ServerName != "mutated.example.com" {
		t.Fatalf("test setup did not mutate original TLSConfig")
	}
}

func TestQuickSendAndAccountValidation(t *testing.T) {
	provider := func(config Config) (Sender, error) {
		return SenderFunc(func(ctx context.Context, message *Message) error { return nil }), nil
	}
	account := Account{Host: "smtp.example.com", Port: 587, Username: "user@example.com"}
	if err := QuickSend(
		context.Background(),
		account,
		WithQuickMessageOptions(WithTo("to@example.com"), WithSubject("subject"), WithHTML("<p>html</p>")),
		WithQuickClientOptions(WithSenderProvider(provider), WithTLSPolicy(TLSNone)),
	); err != nil {
		t.Fatalf("QuickSend() error = %v", err)
	}

	err := QuickSend(
		context.Background(),
		Account{Host: "smtp.example.com", Port: 587},
		WithQuickMessageOptions(WithTo("to@example.com"), WithText("body")),
		WithQuickClientOptions(WithSenderProvider(provider)),
	)
	if !errors.Is(err, ErrMissingFrom) {
		t.Fatalf("QuickSend() error = %v, want %v", err, ErrMissingFrom)
	}

	quickErr := errors.New("quick option failed")
	err = QuickSend(context.Background(), account, func(*quickConfig) error { return quickErr })
	if !errors.Is(err, quickErr) {
		t.Fatalf("QuickSend(option error) = %v, want %v", err, quickErr)
	}
}
