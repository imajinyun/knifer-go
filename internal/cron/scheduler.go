package cron

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// Scheduler is aligned with the utility toolkit Scheduler and is the core scheduler of gkcron.
type Scheduler struct {
	mu          sync.Mutex
	configMu    sync.RWMutex
	runMu       sync.RWMutex
	config      *Config
	started     atomic.Bool
	timer       *cronTimer
	timerWG     sync.WaitGroup
	taskTable   *TaskTable
	launcherMgr *taskLauncherManager
	executorMgr *taskExecutorManager
	listenerMgr *listenerManager

	// executor controls goroutine usage for task execution and may be replaced with a concurrency-limited executor.
	executor    func(func())
	runner      func(func())
	idFunc      func() string
	nowFunc     func() time.Time
	sleeper     func(time.Duration, <-chan struct{}) bool
	patternOpts []PatternOption
}

// SchedulerOption customizes scheduler construction.
type SchedulerOption func(*Scheduler)

// WithLocation sets the scheduler time zone.
func WithLocation(loc *time.Location) SchedulerOption {
	return func(s *Scheduler) { s.SetTimeZone(loc) }
}

// WithMatchSecond sets whether cron expressions match seconds.
func WithMatchSecond(matchSecond bool) SchedulerOption {
	return func(s *Scheduler) { s.SetMatchSecond(matchSecond) }
}

// WithExecutor sets the function used to execute scheduled tasks.
func WithExecutor(exec func(func())) SchedulerOption {
	return func(s *Scheduler) { s.SetExecutor(exec) }
}

// WithRunner sets the function used to launch the scheduler timer loop.
func WithRunner(runner func(func())) SchedulerOption {
	return func(s *Scheduler) { s.SetRunner(runner) }
}

// WithIDGenerator sets the task id generator used by Schedule and ScheduleFunc.
func WithIDGenerator(idFunc func() string) SchedulerOption {
	return func(s *Scheduler) {
		if idFunc != nil {
			s.idFunc = idFunc
		}
	}
}

// WithIDRandomReader sets the random reader used by the default hexadecimal task id generator.
func WithIDRandomReader(reader io.Reader) SchedulerOption {
	return func(s *Scheduler) {
		if reader != nil {
			s.idFunc = newIDGeneratorWithReader(reader)
		}
	}
}

// WithClock sets the time source used by the scheduler timer.
func WithClock(clock func() time.Time) SchedulerOption {
	return func(s *Scheduler) {
		if clock != nil {
			s.nowFunc = clock
		}
	}
}

// WithSleeper sets the sleep function used by the scheduler timer.
func WithSleeper(sleeper func(time.Duration, <-chan struct{}) bool) SchedulerOption {
	return func(s *Scheduler) {
		if sleeper != nil {
			s.sleeper = sleeper
		}
	}
}

// WithSchedulerPatternOptions sets cron pattern parser providers used by scheduler string-pattern APIs.
func WithSchedulerPatternOptions(opts ...PatternOption) SchedulerOption {
	return func(s *Scheduler) {
		s.patternOpts = append([]PatternOption(nil), opts...)
	}
}

// NewScheduler creates a Scheduler.
func NewScheduler() *Scheduler {
	return NewSchedulerWithOptions()
}

// NewSchedulerWithOptions creates a Scheduler customized by options.
func NewSchedulerWithOptions(opts ...SchedulerOption) *Scheduler {
	s := &Scheduler{
		config:    NewConfig(),
		taskTable: NewTaskTable(),
	}
	s.launcherMgr = newTaskLauncherManager(s)
	s.executorMgr = newTaskExecutorManager(s)
	s.listenerMgr = newListenerManager()
	s.executor = func(fn func()) { go fn() }
	s.runner = func(fn func()) { go fn() }
	s.idFunc = newIDGeneratorWithReader(nil)
	s.nowFunc = time.Now
	s.sleeper = defaultTimerSleep
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

func (s *Scheduler) nowMillis() int64 {
	if s.nowFunc != nil {
		return s.nowFunc().UnixMilli()
	}
	return time.Now().UnixMilli()
}

func (s *Scheduler) sleep(d time.Duration, stopCh <-chan struct{}) bool {
	if s.sleeper != nil {
		return s.sleeper(d, stopCh)
	}
	return defaultTimerSleep(d, stopCh)
}

func (s *Scheduler) parsePattern(pattern string, opts ...PatternOption) (*Pattern, error) {
	allOpts := append([]PatternOption(nil), s.patternOpts...)
	allOpts = append(allOpts, opts...)
	return NewPatternWithOptions(pattern, allOpts...)
}

// Config returns a snapshot of the scheduler config.
func (s *Scheduler) Config() *Config {
	s.configMu.RLock()
	defer s.configMu.RUnlock()
	cfg := *s.config
	return &cfg
}

// SetMatchSecond sets whether expressions match seconds; calls while started are ignored.
func (s *Scheduler) SetMatchSecond(b bool) *Scheduler {
	if s.started.Load() {
		return s
	}
	s.configMu.Lock()
	defer s.configMu.Unlock()
	if s.started.Load() {
		return s
	}
	s.config.MatchSecond = b
	return s
}

// IsMatchSecond reports whether expressions match seconds.
func (s *Scheduler) IsMatchSecond() bool {
	s.configMu.RLock()
	defer s.configMu.RUnlock()
	return s.config.MatchSecond
}

// SetTimeZone sets the scheduler time zone; calls while started are ignored.
func (s *Scheduler) SetTimeZone(loc *time.Location) *Scheduler {
	if s.started.Load() {
		return s
	}
	s.configMu.Lock()
	defer s.configMu.Unlock()
	if s.started.Load() {
		return s
	}
	if loc == nil {
		loc = time.Local
	}
	s.config.Location = loc
	return s
}

// SetExecutor sets a custom execution function.
func (s *Scheduler) SetExecutor(exec func(func())) *Scheduler {
	if exec != nil {
		s.runMu.Lock()
		s.executor = exec
		s.runMu.Unlock()
	}
	return s
}

// SetRunner sets the function used to launch the scheduler timer loop.
func (s *Scheduler) SetRunner(runner func(func())) *Scheduler {
	if runner != nil {
		s.runMu.Lock()
		s.runner = runner
		s.runMu.Unlock()
	}
	return s
}

// AddListener adds a listener.
func (s *Scheduler) AddListener(l TaskListener) *Scheduler {
	s.listenerMgr.add(l)
	return s
}

// RemoveListener removes a listener.
func (s *Scheduler) RemoveListener(l TaskListener) *Scheduler {
	s.listenerMgr.remove(l)
	return s
}

// Schedule registers a task with an expression, generates an id automatically, and returns it.
func (s *Scheduler) Schedule(pattern string, task Task) (string, error) {
	return s.ScheduleWithPatternOptions(pattern, task)
}

// ScheduleWithPatternOptions registers a task with parser options, generates an id automatically, and returns it.
func (s *Scheduler) ScheduleWithPatternOptions(pattern string, task Task, opts ...PatternOption) (string, error) {
	id := s.idFunc()
	if err := s.ScheduleWithIDWithPatternOptions(id, pattern, task, opts...); err != nil {
		return "", err
	}
	return id, nil
}

// ScheduleFunc registers a function task.
func (s *Scheduler) ScheduleFunc(pattern string, fn func()) (string, error) {
	return s.ScheduleFuncWithPatternOptions(pattern, fn)
}

// ScheduleFuncWithPatternOptions registers a function task with parser options.
func (s *Scheduler) ScheduleFuncWithPatternOptions(pattern string, fn func(), opts ...PatternOption) (string, error) {
	return s.ScheduleWithPatternOptions(pattern, TaskFunc(fn), opts...)
}

// ScheduleWithID registers a task with the specified id.
func (s *Scheduler) ScheduleWithID(id, pattern string, task Task) error {
	return s.ScheduleWithIDWithPatternOptions(id, pattern, task)
}

// ScheduleWithIDWithPatternOptions registers a task with the specified id and parser options.
func (s *Scheduler) ScheduleWithIDWithPatternOptions(id, pattern string, task Task, opts ...PatternOption) error {
	p, err := s.parsePattern(pattern, opts...)
	if err != nil {
		return err
	}
	return s.SchedulePattern(id, p, task)
}

// SchedulePattern registers a task with an already parsed Pattern.
func (s *Scheduler) SchedulePattern(id string, p *Pattern, task Task) error {
	if p == nil {
		return NewCronError("pattern must not be nil")
	}
	return s.taskTable.Add(id, p, task)
}

// Deschedule deletes a task.
func (s *Scheduler) Deschedule(id string) bool {
	return s.taskTable.Remove(id)
}

// UpdatePattern updates a task expression.
func (s *Scheduler) UpdatePattern(id, pattern string) error {
	return s.UpdatePatternWithPatternOptions(id, pattern)
}

// UpdatePatternWithPatternOptions updates a task expression with parser options.
func (s *Scheduler) UpdatePatternWithPatternOptions(id, pattern string, opts ...PatternOption) error {
	p, err := s.parsePattern(pattern, opts...)
	if err != nil {
		return err
	}
	if !s.taskTable.UpdatePattern(id, p) {
		return NewCronError("task %q not found", id)
	}
	return nil
}

// TaskTable returns the task table.
func (s *Scheduler) TaskTable() *TaskTable { return s.taskTable }

// GetPattern returns a Pattern by id.
func (s *Scheduler) GetPattern(id string) *Pattern { return s.taskTable.GetPattern(id) }

// GetTask returns a Task by id.
func (s *Scheduler) GetTask(id string) Task { return s.taskTable.GetTask(id) }

// IsEmpty reports whether the task table is empty.
func (s *Scheduler) IsEmpty() bool { return s.taskTable.IsEmpty() }

// Size returns the task count.
func (s *Scheduler) Size() int { return s.taskTable.Size() }

// Clear removes all tasks.
func (s *Scheduler) Clear() {
	for _, id := range s.taskTable.IDs() {
		s.taskTable.Remove(id)
	}
}

// IsStarted reports whether the scheduler is started.
func (s *Scheduler) IsStarted() bool { return s.started.Load() }

// Start starts the scheduler.
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started.CompareAndSwap(false, true) {
		return NewCronError("scheduler already started")
	}
	s.timer = newCronTimer(s)
	s.timerWG.Add(1)
	s.run(s.timer.run)
	return nil
}

// Stop stops the scheduler and clears the task table when clearTasks is true.
func (s *Scheduler) Stop(clearTasks ...bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started.CompareAndSwap(true, false) {
		return
	}
	if s.timer != nil {
		s.timer.stopTimer()
		s.timer = nil
	}
	s.timerWG.Wait()
	if len(clearTasks) > 0 && clearTasks[0] {
		s.Clear()
	}
}

// RunningCount returns the number of task executions currently running.
func (s *Scheduler) RunningCount() int { return s.executorMgr.runningCount() }

// LaunchingCount returns the number of scheduler launcher jobs currently dispatching due tasks.
func (s *Scheduler) LaunchingCount() int { return s.launcherMgr.runningCount() }

// Wait blocks until all currently dispatching launchers and running task executions finish.
func (s *Scheduler) Wait() {
	s.launcherMgr.wait()
	s.executorMgr.wait()
}

// Shutdown stops the scheduler timer and waits for running task executions to finish
// or for ctx to be canceled. It does not forcibly cancel already running tasks.
func (s *Scheduler) Shutdown(ctx context.Context, clearTasks ...bool) error {
	s.Stop(clearTasks...)
	if err := s.launcherMgr.waitContext(ctx); err != nil {
		return err
	}
	return s.executorMgr.waitContext(ctx)
}

// submit executes fn asynchronously through the current executor.
func (s *Scheduler) submit(fn func()) {
	s.runMu.RLock()
	exec := s.executor
	s.runMu.RUnlock()
	exec(fn)
}

func (s *Scheduler) run(fn func()) {
	s.runMu.RLock()
	runner := s.runner
	s.runMu.RUnlock()
	if runner != nil {
		runner(fn)
		return
	}
	go fn()
}
