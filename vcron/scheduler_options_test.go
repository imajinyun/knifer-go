package vcron_test

import (
	"bytes"
	"strconv"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vcron"
)

func TestFacadeSchedulerWithOptions(t *testing.T) {
	loc := time.FixedZone("facade", 8*60*60)
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := vcron.NewSchedulerWithOptions(
		vcron.WithLocation(loc),
		vcron.WithMatchSecond(true),
		vcron.WithIDGenerator(func() string { return "facade-task" }),
		vcron.WithClock(func() time.Time { return now }),
		vcron.WithSleeper(func(d time.Duration, stopCh <-chan struct{}) bool {
			now = now.Add(d)
			return true
		}),
		vcron.WithExecutor(func(fn func()) { fn() }),
	)
	if s.Config().Location != loc {
		t.Fatalf("scheduler location = %v, want %v", s.Config().Location, loc)
	}
	if !s.IsMatchSecond() {
		t.Fatal("scheduler should match seconds")
	}
	id, err := s.ScheduleFunc("* * * * * *", func() {})
	if err != nil {
		t.Fatalf("ScheduleFunc with options: %v", err)
	}
	if id != "facade-task" {
		t.Fatalf("scheduled id = %q, want facade-task", id)
	}
}

func TestFacadeSchedulerIDRandomReaderOption(t *testing.T) {
	s := vcron.NewSchedulerWithOptions(vcron.WithIDRandomReader(bytes.NewReader([]byte{8, 7, 6, 5, 4, 3, 2, 1})))
	id, err := s.ScheduleFunc("* * * * *", func() {})
	if err != nil {
		t.Fatalf("ScheduleFunc: %v", err)
	}
	if id != "0807060504030201" {
		t.Fatalf("id = %q, want 0807060504030201", id)
	}
}

func TestFacadeSchedulerPatternOptions(t *testing.T) {
	s := vcron.NewSchedulerWithOptions(
		vcron.WithSchedulerPatternOptions(vcron.WithPatternIntParser(func(s string) (int, error) {
			if s == "custom" {
				return 5, nil
			}
			return strconv.Atoi(s)
		})),
	)
	if err := s.ScheduleWithID("custom", "* custom * * *", vcron.TaskFunc(func() {})); err != nil {
		t.Fatalf("ScheduleWithID with scheduler pattern options: %v", err)
	}
	if s.GetPattern("custom") == nil || s.GetTask("custom") == nil {
		t.Fatal("scheduled task should be retrievable")
	}
}

func TestFacadeTaskConstructors(t *testing.T) {
	pattern := vcron.MustNewPattern("* * * * *")
	executed := false
	task := vcron.TaskFunc(func() { executed = true })
	cronTask := vcron.NewCronTask("task-id", pattern, task)
	if cronTask == nil || cronTask.ID() != "task-id" || cronTask.Pattern() != pattern || cronTask.Raw() == nil {
		t.Fatalf("NewCronTask = %#v", cronTask)
	}
	cronTask.Execute()
	if !executed {
		t.Fatal("CronTask.Execute should delegate to raw task")
	}
	table := vcron.NewTaskTable()
	if table == nil || !table.IsEmpty() {
		t.Fatalf("NewTaskTable = %#v", table)
	}
	if err := table.Add("task-id", pattern, task); err != nil || table.Size() != 1 {
		t.Fatalf("TaskTable.Add err=%v size=%d", err, table.Size())
	}
}
