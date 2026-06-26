package verr_test

import (
	"errors"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"

	"github.com/imajinyun/knifer-go/verr"
)

func TestFacadeGetStack(t *testing.T) {
	if got := verr.GetStack(nil); got != "" {
		t.Fatalf("GetStack(nil) = %q, want empty", got)
	}
	err := errors.New("test")
	if got := verr.GetStack(err); got == "" {
		t.Fatal("GetStack(err) should return fallback stack")
	}
}

func TestFacadeInit(t *testing.T) {
	// Init delegates to InitWithOptions — just verify it doesn't panic.
	verr.Init("")
}

func TestFacadeErrOptionConstructors(t *testing.T) {
	_ = verr.WithSentryDSN("")
	_ = verr.WithSentryEnvKey("")
	_ = verr.WithLogFormatter(nil)
	_ = verr.WithSentryClientOptions(sentry.ClientOptions{})
	_ = verr.WithSentryClient(nil)
	_ = verr.WithInitErrorLogger(nil)
	_ = verr.WithCallersFunc(nil)
	_ = verr.WithFuncForPCFunc(nil)
	_ = verr.WithStackFrameCache(false)
	_ = verr.WithCollectorTimerFactory(nil)
}

func TestFacadeResetStackFrameCache(t *testing.T) {
	verr.ResetStackFrameCache()
}

func TestFacadeCollectorMethods(t *testing.T) {
	c := verr.NewCollector()
	_ = verr.WithLevel(c, logrus.InfoLevel)
	_ = verr.WithTimerFactory(c, nil)
	_ = verr.WithLogFunc(c, nil)
	_ = verr.WithCollectorStackOptions(c)
}

func TestFacadeCollectorTimerFactory(t *testing.T) {
	// WithCollectorTimerFactory uses the internal timer API.
	// Provide a real factory and verify no panic on use.
	collector := verr.NewCollectorWithOptions(verr.WithCollectorTimerFactory(func(d time.Duration) (<-chan time.Time, verr.Timer) {
		return nil, stubTimer{}
	}))
	_ = collector
}

type stubTimer struct{}

func (s stubTimer) Stop() bool { return true }
