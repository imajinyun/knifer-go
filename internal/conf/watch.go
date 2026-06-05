package conf

import (
	"crypto/sha256"
	"os"
	"time"
)

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
}

// Watch polls path and calls onChange after successful reloads. The returned function stops watching.
func Watch(path string, interval time.Duration, onChange func(*Conf, error)) (func(), error) {
	return WatchWithOptions(path, WatchOptions{Interval: interval}, onChange)
}

// WatchWithOptions polls path using options and calls onChange after changes.
func WatchWithOptions(path string, opts WatchOptions, onChange func(*Conf, error)) (func(), error) {
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	last, err := snapshot(path, opts.CompareContent, opts.LoadOptions.MaxBytes)
	if err != nil {
		return nil, err
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(opts.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				current, statErr := snapshot(path, opts.CompareContent, opts.LoadOptions.MaxBytes)
				if statErr != nil {
					onChange(nil, statErr)
					continue
				}
				if sameSnapshot(last, current, opts.CompareContent) {
					continue
				}
				last = current
				if opts.Debounce > 0 {
					select {
					case <-time.After(opts.Debounce):
					case <-stop:
						return
					}
				}
				if opts.OnEvent != nil {
					opts.OnEvent(current)
				}
				onChange(LoadWithOptions(path, opts.LoadOptions))
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop); <-done }, nil
}

func snapshot(path string, compareContent bool, maxBytes int64) (WatchEvent, error) {
	info, err := os.Stat(path)
	if err != nil {
		return WatchEvent{}, wrapConfigIO("stat config file "+path, err)
	}
	event := WatchEvent{Path: path, ModTime: info.ModTime(), Size: info.Size()}
	if compareContent {
		b, err := readFileLimit(path, maxBytes) // #nosec G304 -- watcher intentionally reads the configured file path.
		if err != nil {
			return WatchEvent{}, wrapConfigIO("read config file "+path, err)
		}
		event.Hash = sha256.Sum256(b)
	}
	return event, nil
}

func sameSnapshot(a, b WatchEvent, compareContent bool) bool {
	if compareContent {
		return a.Hash == b.Hash
	}
	return a.ModTime.Equal(b.ModTime) && a.Size == b.Size
}
