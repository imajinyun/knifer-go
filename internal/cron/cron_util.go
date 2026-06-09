package cron

import (
	"context"
	"sync"
)

// defaultScheduler is the package-level scheduler aligned with the utility toolkit CronUtil.scheduler.
var (
	defaultMu        sync.Mutex
	defaultScheduler = NewScheduler()
)

// DefaultSchedulerOption customizes one package-level scheduler operation.
type DefaultSchedulerOption func(*defaultSchedulerConfig)

type defaultSchedulerConfig struct {
	scheduler *Scheduler
}

// WithDefaultScheduler sets the scheduler used by one package-level operation.
func WithDefaultScheduler(s *Scheduler) DefaultSchedulerOption {
	return func(cfg *defaultSchedulerConfig) {
		if s != nil {
			cfg.scheduler = s
		}
	}
}

// WithDefaultSchedulerOptions creates an isolated scheduler for one package-level operation.
func WithDefaultSchedulerOptions(opts ...SchedulerOption) DefaultSchedulerOption {
	return WithDefaultScheduler(NewSchedulerWithOptions(opts...))
}

func getDefaultScheduler() *Scheduler {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultScheduler == nil {
		defaultScheduler = NewScheduler()
	}
	return defaultScheduler
}

func applyDefaultSchedulerOptions(opts []DefaultSchedulerOption) *Scheduler {
	cfg := defaultSchedulerConfig{scheduler: getDefaultScheduler()}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.scheduler == nil {
		return getDefaultScheduler()
	}
	return cfg.scheduler
}

// DefaultScheduler returns the package-level scheduler.
func DefaultScheduler() *Scheduler {
	return DefaultSchedulerWithOptions()
}

// DefaultSchedulerWithOptions returns the package-level scheduler or a per-call override.
func DefaultSchedulerWithOptions(opts ...DefaultSchedulerOption) *Scheduler {
	return applyDefaultSchedulerOptions(opts)
}

// ConfigureDefaultScheduler replaces the package-level scheduler with one created from options.
func ConfigureDefaultScheduler(opts ...SchedulerOption) *Scheduler {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultScheduler != nil && defaultScheduler.IsStarted() {
		defaultScheduler.Stop()
	}
	defaultScheduler = NewSchedulerWithOptions(opts...)
	return defaultScheduler
}

// SetMatchSecond sets whether the package-level scheduler matches seconds.
func SetMatchSecond(b bool) {
	SetMatchSecondWithOptions(b)
}

// SetMatchSecondWithOptions sets whether the selected default scheduler matches seconds.
func SetMatchSecondWithOptions(b bool, opts ...DefaultSchedulerOption) {
	applyDefaultSchedulerOptions(opts).SetMatchSecond(b)
}

// SetMatchSecondE sets whether the package-level scheduler matches seconds.
// It returns ErrSchedulerStarted when the selected scheduler has already been started.
func SetMatchSecondE(b bool) error {
	return SetMatchSecondEWithOptions(b)
}

// SetMatchSecondEWithOptions sets whether the selected default scheduler matches seconds.
// It returns ErrSchedulerStarted when the selected scheduler has already been started.
func SetMatchSecondEWithOptions(b bool, opts ...DefaultSchedulerOption) error {
	return applyDefaultSchedulerOptions(opts).SetMatchSecondE(b)
}

// Schedule registers a task on the package-level scheduler and returns its id.
func Schedule(pattern string, task Task) (string, error) {
	return ScheduleWithOptions(pattern, task)
}

// ScheduleWithOptions registers a task on the selected default scheduler and returns its id.
func ScheduleWithOptions(pattern string, task Task, opts ...DefaultSchedulerOption) (string, error) {
	return applyDefaultSchedulerOptions(opts).Schedule(pattern, task)
}

// ScheduleFunc registers a function task on the package-level scheduler.
func ScheduleFunc(pattern string, fn func()) (string, error) {
	return ScheduleFuncWithOptions(pattern, fn)
}

// ScheduleFuncWithOptions registers a function task on the selected default scheduler.
func ScheduleFuncWithOptions(pattern string, fn func(), opts ...DefaultSchedulerOption) (string, error) {
	return applyDefaultSchedulerOptions(opts).ScheduleFunc(pattern, fn)
}

// ScheduleWithID registers a task with the specified id on the package-level scheduler.
func ScheduleWithID(id, pattern string, task Task) error {
	return ScheduleWithIDWithOptions(id, pattern, task)
}

// ScheduleWithIDWithOptions registers a task with the specified id on the selected default scheduler.
func ScheduleWithIDWithOptions(id, pattern string, task Task, opts ...DefaultSchedulerOption) error {
	return applyDefaultSchedulerOptions(opts).ScheduleWithID(id, pattern, task)
}

// Remove deletes a task from the package-level scheduler.
func Remove(id string) bool {
	return RemoveWithOptions(id)
}

// RemoveWithOptions deletes a task from the selected default scheduler.
func RemoveWithOptions(id string, opts ...DefaultSchedulerOption) bool {
	return applyDefaultSchedulerOptions(opts).Deschedule(id)
}

// UpdatePattern updates a task expression on the package-level scheduler.
func UpdatePattern(id, pattern string) error {
	return UpdatePatternWithOptions(id, pattern)
}

// UpdatePatternWithOptions updates a task expression on the selected default scheduler.
func UpdatePatternWithOptions(id, pattern string, opts ...DefaultSchedulerOption) error {
	return applyDefaultSchedulerOptions(opts).UpdatePattern(id, pattern)
}

// Start starts the package-level scheduler.
func Start() error {
	return StartWithOptions()
}

// StartWithOptions starts the selected default scheduler.
func StartWithOptions(opts ...DefaultSchedulerOption) error {
	if len(opts) > 0 {
		return applyDefaultSchedulerOptions(opts).Start()
	}
	defaultMu.Lock()
	defer defaultMu.Unlock()
	return defaultScheduler.Start()
}

// Stop stops the package-level scheduler.
func Stop() {
	StopWithOptions()
}

// StopWithOptions stops the selected default scheduler.
func StopWithOptions(opts ...DefaultSchedulerOption) {
	if len(opts) > 0 {
		applyDefaultSchedulerOptions(opts).Stop()
		return
	}
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultScheduler.Stop()
}

// Shutdown stops the package-level scheduler and waits for running tasks to finish.
func Shutdown(ctx context.Context, clearTasks ...bool) error {
	defaultMu.Lock()
	s := defaultScheduler
	defaultMu.Unlock()
	return s.Shutdown(ctx, clearTasks...)
}

// ShutdownWithOptions stops the selected default scheduler and waits for running tasks to finish.
func ShutdownWithOptions(ctx context.Context, opts ...DefaultSchedulerOption) error {
	return applyDefaultSchedulerOptions(opts).Shutdown(ctx)
}

// Restart restarts the package-level scheduler.
func Restart() error {
	return RestartWithOptions()
}

// RestartWithOptions restarts the selected default scheduler.
func RestartWithOptions(opts ...DefaultSchedulerOption) error {
	if len(opts) > 0 {
		s := applyDefaultSchedulerOptions(opts)
		s.Stop()
		return s.Start()
	}
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultScheduler.Stop()
	return defaultScheduler.Start()
}
