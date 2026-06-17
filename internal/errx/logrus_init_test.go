package errx

import (
	"io"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

type testSentryHook struct {
	levels []logrus.Level
}

func (h *testSentryHook) Levels() []logrus.Level   { return h.levels }
func (h *testSentryHook) Fire(*logrus.Entry) error { return nil }

func TestInitWithOptionsUsesInjectedProviders(t *testing.T) {
	var reportCaller bool
	var output io.Writer
	var formatter logrus.Formatter
	var getenvKey string
	var clientOptions sentry.ClientOptions
	var hookClient *sentry.Client
	var hookLevels []logrus.Level
	var hookAdded bool

	InitWithOptions(
		WithSentryEnvKey("CUSTOM_DSN"),
		WithEnvLookupFunc(func(key string) string {
			getenvKey = key
			return "https://public@example.com/1"
		}),
		WithLogrusConfigurer(
			func(v bool) { reportCaller = v },
			func(w io.Writer) { output = w },
			func(f logrus.Formatter) { formatter = f },
		),
		WithSentryClientFactory(func(options sentry.ClientOptions) (*sentry.Client, error) {
			clientOptions = options
			return sentry.NewClient(options)
		}),
		WithSentryHookFactory(func(client *sentry.Client, levels []logrus.Level) (logrus.Hook, error) {
			hookClient = client
			hookLevels = append([]logrus.Level(nil), levels...)
			return &testSentryHook{levels: levels}, nil
		}),
		WithLogHookAdder(func(logrus.Hook) { hookAdded = true }),
		WithSentryLevels(logrus.ErrorLevel),
	)

	if !reportCaller || output != io.Discard || formatter != EmptyFormatter {
		t.Fatalf("logrus config = reportCaller %v output %T formatter %T", reportCaller, output, formatter)
	}
	if getenvKey != "CUSTOM_DSN" {
		t.Fatalf("dsn provider key=%q", getenvKey)
	}
	if clientOptions.Dsn != "https://public@example.com/1" || !clientOptions.AttachStacktrace {
		t.Fatalf("client options = %#v", clientOptions)
	}
	if hookClient == nil || len(hookLevels) != 1 || hookLevels[0] != logrus.ErrorLevel || !hookAdded {
		t.Fatalf("hook providers client=%v levels=%v added=%v", hookClient, hookLevels, hookAdded)
	}
}

func TestNewIsolatedLogrusWithOptionsDoesNotUseGlobalConfigurers(t *testing.T) {
	var globalConfigured bool
	logger := NewIsolatedLogrusWithOptions(
		WithEnvLookupFunc(func(string) string { return "" }),
		WithLogOutput(io.Discard),
		WithReportCaller(false),
		WithLogrusConfigurer(
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
		t.Fatalf("isolated logger config out=%T reportCaller=%v", logger.Out, logger.ReportCaller)
	}
}
