package vcron

import (
	"time"

	"github.com/imajinyun/go-knifer/internal/cron"
)

// CronConfig configures a scheduler.
type CronConfig = cron.Config

// Config configures a scheduler.
type Config = cron.Config

// CronError is the cron module error type.
type CronError = cron.CronError

// CronPattern is a parsed cron pattern.
type CronPattern = cron.Pattern

// Pattern is a parsed cron pattern.
type Pattern = cron.Pattern

// Scheduler schedules cron tasks.
type Scheduler = cron.Scheduler

// SchedulerOption customizes scheduler construction.
type SchedulerOption = cron.SchedulerOption

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

// SimpleTaskListener is a no-op task listener base.
type SimpleTaskListener = cron.SimpleTaskListener

// TaskExecutor executes a cron task.
type TaskExecutor = cron.TaskExecutor

// TaskTable stores scheduled tasks.
type TaskTable = cron.TaskTable

// NewCronConfig creates default cron config.
func NewCronConfig() *CronConfig { return cron.NewConfig() }

// NewCronPattern parses a cron expression.
func NewCronPattern(expr string) (*CronPattern, error) { return cron.NewPattern(expr) }

// MustNewCronPattern parses a cron expression or panics.
func MustNewCronPattern(expr string) *CronPattern { return cron.MustNewPattern(expr) }

// NewScheduler creates a cron scheduler.
func NewScheduler() *Scheduler { return cron.NewScheduler() }

// WithLocation sets the scheduler time zone.
func WithLocation(loc *time.Location) SchedulerOption { return cron.WithLocation(loc) }

// WithMatchSecond sets whether cron expressions match seconds.
func WithMatchSecond(matchSecond bool) SchedulerOption { return cron.WithMatchSecond(matchSecond) }

// WithExecutor sets the function used to execute scheduled tasks.
func WithExecutor(exec func(func())) SchedulerOption { return cron.WithExecutor(exec) }

// WithIDGenerator sets the task id generator used by Schedule and ScheduleFunc.
func WithIDGenerator(idFunc func() string) SchedulerOption { return cron.WithIDGenerator(idFunc) }

// WithClock sets the time source used by the scheduler timer.
func WithClock(clock func() time.Time) SchedulerOption { return cron.WithClock(clock) }

// WithSleeper sets the sleep function used by the scheduler timer.
func WithSleeper(sleeper func(time.Duration, <-chan struct{}) bool) SchedulerOption {
	return cron.WithSleeper(sleeper)
}

// NewSchedulerWithOptions creates a cron scheduler customized by options.
func NewSchedulerWithOptions(opts ...SchedulerOption) *Scheduler {
	return cron.NewSchedulerWithOptions(opts...)
}

// DefaultScheduler returns the package-level scheduler.
func DefaultScheduler() *Scheduler { return cron.DefaultScheduler() }

// ConfigureDefaultScheduler replaces the package-level scheduler with one created from options.
func ConfigureDefaultScheduler(opts ...SchedulerOption) *Scheduler {
	return cron.ConfigureDefaultScheduler(opts...)
}

// CronSchedule schedules a task on the default scheduler.
func CronSchedule(pattern string, task Task) (string, error) { return cron.Schedule(pattern, task) }

// CronScheduleFunc schedules fn on the default scheduler.
func CronScheduleFunc(pattern string, fn func()) (string, error) {
	return cron.ScheduleFunc(pattern, fn)
}

// CronScheduleWithID schedules task with id.
func CronScheduleWithID(id, pattern string, task Task) error {
	return cron.ScheduleWithID(id, pattern, task)
}

// CronRemove removes a task by id.
func CronRemove(id string) bool { return cron.Remove(id) }

// CronUpdatePattern updates the pattern for a task.
func CronUpdatePattern(id, pattern string) error { return cron.UpdatePattern(id, pattern) }

// CronStart starts the default scheduler.
func CronStart() error { return cron.Start() }

// CronStop stops the default scheduler.
func CronStop() { cron.Stop() }

// CronRestart restarts the default scheduler.
func CronRestart() error { return cron.Restart() }

// CronSetMatchSecond sets whether expressions include seconds.
func CronSetMatchSecond(b bool) { cron.SetMatchSecond(b) }
