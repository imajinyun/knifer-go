package errx

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

const (
	// SentryDSN is the environment variable used to override the configured DSN.
	SentryDSN = "SENTRY_DSN"
)

type initConfig struct {
	dsn             string
	envKey          string
	output          io.Writer
	formatter       logrus.Formatter
	reportCaller    bool
	levels          []logrus.Level
	getenv          func(string) string
	sentryOptions   sentry.ClientOptions
	sentryClient    *sentry.Client
	newSentryClient func(sentry.ClientOptions) (*sentry.Client, error)
	newSentryHook   func(*sentry.Client, []logrus.Level) (logrus.Hook, error)
	addHook         func(logrus.Hook)
	setReportCaller func(bool)
	setOutput       func(io.Writer)
	setFormatter    func(logrus.Formatter)
	logError        func(error, string)
}

// InitOption customizes logrus/Sentry initialization.
type InitOption func(*initConfig)

// WithSentryDSN sets the Sentry DSN.
func WithSentryDSN(dsn string) InitOption { return func(c *initConfig) { c.dsn = dsn } }

// WithSentryEnvKey sets the environment variable used to override the Sentry DSN.
func WithSentryEnvKey(key string) InitOption { return func(c *initConfig) { c.envKey = key } }

// WithLogOutput sets the logrus output writer.
func WithLogOutput(output io.Writer) InitOption {
	return func(c *initConfig) {
		if output != nil {
			c.output = output
		}
	}
}

// WithLogFormatter sets the logrus formatter.
func WithLogFormatter(formatter logrus.Formatter) InitOption {
	return func(c *initConfig) {
		if formatter != nil {
			c.formatter = formatter
		}
	}
}

// WithReportCaller controls whether logrus records caller information.
func WithReportCaller(reportCaller bool) InitOption {
	return func(c *initConfig) { c.reportCaller = reportCaller }
}

// WithSentryLevels sets the log levels forwarded to Sentry.
func WithSentryLevels(levels ...logrus.Level) InitOption {
	return func(c *initConfig) { c.levels = slices.Clone(levels) }
}

// WithEnvLookupFunc sets the environment lookup used to override the Sentry DSN.
func WithEnvLookupFunc(getenv func(string) string) InitOption {
	return func(c *initConfig) {
		if getenv != nil {
			c.getenv = getenv
		}
	}
}

// WithSentryClientOptions sets sentry-go client options used when creating the Sentry client.
func WithSentryClientOptions(options sentry.ClientOptions) InitOption {
	return func(c *initConfig) { c.sentryOptions = options }
}

// WithSentryClient sets the sentry-go client passed to the Sentry hook factory.
func WithSentryClient(client *sentry.Client) InitOption {
	return func(c *initConfig) {
		if client != nil {
			c.sentryClient = client
		}
	}
}

// WithSentryClientFactory sets the factory used to create sentry-go clients.
func WithSentryClientFactory(factory func(sentry.ClientOptions) (*sentry.Client, error)) InitOption {
	return func(c *initConfig) {
		if factory != nil {
			c.newSentryClient = factory
		}
	}
}

// WithSentryHookFactory sets the factory used to create the Sentry logrus hook.
func WithSentryHookFactory(factory func(*sentry.Client, []logrus.Level) (logrus.Hook, error)) InitOption {
	return func(c *initConfig) {
		if factory != nil {
			c.newSentryHook = factory
		}
	}
}

// WithLogHookAdder sets the function used to register the Sentry hook.
func WithLogHookAdder(addHook func(logrus.Hook)) InitOption {
	return func(c *initConfig) {
		if addHook != nil {
			c.addHook = addHook
		}
	}
}

// WithLogrusConfigurer sets the logrus global configuration functions used during initialization.
func WithLogrusConfigurer(setReportCaller func(bool), setOutput func(io.Writer), setFormatter func(logrus.Formatter)) InitOption {
	return func(c *initConfig) {
		if setReportCaller != nil {
			c.setReportCaller = setReportCaller
		}
		if setOutput != nil {
			c.setOutput = setOutput
		}
		if setFormatter != nil {
			c.setFormatter = setFormatter
		}
	}
}

// WithInitErrorLogger sets the logger used for initialization failures.
func WithInitErrorLogger(logError func(error, string)) InitOption {
	return func(c *initConfig) {
		if logError != nil {
			c.logError = logError
		}
	}
}

func defaultInitErrorLogger(err error, msg string) { logrus.WithError(err).Error(msg) }

var sentryLevelMap = map[logrus.Level]sentry.Level{
	logrus.TraceLevel: sentry.LevelDebug,
	logrus.DebugLevel: sentry.LevelDebug,
	logrus.InfoLevel:  sentry.LevelInfo,
	logrus.WarnLevel:  sentry.LevelWarning,
	logrus.ErrorLevel: sentry.LevelError,
	logrus.FatalLevel: sentry.LevelFatal,
	logrus.PanicLevel: sentry.LevelFatal,
}

type sentryLogrusHook struct {
	client *sentry.Client
	levels []logrus.Level
}

func newSentryLogrusHook(client *sentry.Client, levels []logrus.Level) (logrus.Hook, error) {
	if client == nil {
		return nil, errors.New("nil sentry client")
	}
	return &sentryLogrusHook{
		client: client,
		levels: slices.Clone(levels),
	}, nil
}

func (h *sentryLogrusHook) Levels() []logrus.Level { return slices.Clone(h.levels) }

func (h *sentryLogrusHook) Fire(entry *logrus.Entry) error {
	event := sentry.NewEvent()
	event.Level = sentryLevelMap[entry.Level]
	event.Message = entry.Message
	event.Timestamp = entry.Time
	event.Logger = "logrus"

	for key, value := range entry.Data {
		switch key {
		case logrus.ErrorKey:
			if err, ok := value.(error); ok {
				event.SetException(err, h.client.Options().MaxErrorDepth)
				continue
			}
		case "request":
			switch req := value.(type) {
			case *http.Request:
				event.Request = sentry.NewRequest(req)
				continue
			case sentry.Request:
				event.Request = &req
				continue
			case *sentry.Request:
				event.Request = req
				continue
			}
		case "user":
			switch user := value.(type) {
			case sentry.User:
				event.User = user
				continue
			case *sentry.User:
				event.User = *user
				continue
			}
		case "transaction":
			if transaction, ok := value.(string); ok {
				event.Transaction = transaction
				continue
			}
		case "fingerprint":
			if fingerprint, ok := value.([]string); ok {
				event.Fingerprint = fingerprint
				continue
			}
		}
		event.Tags[key] = fmt.Sprint(value)
	}

	var hint *sentry.EventHint
	if entry.Context != nil {
		hint = &sentry.EventHint{Context: entry.Context}
	}
	if h.client.CaptureEvent(event, hint, nil) == nil {
		return errors.New("failed to send to sentry")
	}
	return nil
}

func sentryClientOptionsWithDSN(cfg initConfig) sentry.ClientOptions {
	options := cfg.sentryOptions
	if options.Dsn == "" {
		options.Dsn = cfg.dsn
	}
	return options
}

func buildSentryClient(cfg initConfig) (*sentry.Client, error) {
	if cfg.sentryClient != nil {
		return cfg.sentryClient, nil
	}
	return cfg.newSentryClient(sentryClientOptionsWithDSN(cfg))
}

func applyInitOptions(opts []InitOption) initConfig {
	cfg := initConfig{
		envKey:       SentryDSN,
		output:       io.Discard,
		formatter:    EmptyFormatter,
		reportCaller: true,
		levels:       []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel},
		getenv:       os.Getenv,
		sentryOptions: sentry.ClientOptions{
			AttachStacktrace: true,
		},
		newSentryClient: sentry.NewClient,
		newSentryHook:   newSentryLogrusHook,
		addHook:         logrus.AddHook,
		setReportCaller: logrus.SetReportCaller,
		setOutput:       logrus.SetOutput,
		setFormatter:    logrus.SetFormatter,
		logError:        defaultInitErrorLogger,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.output == nil {
		cfg.output = io.Discard
	}
	if cfg.formatter == nil {
		cfg.formatter = EmptyFormatter
	}
	if len(cfg.levels) == 0 {
		cfg.levels = []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel}
	}
	if cfg.getenv == nil {
		cfg.getenv = os.Getenv
	}
	if cfg.newSentryClient == nil {
		cfg.newSentryClient = sentry.NewClient
	}
	if cfg.newSentryHook == nil {
		cfg.newSentryHook = newSentryLogrusHook
	}
	if cfg.addHook == nil {
		cfg.addHook = logrus.AddHook
	}
	if cfg.setReportCaller == nil {
		cfg.setReportCaller = logrus.SetReportCaller
	}
	if cfg.setOutput == nil {
		cfg.setOutput = logrus.SetOutput
	}
	if cfg.setFormatter == nil {
		cfg.setFormatter = logrus.SetFormatter
	}
	if cfg.logError == nil {
		cfg.logError = defaultInitErrorLogger
	}
	return cfg
}

// Init configures logrus to forward logs to the internal logs hook and,
// when a DSN is provided, to Sentry as well.
func Init(sentryDSN string) {
	InitWithOptions(WithSentryDSN(sentryDSN))
}

// InitWithOptions configures logrus output and optional Sentry forwarding with custom options.
func InitWithOptions(opts ...InitOption) {
	cfg := applyInitOptions(opts)
	cfg.setReportCaller(cfg.reportCaller)
	cfg.setOutput(cfg.output)
	cfg.setFormatter(cfg.formatter)

	if dsn := cfg.getenv(cfg.envKey); dsn != "" {
		cfg.dsn = dsn
	}
	if cfg.dsn == "" {
		return
	}
	client, err := buildSentryClient(cfg)
	if err != nil {
		cfg.logError(err, "sentry init failed")
		return
	}

	hook, err := cfg.newSentryHook(client, cfg.levels)
	if err != nil {
		cfg.logError(err, "sentry hook init failed")
		return
	}
	cfg.addHook(hook)
}

// NewIsolatedLogrusWithOptions creates and configures a standalone logrus logger.
// Unlike InitWithOptions, it does not mutate the package-level logrus logger or
// package-level Sentry state. When a DSN is configured, the returned logger receives
// a Sentry hook backed by an isolated sentry-go client unless WithSentryClient supplies one.
func NewIsolatedLogrusWithOptions(opts ...InitOption) *logrus.Logger {
	cfg := applyInitOptions(opts)
	logger := logrus.New()
	logger.SetReportCaller(cfg.reportCaller)
	logger.SetOutput(cfg.output)
	logger.SetFormatter(cfg.formatter)

	if dsn := cfg.getenv(cfg.envKey); dsn != "" {
		cfg.dsn = dsn
	}
	if cfg.dsn == "" {
		return logger
	}
	client, err := buildSentryClient(cfg)
	if err != nil {
		cfg.logError(err, "sentry init failed")
		return logger
	}

	hook, err := cfg.newSentryHook(client, cfg.levels)
	if err != nil {
		cfg.logError(err, "sentry hook init failed")
		return logger
	}
	logger.AddHook(hook)
	return logger
}
