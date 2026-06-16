package vmail

import (
	"context"
	"testing"
)

func TestFacadeClientDialUsesSendCloserProvider(t *testing.T) {
	message, err := NewMessage(WithFrom("from@example.com"), WithTo("to@example.com"), WithText("body"))
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	sendCloser := &facadeSendCloser{}
	client, err := NewClient("smtp.example.com", 587, WithSenderProvider(func(Config) (Sender, error) {
		return sendCloser, nil
	}))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	dialed, err := client.Dial(context.Background())
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	if dialed != sendCloser {
		t.Fatalf("Dial() = %p, want %p", dialed, sendCloser)
	}
	if err := dialed.Send(context.Background(), message); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if !sendCloser.sent {
		t.Fatal("SendCloser did not record Send")
	}
	if err := dialed.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !sendCloser.closed {
		t.Fatal("SendCloser did not record Close")
	}
}
