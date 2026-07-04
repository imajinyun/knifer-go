package cron

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultSchedulerConcurrentConfigureAndOperations(t *testing.T) {
	t.Cleanup(func() {
		ConfigureDefaultScheduler()
		Stop()
	})

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(4)
		go func(i int) {
			defer wg.Done()
			ConfigureDefaultScheduler(WithIDGenerator(func() string { return "default-concurrent" }))
		}(i)
		go func(i int) {
			defer wg.Done()
			_, _ = ScheduleFuncWithOptions("* * * * *", func() {}, WithDefaultSchedulerOptions(
				WithIDGenerator(func() string { return "isolated-concurrent" }),
			))
		}(i)
		go func(i int) {
			defer wg.Done()
			_ = StartWithOptions(WithDefaultSchedulerOptions(
				WithRunner(func(fn func()) {}),
			))
		}(i)
		go func(i int) {
			defer wg.Done()
			StopWithOptions(WithDefaultSchedulerOptions())
		}(i)
	}
	wg.Wait()
}

func TestTaskTableConcurrentMutationAndSnapshots(t *testing.T) {
	table := NewTaskTable()
	pattern := MustNewPattern("* * * * *")

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		id := fmt.Sprintf("task-%d", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = table.Add(id, pattern, TaskFunc(func() {}))
			_ = table.GetTask(id)
			_ = table.GetPattern(id)
			_ = table.UpdatePattern(id, pattern)
			_ = table.IDs()
			_ = table.Size()
			_ = table.Remove(id)
		}()
	}
	wg.Wait()
}

func TestListenerCallbacksCanReenterManager(t *testing.T) {
	s := NewSchedulerWithOptions(WithExecutor(func(fn func()) { fn() }))
	defer s.Stop()

	var starts atomic.Int32
	var reentrant TaskListener
	reentrant = &testListener{started: &starts}
	s.AddListener(testTaskListenerFunc{
		onStart: func(e *TaskExecutor) {
			s.RemoveListener(reentrant)
			s.AddListener(reentrant)
		},
	})
	s.AddListener(reentrant)

	done := make(chan struct{})
	s.AddListener(testTaskListenerFunc{
		onSucceeded: func(*TaskExecutor) { close(done) },
	})
	s.executorMgr.spawn(NewCronTask("listener-reentry", MustNewPattern("* * * * *"), TaskFunc(func() {})))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener callback reentry blocked task completion")
	}
	if starts.Load() == 0 {
		t.Fatal("reentrant listener did not observe start")
	}
}

func TestShutdownContextCanRaceWithStartStop(t *testing.T) {
	s := NewSchedulerWithOptions(WithRunner(func(fn func()) { go fn() }))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = s.Start()
		}()
		go func() {
			defer wg.Done()
			s.Stop()
		}()
		go func() {
			defer wg.Done()
			_ = s.Shutdown(ctx)
		}()
	}
	wg.Wait()
}

type testTaskListenerFunc struct {
	onStart     func(*TaskExecutor)
	onSucceeded func(*TaskExecutor)
	onFailed    func(*TaskExecutor, any)
}

func (l testTaskListenerFunc) OnStart(e *TaskExecutor) {
	if l.onStart != nil {
		l.onStart(e)
	}
}

func (l testTaskListenerFunc) OnSucceeded(e *TaskExecutor) {
	if l.onSucceeded != nil {
		l.onSucceeded(e)
	}
}

func (l testTaskListenerFunc) OnFailed(e *TaskExecutor, err any) {
	if l.onFailed != nil {
		l.onFailed(e, err)
	}
}
