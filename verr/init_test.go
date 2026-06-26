package verr_test

import (
	"io"
	"strings"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"

	"github.com/imajinyun/knifer-go/verr"
)

type facadeSentryHook struct {
	levels []logrus.Level
}

func (h *facadeSentryHook) Levels() []logrus.Level   { return h.levels }
func (h *facadeSentryHook) Fire(*logrus.Entry) error { return nil }

func TestInitWithOptionsFacade(t *testing.T) {
	var b strings.Builder
	verr.InitWithOptions(verr.WithLogOutput(&b), verr.WithReportCaller(false))
}

func TestInitWithOptionsFacadeSentryInjection(t *testing.T) {
	var (
		clientOptions sentry.ClientOptions
		hookClient    *sentry.Client
		hookLevels    []logrus.Level
		hookAdded     bool
	)

	verr.InitWithOptions(
		verr.WithEnvLookupFunc(func(key string) string {
			if key != verr.SentryDSN {
				t.Fatalf("env lookup key = %q, want %q", key, verr.SentryDSN)
			}
			return "https://public@example.com/1"
		}),
		verr.WithLogrusConfigurer(func(bool) {}, func(io.Writer) {}, func(logrus.Formatter) {}),
		verr.WithSentryClientFactory(func(options sentry.ClientOptions) (*sentry.Client, error) {
			clientOptions = options
			return sentry.NewClient(options)
		}),
		verr.WithSentryHookFactory(func(client *sentry.Client, levels []logrus.Level) (logrus.Hook, error) {
			hookClient = client
			hookLevels = append([]logrus.Level(nil), levels...)
			return &facadeSentryHook{levels: levels}, nil
		}),
		verr.WithSentryLevels(logrus.ErrorLevel),
		verr.WithLogHookAdder(func(logrus.Hook) { hookAdded = true }),
	)

	if clientOptions.Dsn != "https://public@example.com/1" || !clientOptions.AttachStacktrace {
		t.Fatalf("client options = %#v", clientOptions)
	}
	if hookClient == nil || len(hookLevels) != 1 || hookLevels[0] != logrus.ErrorLevel || !hookAdded {
		t.Fatalf("hook client=%v levels=%v added=%v", hookClient, hookLevels, hookAdded)
	}
}

func TestNewIsolatedLogrusWithOptionsFacade(t *testing.T) {
	var globalConfigured bool
	logger := verr.NewIsolatedLogrusWithOptions(
		verr.WithEnvLookupFunc(func(string) string { return "" }),
		verr.WithLogOutput(io.Discard),
		verr.WithReportCaller(false),
		verr.WithLogrusConfigurer(
			func(bool) { globalConfigured = true },
			func(io.Writer) { globalConfigured = true },
			func(logrus.Formatter) { globalConfigured = true },
		),
	)
	if logger == nil {
		t.Fatal("NewIsolatedLogrusWithOptions returned nil")
	}
	if globalConfigured {
		t.Fatal("isolated logger should not call global logrus configurers")
	}
	if logger.Out != io.Discard || logger.ReportCaller {
		t.Fatalf("isolated logger out=%T reportCaller=%v", logger.Out, logger.ReportCaller)
	}
}
