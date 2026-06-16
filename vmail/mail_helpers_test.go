package vmail

import (
	"context"
	"net/smtp"
	"strings"
	"testing"
)

func sequenceBoundary(values ...string) BoundaryGenerator {
	idx := 0
	return func() (string, error) {
		value := values[idx]
		idx++
		return value, nil
	}
}

type facadeSMTPAuth struct{ mechanism string }

func (a facadeSMTPAuth) Start(*smtp.ServerInfo) (string, []byte, error) { return a.mechanism, nil, nil }

func (a facadeSMTPAuth) Next([]byte, bool) ([]byte, error) { return nil, nil }

type facadeSendCloser struct {
	sent   bool
	closed bool
}

func (s *facadeSendCloser) Send(context.Context, *Message) error {
	s.sent = true
	return nil
}

func (s *facadeSendCloser) Close() error {
	s.closed = true
	return nil
}

func assertContains(t *testing.T, got, expected string) {
	t.Helper()
	if !strings.Contains(got, expected) {
		t.Fatalf("expected %q to contain %q", got, expected)
	}
}
