package vcron_test

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vcron"
)

func ExampleNewPattern() {
	p, err := vcron.NewPattern("* * * * *")
	fmt.Println(p != nil)
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}

func ExampleMustNewPattern() {
	p := vcron.MustNewPattern("0 9 * * *")
	t := time.Date(2026, 6, 15, 9, 0, 30, 0, time.UTC)

	fmt.Println(p.Raw())
	fmt.Println(p.Match(t, false))
	// Output:
	// 0 9 * * *
	// true
}

func ExampleNewConfigWithOptions() {
	loc := time.FixedZone("docs", 8*60*60)
	cfg := vcron.NewConfigWithOptions(
		vcron.WithConfigLocation(loc),
		vcron.WithConfigMatchSecond(true),
	)

	fmt.Println(cfg.Location.String())
	fmt.Println(cfg.MatchSecond)
	// Output:
	// docs
	// true
}

func ExampleNewSchedulerWithOptions() {
	s := vcron.NewSchedulerWithOptions(
		vcron.WithIDGenerator(func() string { return "job-1" }),
	)

	id, err := s.ScheduleFunc("* * * * *", func() {})

	fmt.Println(id, s.Size())
	fmt.Println(err)
	// Output:
	// job-1 1
	// <nil>
}

func ExamplePart_CheckValue() {
	fmt.Println(vcron.PartMinute.CheckValue(59) == nil)
	fmt.Println(vcron.PartMinute.CheckValue(60) != nil)
	// Output:
	// true
	// true
}

func ExampleNewCronError() {
	err := vcron.NewCronError("invalid %s", "pattern")

	fmt.Println(err.Error())
	// Output: invalid pattern
}

func ExampleCronScheduleFuncWithOptions() {
	s := vcron.NewSchedulerWithOptions(vcron.WithIDGenerator(func() string { return "job-1" }))
	id, err := vcron.CronScheduleFuncWithOptions("* * * * *", func() {}, vcron.WithDefaultScheduler(s))

	fmt.Println(id, s.Size())
	fmt.Println(err)
	// Output:
	// job-1 1
	// <nil>
}

func ExampleCronRemoveWithOptions() {
	s := vcron.NewScheduler()
	err := vcron.CronScheduleWithIDWithOptions("job-1", "* * * * *", vcron.TaskFunc(func() {}), vcron.WithDefaultScheduler(s))
	removed := vcron.CronRemoveWithOptions("job-1", vcron.WithDefaultScheduler(s))

	fmt.Println(removed, s.Size())
	fmt.Println(err)
	// Output:
	// true 0
	// <nil>
}

func ExampleCronUpdatePatternWithOptions() {
	s := vcron.NewScheduler()
	err := vcron.CronScheduleWithIDWithOptions("job-1", "* * * * *", vcron.TaskFunc(func() {}), vcron.WithDefaultScheduler(s))
	updateErr := vcron.CronUpdatePatternWithOptions("job-1", "0 9 * * *", vcron.WithDefaultScheduler(s))

	fmt.Println(s.Size())
	fmt.Println(err)
	fmt.Println(updateErr)
	// Output:
	// 1
	// <nil>
	// <nil>
}
