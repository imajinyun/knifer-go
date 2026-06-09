package vcron

import (
	"context"
	"io"
	"time"

	"github.com/imajinyun/go-knifer/internal/cron"
)

// Config configures a scheduler.
type Config = cron.Config

// CronError is the cron module error type.
type CronError = cron.CronError

// Pattern is a parsed cron pattern.
type Pattern = cron.Pattern

// Scheduler schedules cron tasks.
type Scheduler = cron.Scheduler

// SchedulerOption customizes scheduler construction.
type SchedulerOption = cron.SchedulerOption

// DefaultSchedulerOption customizes one package-level scheduler operation.
type DefaultSchedulerOption = cron.DefaultSchedulerOption

// ConfigOption customizes cron config construction.
type ConfigOption = cron.ConfigOption

// CronTask is a scheduled task entry.
type CronTask = cron.CronTask

// Task is a cron task.
type Task = cron.Task

// TaskFunc adapts a function into Task.
type TaskFunc = cron.TaskFunc

// TaskListener listens to task execution events.
type TaskListener = cron.TaskListener

// Part identifies a cron expression part.
type Part = cron.Part

// PartMatcher matches a cron expression part.
type PartMatcher = cron.PartMatcher

// PatternOption customizes cron pattern parsing per call.
type PatternOption = cron.PatternOption

// SimpleTaskListener is a no-op task listener base.
type SimpleTaskListener = cron.SimpleTaskListener

// TaskExecutor executes a cron task.
type TaskExecutor = cron.TaskExecutor

// TaskTable stores scheduled tasks.
type TaskTable = cron.TaskTable

const (
	PartSecond     Part = cron.PartSecond
	PartMinute     Part = cron.PartMinute
	PartHour       Part = cron.PartHour
	PartDayOfMonth Part = cron.PartDayOfMonth
	PartMonth      Part = cron.PartMonth
	PartDayOfWeek  Part = cron.PartDayOfWeek
	PartYear       Part = cron.PartYear
)

var AlwaysTrueMatcher PartMatcher = cron.AlwaysTrueMatcher

// ErrSchedulerStarted is returned when immutable scheduler configuration is changed after Start.
var ErrSchedulerStarted = cron.ErrSchedulerStarted

// WithConfigLocation sets the scheduler time zone on CronConfig.
func WithConfigLocation(loc *time.Location) ConfigOption { return cron.WithConfigLocation(loc) }

// WithConfigMatchSecond sets whether cron expressions match seconds on CronConfig.
func WithConfigMatchSecond(matchSecond bool) ConfigOption {
	return cron.WithConfigMatchSecond(matchSecond)
}

// NewConfigWithOptions creates cron config customized by options.
func NewConfigWithOptions(opts ...ConfigOption) *Config {
	return cron.NewConfigWithOptions(opts...)
}

// WithPatternIntParser sets the integer parser used by NewPatternWithOptions.
func WithPatternIntParser(parser func(string) (int, error)) PatternOption {
	return cron.WithPatternIntParser(parser)
}

// NewScheduler creates a cron scheduler.
func NewScheduler() *Scheduler { return NewSchedulerWithOptions() }

// WithLocation sets the scheduler time zone.
func WithLocation(loc *time.Location) SchedulerOption { return cron.WithLocation(loc) }

// WithMatchSecond sets whether cron expressions match seconds.
func WithMatchSecond(matchSecond bool) SchedulerOption { return cron.WithMatchSecond(matchSecond) }

// WithExecutor sets the function used to execute scheduled tasks.
func WithExecutor(exec func(func())) SchedulerOption { return cron.WithExecutor(exec) }

// WithRunner sets the function used to launch the scheduler timer loop.
func WithRunner(runner func(func())) SchedulerOption { return cron.WithRunner(runner) }

// WithIDGenerator sets the task id generator used by Schedule and ScheduleFunc.
func WithIDGenerator(idFunc func() string) SchedulerOption { return cron.WithIDGenerator(idFunc) }

// WithIDRandomReader sets the random reader used by the default hexadecimal task id generator.
func WithIDRandomReader(reader io.Reader) SchedulerOption { return cron.WithIDRandomReader(reader) }

// WithClock sets the time source used by the scheduler timer.
func WithClock(clock func() time.Time) SchedulerOption { return cron.WithClock(clock) }

// WithSleeper sets the sleep function used by the scheduler timer.
func WithSleeper(sleeper func(time.Duration, <-chan struct{}) bool) SchedulerOption {
	return cron.WithSleeper(sleeper)
}

// WithSchedulerPatternOptions sets cron pattern parser providers used by scheduler string-pattern APIs.
func WithSchedulerPatternOptions(opts ...PatternOption) SchedulerOption {
	return cron.WithSchedulerPatternOptions(opts...)
}

// WithDefaultScheduler sets the scheduler used by one package-level operation.
func WithDefaultScheduler(s *Scheduler) DefaultSchedulerOption { return cron.WithDefaultScheduler(s) }

// WithDefaultSchedulerOptions creates an isolated scheduler for one package-level operation.
func WithDefaultSchedulerOptions(opts ...SchedulerOption) DefaultSchedulerOption {
	return cron.WithDefaultSchedulerOptions(opts...)
}

// NewSchedulerWithOptions creates a cron scheduler customized by options.
func NewSchedulerWithOptions(opts ...SchedulerOption) *Scheduler {
	return cron.NewSchedulerWithOptions(opts...)
}

// DefaultScheduler returns the package-level scheduler.
func DefaultScheduler() *Scheduler { return cron.DefaultScheduler() }

// DefaultSchedulerWithOptions returns the package-level scheduler or a per-call override.
func DefaultSchedulerWithOptions(opts ...DefaultSchedulerOption) *Scheduler {
	return cron.DefaultSchedulerWithOptions(opts...)
}

// ConfigureDefaultScheduler replaces the package-level scheduler with one created from options.
func ConfigureDefaultScheduler(opts ...SchedulerOption) *Scheduler {
	return cron.ConfigureDefaultScheduler(opts...)
}

// CronSchedule schedules a task on the default scheduler.
func CronSchedule(pattern string, task Task) (string, error) { return cron.Schedule(pattern, task) }

// CronScheduleWithOptions schedules a task on the selected default scheduler.
func CronScheduleWithOptions(pattern string, task Task, opts ...DefaultSchedulerOption) (string, error) {
	return cron.ScheduleWithOptions(pattern, task, opts...)
}

// CronScheduleFunc schedules fn on the default scheduler.
func CronScheduleFunc(pattern string, fn func()) (string, error) {
	return cron.ScheduleFunc(pattern, fn)
}

// CronScheduleFuncWithOptions schedules fn on the selected default scheduler.
func CronScheduleFuncWithOptions(pattern string, fn func(), opts ...DefaultSchedulerOption) (string, error) {
	return cron.ScheduleFuncWithOptions(pattern, fn, opts...)
}

// CronScheduleWithID schedules task with id.
func CronScheduleWithID(id, pattern string, task Task) error {
	return cron.ScheduleWithID(id, pattern, task)
}

// CronScheduleWithIDWithOptions schedules task with id on the selected default scheduler.
func CronScheduleWithIDWithOptions(id, pattern string, task Task, opts ...DefaultSchedulerOption) error {
	return cron.ScheduleWithIDWithOptions(id, pattern, task, opts...)
}

// CronRemove removes a task by id.
func CronRemove(id string) bool { return cron.Remove(id) }

// CronRemoveWithOptions removes a task by id from the selected default scheduler.
func CronRemoveWithOptions(id string, opts ...DefaultSchedulerOption) bool {
	return cron.RemoveWithOptions(id, opts...)
}

// CronUpdatePattern updates the pattern for a task.
func CronUpdatePattern(id, pattern string) error { return cron.UpdatePattern(id, pattern) }

// CronUpdatePatternWithOptions updates the pattern for a task on the selected default scheduler.
func CronUpdatePatternWithOptions(id, pattern string, opts ...DefaultSchedulerOption) error {
	return cron.UpdatePatternWithOptions(id, pattern, opts...)
}

// CronStart starts the default scheduler.
func CronStart() error { return cron.Start() }

// CronStartWithOptions starts the selected default scheduler.
func CronStartWithOptions(opts ...DefaultSchedulerOption) error {
	return cron.StartWithOptions(opts...)
}

// CronStop stops the default scheduler.
func CronStop() { cron.Stop() }

// CronStopWithOptions stops the selected default scheduler.
func CronStopWithOptions(opts ...DefaultSchedulerOption) { cron.StopWithOptions(opts...) }

// CronShutdown stops the default scheduler and waits for running tasks to finish.
func CronShutdown(ctx context.Context, clearTasks ...bool) error {
	return cron.Shutdown(ctx, clearTasks...)
}

// CronRunningCount returns the number of running task executions on the default scheduler.
func CronRunningCount() int { return cron.DefaultScheduler().RunningCount() }

// CronLaunchingCount returns the number of launcher jobs currently dispatching due tasks.
func CronLaunchingCount() int { return cron.DefaultScheduler().LaunchingCount() }

// CronShutdownWithOptions stops the selected default scheduler and waits for running tasks to finish.
func CronShutdownWithOptions(ctx context.Context, opts ...DefaultSchedulerOption) error {
	return cron.ShutdownWithOptions(ctx, opts...)
}

// CronRestart restarts the default scheduler.
func CronRestart() error { return cron.Restart() }

// CronRestartWithOptions restarts the selected default scheduler.
func CronRestartWithOptions(opts ...DefaultSchedulerOption) error {
	return cron.RestartWithOptions(opts...)
}

// CronSetMatchSecond sets whether expressions include seconds.
func CronSetMatchSecond(b bool) { cron.SetMatchSecond(b) }

// CronSetMatchSecondWithOptions sets whether expressions include seconds on the selected default scheduler.
func CronSetMatchSecondWithOptions(b bool, opts ...DefaultSchedulerOption) {
	cron.SetMatchSecondWithOptions(b, opts...)
}

// CronSetMatchSecondE sets whether expressions include seconds.
func CronSetMatchSecondE(b bool) error { return cron.SetMatchSecondE(b) }

// CronSetMatchSecondEWithOptions sets whether expressions include seconds on the selected default scheduler.
func CronSetMatchSecondEWithOptions(b bool, opts ...DefaultSchedulerOption) error {
	return cron.SetMatchSecondEWithOptions(b, opts...)
}
