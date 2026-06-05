package cron

import (
	"sync"
	"sync/atomic"
	"time"
)

// Scheduler is aligned with the utility toolkit Scheduler and is the core scheduler of gkcron.
type Scheduler struct {
	mu          sync.Mutex
	config      *Config
	started     atomic.Bool
	timer       *cronTimer
	taskTable   *TaskTable
	launcherMgr *taskLauncherManager
	executorMgr *taskExecutorManager
	listenerMgr *listenerManager

	// executor controls goroutine usage for task execution and may be replaced with a concurrency-limited executor.
	executor func(func())
	idFunc   func() string
	nowFunc  func() time.Time
	sleeper  func(time.Duration, <-chan struct{}) bool
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

// WithIDGenerator sets the task id generator used by Schedule and ScheduleFunc.
func WithIDGenerator(idFunc func() string) SchedulerOption {
	return func(s *Scheduler) {
		if idFunc != nil {
			s.idFunc = idFunc
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
	s.idFunc = generateID
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

// Config returns the scheduler config.
func (s *Scheduler) Config() *Config { return s.config }

// SetMatchSecond sets whether expressions match seconds; changes after start do not take effect.
func (s *Scheduler) SetMatchSecond(b bool) *Scheduler {
	s.config.MatchSecond = b
	return s
}

// IsMatchSecond reports whether expressions match seconds.
func (s *Scheduler) IsMatchSecond() bool { return s.config.MatchSecond }

// SetTimeZone sets the scheduler time zone.
func (s *Scheduler) SetTimeZone(loc *time.Location) *Scheduler {
	if loc == nil {
		loc = time.Local
	}
	s.config.Location = loc
	return s
}

// SetExecutor sets a custom execution function.
func (s *Scheduler) SetExecutor(exec func(func())) *Scheduler {
	if exec != nil {
		s.executor = exec
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
	id := s.idFunc()
	if err := s.ScheduleWithID(id, pattern, task); err != nil {
		return "", err
	}
	return id, nil
}

// ScheduleFunc registers a function task.
func (s *Scheduler) ScheduleFunc(pattern string, fn func()) (string, error) {
	return s.Schedule(pattern, TaskFunc(fn))
}

// ScheduleWithID registers a task with the specified id.
func (s *Scheduler) ScheduleWithID(id, pattern string, task Task) error {
	p, err := NewPattern(pattern)
	if err != nil {
		return err
	}
	return s.SchedulePattern(id, p, task)
}

// SchedulePattern registers a task with an already parsed Pattern.
func (s *Scheduler) SchedulePattern(id string, p *Pattern, task Task) error {
	return s.taskTable.Add(id, p, task)
}

// Deschedule deletes a task.
func (s *Scheduler) Deschedule(id string) bool {
	return s.taskTable.Remove(id)
}

// UpdatePattern updates a task expression.
func (s *Scheduler) UpdatePattern(id, pattern string) error {
	p, err := NewPattern(pattern)
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
	go s.timer.run()
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
	if len(clearTasks) > 0 && clearTasks[0] {
		s.Clear()
	}
}

// submit executes fn asynchronously through the current executor.
func (s *Scheduler) submit(fn func()) {
	s.executor(fn)
}
