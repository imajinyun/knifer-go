package vcron_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vcron"
)

func TestFacadeDefaultSchedulerOptions(t *testing.T) {
	global := vcron.ConfigureDefaultScheduler(vcron.WithIDGenerator(func() string { return "global-id" }))
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })
	isolated := vcron.NewSchedulerWithOptions(vcron.WithIDGenerator(func() string { return "facade-isolated" }))

	id, err := vcron.CronScheduleFuncWithOptions("* * * * *", func() {}, vcron.WithDefaultScheduler(isolated))
	if err != nil {
		t.Fatalf("CronScheduleFuncWithOptions: %v", err)
	}
	if id != "facade-isolated" || isolated.Size() != 1 || global.Size() != 0 {
		t.Fatalf("default scheduler option not isolated: id=%q isolated=%d global=%d", id, isolated.Size(), global.Size())
	}
	if !vcron.CronRemoveWithOptions(id, vcron.WithDefaultScheduler(isolated)) {
		t.Fatal("CronRemoveWithOptions should remove isolated task")
	}
}

func TestFacadeDefaultSchedulerGeneratedOptions(t *testing.T) {
	global := vcron.ConfigureDefaultScheduler(vcron.WithIDGenerator(func() string { return "global-id" }))
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })

	selected := vcron.DefaultSchedulerWithOptions(vcron.WithDefaultSchedulerOptions(vcron.WithIDGenerator(func() string { return "isolated-id" })))
	if selected == nil || selected == global {
		t.Fatal("DefaultSchedulerWithOptions should return isolated scheduler from options")
	}
	id, err := vcron.CronScheduleFuncWithOptions("* * * * *", func() {}, vcron.WithDefaultScheduler(selected))
	if err != nil {
		t.Fatalf("CronScheduleFuncWithOptions isolated: %v", err)
	}
	if id != "isolated-id" || selected.Size() != 1 || global.Size() != 0 {
		t.Fatalf("isolated scheduler mismatch: id=%q selected=%d global=%d", id, selected.Size(), global.Size())
	}
	if vcron.DefaultScheduler() != global {
		t.Fatal("DefaultScheduler should still return configured global scheduler")
	}
}

func TestFacadeDefaultSchedulerDelegates(t *testing.T) {
	s := vcron.ConfigureDefaultScheduler(
		vcron.WithMatchSecond(true),
		vcron.WithIDGenerator(func() string { return "auto-id" }),
	)
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })

	if err := vcron.CronScheduleWithID("manual", "* * * * * *", vcron.TaskFunc(func() {})); err != nil {
		t.Fatalf("CronScheduleWithID: %v", err)
	}
	id, err := vcron.CronSchedule("* * * * * *", vcron.TaskFunc(func() {}))
	if err != nil || id != "auto-id" {
		t.Fatalf("CronSchedule = %q, %v", id, err)
	}
	if s.Size() != 2 {
		t.Fatalf("scheduled task count = %d, want 2", s.Size())
	}
	if err := vcron.CronUpdatePattern("manual", "*/2 * * * * *"); err != nil {
		t.Fatalf("CronUpdatePattern: %v", err)
	}
	if err := vcron.CronUpdatePatternWithOptions("auto-id", "*/3 * * * * *", vcron.WithDefaultScheduler(s)); err != nil {
		t.Fatalf("CronUpdatePatternWithOptions: %v", err)
	}
	if !vcron.CronRemove("manual") || !vcron.CronRemove("auto-id") || !s.IsEmpty() {
		t.Fatalf("remove delegates failed, size=%d", s.Size())
	}
}

func TestFacadeCronScheduleWithOptions(t *testing.T) {
	s := vcron.ConfigureDefaultScheduler(
		vcron.WithMatchSecond(true),
		vcron.WithIDGenerator(func() string { return "sched-id" }),
	)
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })

	id, err := vcron.CronScheduleWithOptions("* * * * * *", vcron.TaskFunc(func() {}), vcron.WithDefaultScheduler(s))
	if err != nil || id != "sched-id" || s.Size() != 1 {
		t.Fatalf("CronScheduleWithOptions id=%q err=%v", id, err)
	}
	vcron.CronRemoveWithOptions(id, vcron.WithDefaultScheduler(s))
}

func TestFacadeCronScheduleWithIDWithOptions(t *testing.T) {
	s := vcron.ConfigureDefaultScheduler(
		vcron.WithMatchSecond(true),
	)
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })

	err := vcron.CronScheduleWithIDWithOptions("my-id", "* * * * * *", vcron.TaskFunc(func() {}), vcron.WithDefaultScheduler(s))
	if err != nil || s.Size() != 1 {
		t.Fatalf("CronScheduleWithIDWithOptions err=%v", err)
	}
	if !vcron.CronRemoveWithOptions("my-id", vcron.WithDefaultScheduler(s)) {
		t.Fatal("CronRemoveWithOptions should succeed")
	}
}

func TestFacadeCronSetMatchSecondWithOptions(t *testing.T) {
	s := vcron.ConfigureDefaultScheduler(vcron.WithMatchSecond(false))
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })

	vcron.CronSetMatchSecondWithOptions(true, vcron.WithDefaultScheduler(s))
	if !s.IsMatchSecond() {
		t.Fatal("CronSetMatchSecondWithOptions should set match second on target scheduler")
	}
}
