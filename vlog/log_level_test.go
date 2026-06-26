package vlog_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vlog"
)

func TestFacadeLogLevel(t *testing.T) {
	old := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelDebug)
	if vlog.GetLogLevel() != vlog.LogLevelDebug {
		t.Fatal("expected log level to be set to Debug")
	}
	vlog.SetLogLevel(old)
}

func TestFacadeConsoleLevelAliasesAndStrings(t *testing.T) {
	old := vlog.GetConsoleLevel()
	t.Cleanup(func() { vlog.SetConsoleLevel(old) })

	vlog.SetConsoleLevel(vlog.LogLevelWarn)
	if vlog.GetConsoleLevel() != vlog.LogLevelWarn || vlog.GetLogLevel() != vlog.LogLevelWarn {
		t.Fatalf("console level aliases mismatch: console=%v log=%v", vlog.GetConsoleLevel(), vlog.GetLogLevel())
	}

	cases := map[vlog.Level]string{
		vlog.LogLevelAll:   "ALL",
		vlog.LogLevelTrace: "TRACE",
		vlog.LogLevelDebug: "DEBUG",
		vlog.LogLevelInfo:  "INFO",
		vlog.LogLevelWarn:  "WARN",
		vlog.LogLevelError: "ERROR",
		vlog.LogLevelFatal: "FATAL",
		vlog.LogLevelOff:   "OFF",
		vlog.Level(99):     "UNKNOWN",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Fatalf("%v.String() = %q, want %q", int(level), got, want)
		}
	}
}
