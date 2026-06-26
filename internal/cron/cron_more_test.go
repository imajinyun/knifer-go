package cron

import (
	"context"
	"errors"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestCronErrorMessage(t *testing.T) {
	// without cause
	err := NewCronError("test error %d", 1)
	if got := err.Error(); got != "test error 1" {
		t.Fatalf("Error() = %q, want %q", got, "test error 1")
	}

	// with cause
	err2 := WrapCronError(errors.New("root cause"), "wrapped")
	if got := err2.Error(); got != "wrapped: root cause" {
		t.Fatalf("Error() with cause = %q", got)
	}
}

func TestCronErrorCode(t *testing.T) {
	err := NewCronError("invalid")
	if got := err.ErrorCode(); got != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrorCode = %v, want ErrCodeInvalidInput", got)
	}

	err2 := newSchedulerStartedError()
	if got := err2.ErrorCode(); got != knifer.ErrCodeUnsupported {
		t.Fatalf("ErrorCode started = %v, want ErrCodeUnsupported", got)
	}
}

func TestSimpleTaskListenerNoOps(t *testing.T) {
	l := SimpleTaskListener{}
	// These should not panic
	l.OnStart(nil)
	l.OnSucceeded(nil)
	l.OnFailed(nil, nil)
}

func TestCronTaskGettersAndSetters(t *testing.T) {
	p, err := NewPattern("* * * * *")
	if err != nil {
		t.Fatal(err)
	}
	task := TaskFunc(func() {})
	ct := NewCronTask("test-id", p, task)

	if got := ct.ID(); got != "test-id" {
		t.Fatalf("ID = %q, want %q", got, "test-id")
	}
	if got := ct.Pattern(); got != p {
		t.Fatal("Pattern mismatch")
	}
	if got := ct.Raw(); got == nil {
		t.Fatal("Raw task should not be nil")
	}

	p2, err := NewPattern("0 0 * * *")
	if err != nil {
		t.Fatal(err)
	}
	ct.SetPattern(p2)
	if got := ct.Pattern(); got != p2 {
		t.Fatal("SetPattern did not update")
	}
}

func TestCronTaskExecuteNilRaw(t *testing.T) {
	ct := &CronTask{}
	ct.Execute() // should not panic
}

func TestTaskExecutorGetters(t *testing.T) {
	s := NewScheduler()
	task := TaskFunc(func() {})
	ct := NewCronTask("exec-id", MustNewPattern("* * * * *"), task)

	e := &TaskExecutor{scheduler: s, task: ct}
	if got := e.CronTask(); got != ct {
		t.Fatal("CronTask getter mismatch")
	}
	if got := e.Task(); got == nil {
		t.Fatal("Task getter should not return nil")
	}
}

func TestBoolArrayMatcherMinMax(t *testing.T) {
	m := newBoolArrayMatcher([]int{2, 5, 8})
	if got := m.MinValue(); got != 2 {
		t.Fatalf("MinValue = %d, want 2", got)
	}
	if got := m.MaxValue(); got != 8 {
		t.Fatalf("MaxValue = %d, want 8", got)
	}

	// Match
	if !m.Match(5) {
		t.Fatal("Match(5) should be true")
	}
	if m.Match(3) {
		t.Fatal("Match(3) should be false")
	}

	// NextAfter
	if got := m.NextAfter(6); got != 8 {
		t.Fatalf("NextAfter(6) = %d, want 8", got)
	}
	if got := m.NextAfter(9); got != 2 {
		t.Fatalf("NextAfter(9) should wrap to %d, got %d", 2, got)
	}
}

func TestBoolArrayMatcherEmpty(t *testing.T) {
	m := newBoolArrayMatcher(nil)
	if m.Match(0) {
		t.Fatal("empty matcher should not match")
	}
	if got := m.MinValue(); got != 0 {
		t.Fatalf("empty MinValue = %d, want 0", got)
	}
}

func TestAlwaysTrueMatcher(t *testing.T) {
	if !AlwaysTrueMatcher.Match(0) {
		t.Fatal("AlwaysTrueMatcher should match 0")
	}
	if !AlwaysTrueMatcher.Match(100) {
		t.Fatal("AlwaysTrueMatcher should match 100")
	}
	if got := AlwaysTrueMatcher.NextAfter(42); got != 42 {
		t.Fatalf("NextAfter(42) = %d, want 42", got)
	}
}

func TestDefaultScheduler(t *testing.T) {
	s := DefaultScheduler()
	if s == nil {
		t.Fatal("DefaultScheduler returned nil")
	}
}

func TestDefaultSchedulerWithOptions(t *testing.T) {
	s := DefaultSchedulerWithOptions()
	if s == nil {
		t.Fatal("DefaultSchedulerWithOptions returned nil")
	}

	// With per-call option
	s2 := DefaultSchedulerWithOptions(WithDefaultSchedulerOptions())
	if s2 == nil {
		t.Fatal("isolated DefaultSchedulerWithOptions returned nil")
	}
}

func TestSetMatchSecondOnNewScheduler(t *testing.T) {
	s := NewScheduler()
	s.SetMatchSecond(true)
	if !s.IsMatchSecond() {
		t.Fatal("SetMatchSecond(true) failed")
	}
	SetMatchSecondWithOptions(false, WithDefaultScheduler(s))
	if s.IsMatchSecond() {
		t.Fatal("SetMatchSecondWithOptions(false) failed")
	}
}

func TestScheduleAndRunOnNewScheduler(t *testing.T) {
	s := NewScheduler()
	s.SetMatchSecond(true)
	done := make(chan struct{})
	id, err := ScheduleWithOptions("* * * * * *", TaskFunc(func() { close(done) }), WithDefaultScheduler(s))
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("ScheduleWithOptions returned empty ID")
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for task execution")
	}
}

func TestRemoveAndUpdatePattern(t *testing.T) {
	s := NewScheduler()
	if err := s.SchedulePattern("t1", MustNewPattern("* * * * *"), TaskFunc(func() {})); err != nil {
		t.Fatal(err)
	}
	if s.Size() != 1 {
		t.Fatal("expected 1 task")
	}
	if !RemoveWithOptions("t1", WithDefaultScheduler(s)) {
		t.Fatal("RemoveWithOptions returned false")
	}
	if s.Size() != 0 {
		t.Fatal("expected 0 tasks after remove")
	}

	if err := s.SchedulePattern("t2", MustNewPattern("0 0 * * *"), TaskFunc(func() {})); err != nil {
		t.Fatal(err)
	}
	if err := UpdatePatternWithOptions("t2", "0 12 * * *", WithDefaultScheduler(s)); err != nil {
		t.Fatal(err)
	}
}

func TestApplyDefaultSchedulerOptionsNilOption(t *testing.T) {
	cfg := defaultSchedulerConfig{scheduler: NewScheduler()}
	WithDefaultScheduler(nil)(&cfg)
	if cfg.scheduler == nil {
		t.Fatal("nil WithDefaultScheduler should not replace scheduler")
	}
}

func TestShutdownWithOptions(t *testing.T) {
	s := NewScheduler()
	if err := ShutdownWithOptions(context.Background(), WithDefaultScheduler(s)); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultSchedulerStop(t *testing.T) {
	s := DefaultScheduler()
	StopWithOptions(WithDefaultScheduler(s)) // should not panic
}

func TestDefaultSchedulerStartError(t *testing.T) {
	s := NewScheduler()
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	// Starting an already started scheduler should return ErrSchedulerStarted
	StartWithOptions(WithDefaultScheduler(s))
}

func TestLastDayOfMonth(t *testing.T) {
	tests := []struct {
		month int
		leap  bool
		want  int
	}{
		{1, false, 31},
		{2, false, 28},
		{2, true, 29},
		{4, false, 30},
		{12, false, 31},
	}
	for _, tc := range tests {
		if got := lastDayOfMonth(tc.month, tc.leap); got != tc.want {
			t.Fatalf("lastDayOfMonth(%d, %v) = %d, want %d", tc.month, tc.leap, got, tc.want)
		}
	}
}

func TestIsLeapYear(t *testing.T) {
	if !isLeapYear(2000) {
		t.Fatal("2000 should be leap")
	}
	if isLeapYear(1900) {
		t.Fatal("1900 should not be leap")
	}
	if !isLeapYear(2024) {
		t.Fatal("2024 should be leap")
	}
	if isLeapYear(2023) {
		t.Fatal("2023 should not be leap")
	}
}

func TestSchedulerRemoveListener(t *testing.T) {
	s := NewScheduler()
	l := SimpleTaskListener{}
	s.AddListener(l)
	s.RemoveListener(l) // should not panic
}

func TestSchedulerTaskTableGetTaskIsEmptyClear(t *testing.T) {
	s := NewScheduler()
	if !s.IsEmpty() {
		t.Fatal("new scheduler should be empty")
	}
	tt := s.TaskTable()
	if tt == nil {
		t.Fatal("TaskTable should not be nil")
	}
	if got := s.GetTask("nonexistent"); got != nil {
		t.Fatal("GetTask nonexistent should return nil")
	}
	s.SchedulePattern("t1", MustNewPattern("* * * * *"), TaskFunc(func() {}))
	if s.IsEmpty() {
		t.Fatal("scheduler should not be empty after schedule")
	}
	if s.Size() != 1 {
		t.Fatal("size should be 1")
	}
	s.Clear()
	if !s.IsEmpty() {
		t.Fatal("scheduler should be empty after Clear")
	}
}

func TestMustNewPatternWithOptions(t *testing.T) {
	p := MustNewPatternWithOptions("* * * * *")
	if p == nil {
		t.Fatal("MustNewPatternWithOptions returned nil")
	}
	if p.Raw() != "* * * * *" {
		t.Fatalf("Raw = %q", p.Raw())
	}
}

func TestYearValueMatcherNextAfter(t *testing.T) {
	m := newYearValueMatcher([]int{2020, 2024, 2025, 2030})
	if !m.Match(2024) {
		t.Fatal("Match 2024 should be true")
	}
	if m.Match(2023) {
		t.Fatal("Match 2023 should be false")
	}
	if got := m.NextAfter(2025); got != 2025 {
		t.Fatalf("NextAfter(2025) = %d, want 2025", got)
	}
	if got := m.NextAfter(2026); got != 2030 {
		t.Fatalf("NextAfter(2026) = %d, want 2030", got)
	}
	if got := m.NextAfter(2031); got != -1 {
		t.Fatalf("NextAfter(2031) = %d, want -1", got)
	}
}

func TestListenerManagerRemove(t *testing.T) {
	m := newListenerManager()
	l := SimpleTaskListener{}
	m.add(l)
	m.add(SimpleTaskListener{})
	m.remove(l) // should remove first instance only
	// Should not panic removing non-existent
	m.remove(SimpleTaskListener{})
}

func TestScheduleWrappersWithNewScheduler(t *testing.T) {
	s := NewScheduler()
	// Schedule (no WithOptions)
	_, err := ScheduleWithOptions("* * * * *", TaskFunc(func() {}), WithDefaultScheduler(s))
	if err != nil {
		t.Fatal(err)
	}
	// ScheduleFunc
	_, err = ScheduleFuncWithOptions("* * * * *", func() {}, WithDefaultScheduler(s))
	if err != nil {
		t.Fatal(err)
	}
	// ScheduleWithID
	err = ScheduleWithIDWithOptions("fixed-id", "* * * * *", TaskFunc(func() {}), WithDefaultScheduler(s))
	if err != nil {
		t.Fatal(err)
	}
	if s.Size() != 3 {
		t.Fatalf("expected 3 tasks, got %d", s.Size())
	}
	// Remove
	if !RemoveWithOptions("fixed-id", WithDefaultScheduler(s)) {
		t.Fatal("RemoveWithOptions failed")
	}
}

func TestRestartWithOptions(t *testing.T) {
	s := NewScheduler()
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if err := RestartWithOptions(WithDefaultScheduler(s)); err != nil {
		t.Fatal(err)
	}
	s.Stop()
}

func TestStopShutdownWithOptions(t *testing.T) {
	s := NewScheduler()
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	StopWithOptions(WithDefaultScheduler(s))
	// Stopped scheduler can be shutdown
	if err := ShutdownWithOptions(context.Background(), WithDefaultScheduler(s)); err != nil {
		t.Fatal(err)
	}
}

func TestDayOfMonthMatcherIsLast(t *testing.T) {
	m := newDayOfMonthMatcher([]int{1, 15})
	if m.IsLast() {
		t.Fatal("should not have L sentinel")
	}
	m2 := newDayOfMonthMatcher([]int{1, 32}) // 32 = lastDayOfMonthSentinel
	if !m2.IsLast() {
		t.Fatal("should have L sentinel")
	}
}

func TestDayOfMonthMatcherMatchDay(t *testing.T) {
	m := newDayOfMonthMatcher([]int{15})
	if !m.MatchDay(15, 1, false) {
		t.Fatal("MatchDay 15 should be true")
	}
	if m.MatchDay(10, 1, false) {
		t.Fatal("MatchDay 10 should be false")
	}
	// L sentinel for last day
	mL := newDayOfMonthMatcher([]int{1, 32})
	if !mL.MatchDay(31, 1, false) {
		t.Fatal("MatchDay 31 (January) should match L")
	}
	if !mL.MatchDay(28, 2, false) {
		t.Fatal("MatchDay 28 (Feb non-leap) should match L (28 is last day)")
	}
	if !mL.MatchDay(29, 2, true) {
		t.Fatal("MatchDay 29 (Feb leap) should match L")
	}
	// day that is neither explicit nor last
	if mL.MatchDay(15, 1, false) {
		t.Fatal("MatchDay 15 (January) should not match L or explicit values")
	}
}

// TestDefaultSchedulerWrappers tests the non-option wrapper functions that delegate to *WithOptions.
func TestDefaultSchedulerWrappers(t *testing.T) {
	prev := ConfigureDefaultScheduler()
	prev.Stop()
	defer prev.Start()

	s := getDefaultScheduler()
	SetMatchSecond(true)
	if !s.IsMatchSecond() {
		t.Fatal("SetMatchSecond failed")
	}
	if err := SetMatchSecondE(false); err != nil {
		t.Fatal(err)
	}
	if s.IsMatchSecond() {
		t.Fatal("SetMatchSecondE(false) failed")
	}

	s.SetMatchSecond(true)
	id, err := Schedule("* * * * * *", TaskFunc(func() {}))
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("Schedule returned empty id")
	}

	id2, err := ScheduleFunc("* * * * * *", func() {})
	if err != nil {
		t.Fatal(err)
	}
	if id2 == "" {
		t.Fatal("ScheduleFunc returned empty id")
	}

	if err := ScheduleWithID("fixed-id", "* * * * * *", TaskFunc(func() {})); err != nil {
		t.Fatal(err)
	}
	if s.Size() != 3 {
		t.Fatalf("expected 3 tasks, got %d", s.Size())
	}

	if !Remove("fixed-id") {
		t.Fatal("Remove failed")
	}
	if s.Size() != 2 {
		t.Fatalf("expected 2 tasks after remove, got %d", s.Size())
	}

	if err := UpdatePattern(id, "0 0 * * *"); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultSchedulerLifecycleWrappers(t *testing.T) {
	prev := ConfigureDefaultScheduler()
	prev.Stop()
	defer func() {
		_ = prev.Start()
	}()

	if err := Start(); err != nil {
		t.Fatal(err)
	}
	Stop()
	// Shutdown on stopped scheduler
	if err := Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := ShutdownWithOptions(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := Start(); err != nil {
		t.Fatal(err)
	}
	if err := Restart(); err != nil {
		t.Fatal(err)
	}
	Stop()
}

func TestConfigureDefaultScheduler(t *testing.T) {
	s := ConfigureDefaultScheduler()
	if s == nil {
		t.Fatal("ConfigureDefaultScheduler returned nil")
	}
	s.Stop()
}

func TestDefaultSchedulerWrappersOnNilDefault(t *testing.T) {
	defaultMu.Lock()
	old := defaultScheduler
	defaultScheduler = nil
	defaultMu.Unlock()
	defer func() {
		defaultMu.Lock()
		defaultScheduler = old
		defaultMu.Unlock()
	}()

	// Should not panic, should create a new scheduler
	s := DefaultScheduler()
	if s == nil {
		t.Fatal("DefaultScheduler should not return nil")
	}
	s.Stop()
}
