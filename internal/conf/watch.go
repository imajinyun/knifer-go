package conf

import (
	"crypto/sha256"
	"os"
	"sync"
	"time"
)

// WatchTicker stops a watch polling ticker created by WatchTickerFactory.
type WatchTicker interface {
	Stop()
}

// WatchTickerFactory creates a ticker channel and stopper for WatchWithOptions.
type WatchTickerFactory func(time.Duration) (<-chan time.Time, WatchTicker)

// WatchEvent describes a watched file change.
type WatchEvent struct {
	Path    string
	ModTime time.Time
	Size    int64
	Hash    [32]byte
}

// WatchOptions controls polling watch behavior.
type WatchOptions struct {
	Interval       time.Duration
	Debounce       time.Duration
	CompareContent bool
	LoadOptions    LoadOptions
	OnEvent        func(WatchEvent)
	TickerFactory  WatchTickerFactory
	After          func(time.Duration) <-chan time.Time
	Runner         func(func())
	Stat           func(string) (os.FileInfo, error)
	ReadFile       func(path string, maxBytes int64) ([]byte, error)
}

// Watch polls path and calls onChange after successful reloads. The returned function stops watching.
func Watch(path string, interval time.Duration, onChange func(*Conf, error)) (func(), error) {
	return WatchWithOptions(path, WatchOptions{Interval: interval}, onChange)
}

// WatchWithOptions polls path using options and calls onChange after changes.
func WatchWithOptions(path string, opts WatchOptions, onChange func(*Conf, error)) (func(), error) {
	if onChange == nil {
		return nil, invalidInputf("watch callback is nil")
	}
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	opts = normalizeWatchOptions(opts)
	last, err := snapshot(path, opts)
	if err != nil {
		return nil, err
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	opts.Runner(func() {
		defer close(done)
		ticks, ticker := opts.TickerFactory(opts.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticks:
				current, statErr := snapshot(path, opts)
				if statErr != nil {
					safeWatchChange(onChange, nil, statErr)
					continue
				}
				if sameSnapshot(last, current, opts.CompareContent) {
					continue
				}
				last = current
				if opts.Debounce > 0 {
					select {
					case <-opts.After(opts.Debounce):
					case <-stop:
						return
					}
				}
				safeWatchEvent(opts.OnEvent, current)
				conf, loadErr := LoadWithOptions(path, watchLoadOptions(opts))
				safeWatchChange(onChange, conf, loadErr)
			case <-stop:
				return
			}
		}
	})
	var stopOnce sync.Once
	return func() {
		stopOnce.Do(func() {
			close(stop)
			<-done
		})
	}, nil
}

func safeWatchEvent(fn func(WatchEvent), event WatchEvent) {
	if fn == nil {
		return
	}
	defer func() { _ = recover() }()
	fn(event)
}

func safeWatchChange(fn func(*Conf, error), conf *Conf, err error) {
	if fn == nil {
		return
	}
	defer func() { _ = recover() }()
	fn(conf, err)
}

func snapshot(path string, opts WatchOptions) (WatchEvent, error) {
	info, err := opts.Stat(path)
	if err != nil {
		return WatchEvent{}, wrapConfigIO("stat config file "+path, err)
	}
	event := WatchEvent{Path: path, ModTime: info.ModTime(), Size: info.Size()}
	if opts.CompareContent {
		b, err := opts.ReadFile(path, opts.LoadOptions.MaxBytes) // #nosec G304 -- watcher intentionally reads the configured file path.
		if err != nil {
			return WatchEvent{}, wrapConfigIO("read config file "+path, err)
		}
		event.Hash = sha256.Sum256(b)
	}
	return event, nil
}

func normalizeWatchOptions(opts WatchOptions) WatchOptions {
	if opts.TickerFactory == nil {
		opts.TickerFactory = newWatchTicker
	}
	if opts.After == nil {
		opts.After = time.After
	}
	if opts.Runner == nil {
		opts.Runner = defaultWatchRunner
	}
	if opts.Stat == nil {
		opts.Stat = os.Stat
	}
	if opts.ReadFile == nil {
		loadOpts := opts.LoadOptions
		opts.ReadFile = func(path string, maxBytes int64) ([]byte, error) {
			loadOpts.MaxBytes = maxBytes
			return readFileWithOptions(path, loadOpts)
		}
	}
	return opts
}

func defaultWatchRunner(fn func()) { go fn() }

func newWatchTicker(interval time.Duration) (<-chan time.Time, WatchTicker) {
	ticker := time.NewTicker(interval)
	return ticker.C, ticker
}

func watchLoadOptions(opts WatchOptions) LoadOptions {
	loadOpts := opts.LoadOptions
	if loadOpts.ReadFile == nil {
		loadOpts.ReadFile = opts.ReadFile
	}
	return loadOpts
}

func sameSnapshot(a, b WatchEvent, compareContent bool) bool {
	if compareContent {
		return a.Hash == b.Hash
	}
	return a.ModTime.Equal(b.ModTime) && a.Size == b.Size
}
