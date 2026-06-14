package errx

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
)

type collectorContextKey struct{}

func TestCollectorRecoverCollectsReturnedError(t *testing.T) {
	silenceLogrus(t)

	want := errors.New("boom")
	c := NewCollector().WithContext(context.TODO()).WithLevel(logrus.WarnLevel)
	got := c.Recover(func() error { return want }, "run %s", "job")
	if !ErrorIs(got, want) {
		t.Fatalf("Recover() error = %v, want wrapped %v", got, want)
	}
	if !ErrorIs(c.Error(), want) {
		t.Fatalf("Collector.Error() should include %v", want)
	}
}

func TestCollectorRecoverConvertsPanicToError(t *testing.T) {
	silenceLogrus(t)

	c := NewCollector()
	got := c.Recover(func() error {
		panic("panic-value")
	}, "panic job")
	if got == nil || !strings.Contains(got.Error(), "panic-value") {
		t.Fatalf("Recover() panic error = %v, want panic value", got)
	}
	if err := c.Error(); err == nil || !strings.Contains(err.Error(), "panic-value") {
		t.Fatalf("Collector.Error() = %v, want panic value", err)
	}
}

func TestCollectorCollectErrorAliasAndNilFunctions(t *testing.T) {
	silenceLogrus(t)

	var called atomic.Bool
	c := NewCollector()
	c.CollectError(func() error {
		called.Store(true)
		return nil
	}, "alias")
	if !called.Load() {
		t.Fatal("CollectError() did not run the function")
	}
	if err := c.Recover(nil, "nil function"); err != nil {
		t.Fatalf("Recover(nil) error = %v, want nil", err)
	}
	if err := c.Error(); err != nil {
		t.Fatalf("Collector.Error() = %v, want nil", err)
	}
}

func TestCollectorWithContextUsesProvidedContext(t *testing.T) {
	c := NewCollector()
	ctx := context.WithValue(context.Background(), collectorContextKey{}, "value")
	if got := c.WithContext(ctx); got != c {
		t.Fatal("WithContext should return the receiver")
	}
}
