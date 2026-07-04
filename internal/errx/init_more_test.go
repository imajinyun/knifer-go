package errx

import (
	"errors"
	"io"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

func TestWithSentryDSN(t *testing.T) {
	cfg := applyInitOptions([]InitOption{WithSentryDSN("https://public@example.com/1")})
	if cfg.dsn != "https://public@example.com/1" {
		t.Fatalf("dsn = %q, want %q", cfg.dsn, "https://public@example.com/1")
	}
}

func TestWithLogFormatter(t *testing.T) {
	f := &emptyFormatter{}
	cfg := applyInitOptions([]InitOption{WithLogFormatter(f)})
	if cfg.formatter != f {
		t.Fatal("WithLogFormatter did not set formatter")
	}
}

func TestWithSentryClientOptions(t *testing.T) {
	opts := sentry.ClientOptions{AttachStacktrace: false, Dsn: "https://public@example.com/1"}
	cfg := applyInitOptions([]InitOption{WithSentryClientOptions(opts)})
	if cfg.sentryOptions.Dsn != "https://public@example.com/1" {
		t.Fatalf("sentryOptions.Dsn = %q", cfg.sentryOptions.Dsn)
	}
	if cfg.sentryOptions.AttachStacktrace {
		t.Fatal("WithSentryClientOptions should override AttachStacktrace")
	}
}

func TestWithSentryClient(t *testing.T) {
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://public@example.com/1"})
	if err != nil {
		t.Fatal(err)
	}
	cfg := applyInitOptions([]InitOption{WithSentryClient(client)})
	if cfg.sentryClient != client {
		t.Fatal("WithSentryClient did not set client")
	}

	// buildSentryClient should return the pre-set client
	got, err := buildSentryClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got != client {
		t.Fatal("buildSentryClient should return pre-set client")
	}
}

func TestWithInitErrorLogger(t *testing.T) {
	var loggedErr error
	var loggedMsg string
	cfg := applyInitOptions([]InitOption{
		WithInitErrorLogger(func(err error, msg string) {
			loggedErr = err
			loggedMsg = msg
		}),
	})
	if cfg.logError == nil {
		t.Fatal("logError should not be nil")
	}
	cfg.logError(errors.New("test"), "test message")
	if loggedErr == nil || loggedMsg != "test message" {
		t.Fatalf("logError = %v, %q", loggedErr, loggedMsg)
	}

	// nil option should keep the default
	cfg2 := applyInitOptions([]InitOption{WithInitErrorLogger(nil)})
	if cfg2.logError == nil {
		t.Fatal("nil WithInitErrorLogger should not overwrite default")
	}
}

func TestDefaultInitErrorLogger(t *testing.T) {
	// Verify it doesn't panic
	defaultInitErrorLogger(errors.New("test error"), "test message")
}

func TestInitNoDSN(t *testing.T) {
	silenceLogrus(t)
	// Init with empty DSN should configure logrus and return without sentry
	Init("")
}

func TestNewSentryLogrusHook(t *testing.T) {
	// nil client returns error
	_, err := newSentryLogrusHook(nil, []logrus.Level{logrus.ErrorLevel})
	if err == nil {
		t.Fatal("expected error for nil client")
	}

	// valid client returns hook
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://public@example.com/1"})
	if err != nil {
		t.Fatal(err)
	}
	hook, err := newSentryLogrusHook(client, []logrus.Level{logrus.ErrorLevel})
	if err != nil {
		t.Fatal(err)
	}
	if hook == nil {
		t.Fatal("expected non-nil hook")
	}
}

func TestSentryLogrusHookLevels(t *testing.T) {
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://public@example.com/1"})
	if err != nil {
		t.Fatal(err)
	}
	hook := &sentryLogrusHook{
		client: client,
		levels: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel},
	}
	levels := hook.Levels()
	if len(levels) != 2 || levels[0] != logrus.ErrorLevel || levels[1] != logrus.FatalLevel {
		t.Fatalf("Levels = %v, want [error, fatal]", levels)
	}
}

func TestSentryLogrusHookFire(t *testing.T) {
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://public@example.com/1"})
	if err != nil {
		t.Fatal(err)
	}
	hook := &sentryLogrusHook{client: client, levels: []logrus.Level{logrus.ErrorLevel}}

	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetFormatter(EmptyFormatter)
	entry := logrus.NewEntry(logger)
	entry.Data[logrus.ErrorKey] = errors.New("test error")
	entry.Data["request"] = "not a request"
	entry.Data["user"] = sentry.User{ID: "user-1"}
	entry.Data["transaction"] = "test-transaction"
	entry.Data["fingerprint"] = []string{"fp1", "fp2"}
	entry.Data["custom_tag"] = "tag-value"
	entry.Message = "test message"
	entry.Level = logrus.ErrorLevel

	if err := hook.Fire(entry); err != nil {
		t.Fatalf("Fire returned error: %v", err)
	}
}

func TestApplyInitOptionsNilGuards(t *testing.T) {
	cfg := applyInitOptions([]InitOption{
		WithLogOutput(nil),
		WithLogFormatter(nil),
		WithSentryLevels(),
	})
	if cfg.output != io.Discard {
		t.Fatalf("nil output should default to io.Discard, got %T", cfg.output)
	}
	if cfg.formatter != EmptyFormatter {
		t.Fatalf("nil formatter should default to EmptyFormatter")
	}
	if len(cfg.levels) == 0 {
		t.Fatal("empty levels should have defaults")
	}
}

func TestNilInitOptionsDoNotClearPreviousProviders(t *testing.T) {
	var out discardRecorder
	formatter := &emptyFormatter{}
	cfg := applyInitOptions([]InitOption{
		WithLogOutput(&out),
		WithLogOutput(nil),
		WithLogFormatter(formatter),
		WithLogFormatter(nil),
	})
	if cfg.output != &out {
		t.Fatal("nil WithLogOutput cleared previous writer")
	}
	if cfg.formatter != formatter {
		t.Fatal("nil WithLogFormatter cleared previous formatter")
	}
}

type discardRecorder struct{}

func (discardRecorder) Write(p []byte) (int, error) { return len(p), nil }

func TestInitWithOptionsSentryClientFailure(t *testing.T) {
	silenceLogrus(t)

	var logged bool
	InitWithOptions(
		WithSentryDSN("https://public@example.com/1"),
		WithSentryClientFactory(func(sentry.ClientOptions) (*sentry.Client, error) {
			return nil, errors.New("client error")
		}),
		WithInitErrorLogger(func(error, string) { logged = true }),
	)
	if !logged {
		t.Fatal("expected sentry client init error to be logged")
	}
}

func TestInitWithOptionsSentryHookFailure(t *testing.T) {
	silenceLogrus(t)

	var logged bool
	InitWithOptions(
		WithSentryDSN("https://public@example.com/1"),
		WithSentryHookFactory(func(*sentry.Client, []logrus.Level) (logrus.Hook, error) {
			return nil, errors.New("hook error")
		}),
		WithInitErrorLogger(func(error, string) { logged = true }),
	)
	if !logged {
		t.Fatal("expected sentry hook init error to be logged")
	}
}

func TestInitWithOptionsEnvOverridesDSN(t *testing.T) {
	silenceLogrus(t)

	var logged bool
	InitWithOptions(
		WithSentryDSN(""),
		WithEnvLookupFunc(func(string) string { return "https://public@example.com/1" }),
		WithSentryClientFactory(func(sentry.ClientOptions) (*sentry.Client, error) {
			return nil, errors.New("client error")
		}),
		WithInitErrorLogger(func(error, string) { logged = true }),
	)
	if !logged {
		t.Fatal("expected sentry client init error to be logged")
	}
}

func TestNewIsolatedLogrusWithOptionsSentryClientFailure(t *testing.T) {
	logger := NewIsolatedLogrusWithOptions(
		WithSentryDSN("https://public@example.com/1"),
		WithEnvLookupFunc(func(string) string { return "" }),
		WithSentryClientFactory(func(sentry.ClientOptions) (*sentry.Client, error) {
			return nil, errors.New("client error")
		}),
		WithLogOutput(io.Discard),
	)
	if logger == nil {
		t.Fatal("expected non-nil logger even on sentry failure")
	}
}

func TestNewIsolatedLogrusWithOptionsEnvOverridesDSN(t *testing.T) {
	logger := NewIsolatedLogrusWithOptions(
		WithSentryDSN(""),
		WithEnvLookupFunc(func(string) string { return "https://public@example.com/1" }),
		WithSentryClientFactory(func(sentry.ClientOptions) (*sentry.Client, error) {
			return nil, errors.New("client error")
		}),
		WithLogOutput(io.Discard),
	)
	if logger == nil {
		t.Fatal("expected non-nil logger even on sentry failure")
	}
}
