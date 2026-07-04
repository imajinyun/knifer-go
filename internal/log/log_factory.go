package log

import "sync"

// LogFactory matches the utility toolkit LogFactory and provides Log lookup by name.
//
// SetFactory can replace the global implementation; the default returns ConsoleLog.
type LogFactory interface {
	// CreateLog creates a Log instance by name.
	CreateLog(name string) Log
}

// LogFactoryFunc adapts a function to LogFactory.
type LogFactoryFunc func(name string) Log

// CreateLog calls the underlying function.
func (f LogFactoryFunc) CreateLog(name string) Log { return f(name) }

// LoggerOption customizes logger lookup/creation for one call.
type LoggerOption func(*loggerConfig)

type loggerConfig struct {
	factory LogFactory
	cache   bool
}

// WithLoggerFactory sets the logger factory used by GetWithOptions or NewIsolatedLogger.
func WithLoggerFactory(factory LogFactory) LoggerOption {
	return func(cfg *loggerConfig) {
		if factory != nil {
			cfg.factory = factory
			cfg.cache = false
		}
	}
}

// WithLoggerConsoleOptions builds loggers with console options for one lookup/creation call.
func WithLoggerConsoleOptions(opts ...ConsoleLogOption) LoggerOption {
	return WithLoggerFactory(LogFactoryFunc(func(name string) Log {
		return NewConsoleLogWithOptions(name, opts...)
	}))
}

// WithLoggerCache controls whether GetWithOptions may use the package-level logger cache.
func WithLoggerCache(enabled bool) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.cache = enabled
	}
}

func defaultLogFactory() LogFactory {
	return LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) })
}

func applyLoggerOptions(base loggerConfig, opts ...LoggerOption) loggerConfig {
	if base.factory == nil {
		base.factory = defaultLogFactory()
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&base)
		}
	}
	if base.factory == nil {
		base.factory = defaultLogFactory()
	}
	return base
}

var (
	factoryMu      sync.RWMutex
	currentFactory LogFactory = defaultLogFactory()

	logCache   = make(map[string]Log)
	logCacheMu sync.RWMutex
)

// SetFactory sets the global log factory and clears cached Log instances.
func SetFactory(factory LogFactory) {
	if factory == nil {
		return
	}
	factoryMu.Lock()
	currentFactory = factory
	factoryMu.Unlock()

	logCacheMu.Lock()
	logCache = make(map[string]Log)
	logCacheMu.Unlock()
}

// GetFactory returns the current log factory.
func GetFactory() LogFactory {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	return currentFactory
}

// Get gets a cached Log instance by name.
func Get(name string) Log {
	logCacheMu.RLock()
	if l, ok := logCache[name]; ok {
		logCacheMu.RUnlock()
		return l
	}
	logCacheMu.RUnlock()

	factory := GetFactory()
	created := factory.CreateLog(name)

	logCacheMu.Lock()
	defer logCacheMu.Unlock()
	// Double-check.
	if l, ok := logCache[name]; ok {
		return l
	}
	logCache[name] = created
	return created
}

// GetWithOptions returns a logger by name with per-call factory/cache options.
func GetWithOptions(name string, opts ...LoggerOption) Log {
	if len(opts) == 0 {
		return Get(name)
	}
	cfg := applyLoggerOptions(loggerConfig{factory: GetFactory(), cache: true}, opts...)
	if cfg.cache {
		return Get(name)
	}
	return cfg.factory.CreateLog(name)
}

// NewIsolatedLogger creates a logger without reading package-level factory/cache state.
func NewIsolatedLogger(name string, opts ...LoggerOption) Log {
	cfg := applyLoggerOptions(loggerConfig{factory: defaultLogFactory(), cache: false}, opts...)
	return cfg.factory.CreateLog(name)
}

// GetDefault returns the Log instance named "default".
func GetDefault() Log {
	return Get("default")
}

// GetDefaultWithOptions returns the default logger with per-call factory/cache options.
func GetDefaultWithOptions(opts ...LoggerOption) Log {
	return GetWithOptions("default", opts...)
}
