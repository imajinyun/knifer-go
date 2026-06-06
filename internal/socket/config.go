package socket

import (
	"runtime"
	"time"
)

// DefaultBufferSize is aligned with the utility toolkit IoUtil.DEFAULT_BUFFER_SIZE at 8 KB.
const DefaultBufferSize = 8 * 1024

// SocketConfig is aligned with the utility toolkit SocketConfig.
// It provides thread-pool size, timeout, buffer-size, and related socket options.
type SocketConfig struct {
	// ThreadPoolSize is the shared pool size and maps to the concurrency limit for accepting and handling connections in Go.
	ThreadPoolSize int

	// ReadTimeout is the read timeout in milliseconds; <= 0 means no timeout.
	ReadTimeout int64
	// WriteTimeout is the write timeout in milliseconds; <= 0 means no timeout.
	WriteTimeout int64

	// ReadBufferSize is the read buffer size.
	ReadBufferSize int
	// WriteBufferSize is the write buffer size.
	WriteBufferSize int

	// Clock returns the current time used to derive read/write deadlines. nil means time.Now.
	Clock func() time.Time
}

// ConfigOption customizes SocketConfig creation.
type ConfigOption func(*SocketConfig)

// WithThreadPoolSize sets the configured thread-pool/concurrency size.
func WithThreadPoolSize(n int) ConfigOption {
	return func(c *SocketConfig) { c.ThreadPoolSize = n }
}

// WithReadTimeout sets the read timeout in milliseconds.
func WithReadTimeout(ms int64) ConfigOption {
	return func(c *SocketConfig) { c.ReadTimeout = ms }
}

// WithWriteTimeout sets the write timeout in milliseconds.
func WithWriteTimeout(ms int64) ConfigOption {
	return func(c *SocketConfig) { c.WriteTimeout = ms }
}

// WithReadBufferSize sets the read buffer size.
func WithReadBufferSize(n int) ConfigOption {
	return func(c *SocketConfig) { c.ReadBufferSize = n }
}

// WithWriteBufferSize sets the write buffer size.
func WithWriteBufferSize(n int) ConfigOption {
	return func(c *SocketConfig) { c.WriteBufferSize = n }
}

// WithClock sets the clock used to derive read/write deadlines.
func WithClock(clock func() time.Time) ConfigOption {
	return func(c *SocketConfig) {
		if clock != nil {
			c.Clock = clock
		}
	}
}

// NewSocketConfig creates the default configuration.
func NewSocketConfig() *SocketConfig {
	return NewSocketConfigWithOptions()
}

// NewSocketConfigWithOptions creates a socket config customized by options.
func NewSocketConfigWithOptions(opts ...ConfigOption) *SocketConfig {
	cfg := &SocketConfig{
		ThreadPoolSize:  runtime.NumCPU(),
		ReadBufferSize:  DefaultBufferSize,
		WriteBufferSize: DefaultBufferSize,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	return cfg
}

func newConcurrencyLimiter(cfg *SocketConfig) chan struct{} {
	if cfg == nil || cfg.ThreadPoolSize <= 0 {
		return nil
	}
	return make(chan struct{}, cfg.ThreadPoolSize)
}

func acquireConcurrencySlot(limiter chan struct{}, done <-chan struct{}) bool {
	if limiter == nil {
		return true
	}
	select {
	case limiter <- struct{}{}:
		return true
	case <-done:
		return false
	}
}

func releaseConcurrencySlot(limiter chan struct{}) {
	if limiter != nil {
		<-limiter
	}
}

// SetThreadPoolSize sets the thread-pool size.
func (c *SocketConfig) SetThreadPoolSize(n int) *SocketConfig {
	c.ThreadPoolSize = n
	return c
}

// SetReadTimeout sets the read timeout in milliseconds.
func (c *SocketConfig) SetReadTimeout(ms int64) *SocketConfig {
	c.ReadTimeout = ms
	return c
}

// SetWriteTimeout sets the write timeout in milliseconds.
func (c *SocketConfig) SetWriteTimeout(ms int64) *SocketConfig {
	c.WriteTimeout = ms
	return c
}

// SetReadBufferSize sets the read buffer size.
func (c *SocketConfig) SetReadBufferSize(n int) *SocketConfig {
	c.ReadBufferSize = n
	return c
}

// SetWriteBufferSize sets the write buffer size.
func (c *SocketConfig) SetWriteBufferSize(n int) *SocketConfig {
	c.WriteBufferSize = n
	return c
}

// SetClock sets the clock used to derive read/write deadlines.
func (c *SocketConfig) SetClock(clock func() time.Time) *SocketConfig {
	if clock != nil {
		c.Clock = clock
	}
	return c
}
