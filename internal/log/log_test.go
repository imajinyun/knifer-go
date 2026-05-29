package log

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestLevelString(t *testing.T) {
	cases := map[Level]string{
		LevelAll:   "ALL",
		LevelTrace: "TRACE",
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
		LevelOff:   "OFF",
	}
	for l, want := range cases {
		if got := l.String(); got != want {
			t.Errorf("Level(%d).String()=%q, want %q", l, got, want)
		}
	}
}

func TestFormatTemplatePlaceholder(t *testing.T) {
	template := strings.Join([]string{"hello {}", "age={}"}, ", ")
	args := []any{"world", 18}
	got := renderLogMessage(template, args...)
	want := "hello world, age=18"
	if got != want {
		t.Errorf("formatTemplate placeholder got=%q want=%q", got, want)
	}
}

func TestFormatTemplatePrintfFallback(t *testing.T) {
	got := renderLogMessage("a=%d, b=%s", 1, "x")
	want := "a=1, b=x"
	if got != want {
		t.Errorf("formatTemplate printf got=%q want=%q", got, want)
	}
}

func TestFormatTemplateNoArgs(t *testing.T) {
	if got := renderLogMessage("plain"); got != "plain" {
		t.Errorf("formatTemplate plain got=%q", got)
	}
}

func TestFormatTemplateConcat(t *testing.T) {
	args := []any{"a", "b", 1}
	if got := renderLogMessage("", args...); got != "ab1" {
		t.Errorf("formatTemplate concat got=%q", got)
	}
}

func TestFormatTemplateExtraPlaceholders(t *testing.T) {
	template := strings.Repeat("{}-", 2) + "{}"
	args := []any{"a"}
	got := renderLogMessage(template, args...)
	if got != "a-{}-{}" {
		t.Errorf("formatTemplate extra got=%q", got)
	}
}

func newTestConsoleLog(name string) (*ConsoleLog, *bytes.Buffer, *bytes.Buffer) {
	c := NewConsoleLog(name)
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	c.SetOutput(out, errOut)
	return c, out, errOut
}

func TestConsoleLogLevels(t *testing.T) {
	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelDebug)
	defer SetConsoleLevel(prevLevel)

	c, out, errOut := newTestConsoleLog("test.console")
	c.Debug("debug msg")
	c.Infof("user={}", "alice")
	c.Warnf("warn-{}-{}", 1, 2)
	c.Errorf("err=%d", 7)

	stdoutText := out.String()
	stderrText := errOut.String()

	if !strings.Contains(stdoutText, "[DEBUG]") || !strings.Contains(stdoutText, "debug msg") {
		t.Errorf("expected debug log in stdout, got %q", stdoutText)
	}
	if !strings.Contains(stdoutText, "[INFO ]") || !strings.Contains(stdoutText, "user=alice") {
		t.Errorf("expected info log with placeholder, got %q", stdoutText)
	}
	if !strings.Contains(stderrText, "[WARN ]") || !strings.Contains(stderrText, "warn-1-2") {
		t.Errorf("expected warn log in stderr, got %q", stderrText)
	}
	if !strings.Contains(stderrText, "[ERROR]") || !strings.Contains(stderrText, "err=7") {
		t.Errorf("expected error log in stderr, got %q", stderrText)
	}
}

func TestConsoleLogFiltering(t *testing.T) {
	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelWarn)
	defer SetConsoleLevel(prevLevel)

	c, out, errOut := newTestConsoleLog("test.filter")
	c.Debug("should be filtered")
	c.Info("should be filtered")
	c.Warn("kept warn")

	if out.Len() != 0 {
		t.Errorf("expected no debug/info output, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "kept warn") {
		t.Errorf("expected warn output, got %q", errOut.String())
	}
	if !c.IsWarnEnabled() || c.IsDebugEnabled() {
		t.Errorf("level checks wrong: warn=%v debug=%v", c.IsWarnEnabled(), c.IsDebugEnabled())
	}
}

func TestConsoleLogWithError(t *testing.T) {
	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelDebug)
	defer SetConsoleLevel(prevLevel)

	c, _, errOut := newTestConsoleLog("test.err")
	c.LogE(LevelError, errors.New("boom"), "operation {} failed", "save")
	got := errOut.String()
	if !strings.Contains(got, "operation save failed") || !strings.Contains(got, "error: boom") {
		t.Errorf("error formatted output unexpected: %q", got)
	}
}

func TestConsoleColorLog(t *testing.T) {
	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelDebug)
	defer SetConsoleLevel(prevLevel)

	c := NewConsoleColorLog("test.color")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	c.SetOutput(out, errOut)

	c.Info("hi")
	c.Warn("careful")

	if !strings.Contains(out.String(), "hi") || !strings.Contains(out.String(), "\033[") {
		t.Errorf("color info expected ANSI codes, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "careful") {
		t.Errorf("color warn output unexpected: %q", errOut.String())
	}
}

func TestSetColorFactory(t *testing.T) {
	called := false
	SetColorFactory(func(level Level) string {
		called = true
		return ansiBlue
	})
	defer SetColorFactory(defaultColorFactory)

	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelDebug)
	defer SetConsoleLevel(prevLevel)

	c := NewConsoleColorLog("test.colorfactory")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	c.SetOutput(out, errOut)
	c.Info("x")

	if !called {
		t.Error("custom color factory was not called")
	}
	if !strings.Contains(out.String(), ansiBlue) {
		t.Errorf("custom color expected, got %q", out.String())
	}
}

func TestLogFactoryCache(t *testing.T) {
	a := Get("cache.same")
	b := Get("cache.same")
	if a != b {
		t.Error("expected cached Log instance to be identical")
	}
	c := Get("cache.different")
	if a == c {
		t.Error("expected different name to produce different instance")
	}
}

func TestSetFactoryReplacesCache(t *testing.T) {
	first := Get("factory.test")
	firstName := first.GetName()

	var lock sync.Mutex
	created := 0
	SetFactory(LogFactoryFunc(func(name string) Log {
		lock.Lock()
		created++
		lock.Unlock()
		// 自定义工厂返回带前缀名的 ConsoleLog，便于区分。
		return NewConsoleLog("custom:" + name)
	}))
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	second := Get("factory.test")
	if first == second {
		t.Error("expected new factory to produce a new instance")
	}
	if firstName == second.GetName() {
		t.Errorf("expected different name from custom factory, both got %q", firstName)
	}
	if !strings.HasPrefix(second.GetName(), "custom:") {
		t.Errorf("expected custom factory output, got name=%q", second.GetName())
	}
	if created == 0 {
		t.Error("expected custom factory to be invoked")
	}
}

func TestStaticLogPipeline(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	SetFactory(LogFactoryFunc(func(name string) Log {
		c := NewConsoleLog(name)
		c.SetOutput(out, errOut)
		return c
	}))
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelTrace)
	defer SetConsoleLevel(prevLevel)

	Tracef("trace {}", 1)
	Debugf("debug {}", 2)
	Infof("info {}", 3)
	Warnf("warn {}", 4)
	Errorf("error {}", 5)

	stdout := out.String()
	stderr := errOut.String()
	for _, want := range []string{"trace 1", "debug 2", "info 3"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q in %q", want, stdout)
		}
	}
	for _, want := range []string{"warn 4", "error 5"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q in %q", want, stderr)
		}
	}
}

func TestLogAtAndLogAtE(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	SetFactory(LogFactoryFunc(func(name string) Log {
		c := NewConsoleLog(name)
		c.SetOutput(out, errOut)
		return c
	}))
	defer SetFactory(LogFactoryFunc(func(name string) Log { return NewConsoleLog(name) }))

	prevLevel := GetConsoleLevel()
	SetConsoleLevel(LevelDebug)
	defer SetConsoleLevel(prevLevel)

	LogAt(LevelInfo, "hello {}", "world")
	LogAtE(LevelError, errors.New("oops"), "boom {}", "now")

	if !strings.Contains(out.String(), "hello world") {
		t.Errorf("LogAt info missing: %q", out.String())
	}
	if !strings.Contains(errOut.String(), "boom now") || !strings.Contains(errOut.String(), "error: oops") {
		t.Errorf("LogAtE error missing: %q", errOut.String())
	}
}
