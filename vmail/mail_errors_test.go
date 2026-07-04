package vmail

import (
	"context"
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFacadeExportsSentinelErrors(t *testing.T) {
	_, err := NewMessage(WithTo("to@example.com"), WithText("body"))
	if !errors.Is(err, ErrMissingFrom) {
		t.Fatalf("NewMessage() error = %v, want %v", err, ErrMissingFrom)
	}
}

func TestFacadeProviderErrorContract(t *testing.T) {
	msg, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	cause := errors.New("provider unavailable")
	err = Send(context.Background(), "smtp.example.com", 587, msg,
		WithAuth("user", "secret-password"),
		WithSenderProvider(func(Config) (Sender, error) {
			return nil, cause
		}),
	)
	if !errors.Is(err, cause) {
		t.Fatalf("Send() error = %v, want provider cause", err)
	}
	if !errors.Is(err, knifer.ErrCodeProviderFailure) {
		t.Fatalf("Send() error = %v, want ErrCodeProviderFailure", err)
	}
	if strings.Contains(err.Error(), "secret-password") {
		t.Fatalf("Send() error leaked password: %v", err)
	}
}
