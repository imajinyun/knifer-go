package vlog_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vlog"
)

func TestFacadeStaticLogWithOptions(t *testing.T) {
	old := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelInfo)
	defer vlog.SetLogLevel(old)

	out := &bytes.Buffer{}
	fixed := time.Date(2024, 7, 8, 9, 10, 11, 0, time.UTC)
	vlog.InfoWithOptions([]vlog.LoggerOption{vlog.WithLoggerConsoleOptions(
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(out, &bytes.Buffer{}),
	)}, "facade-static")
	if !strings.Contains(out.String(), "2024-07-08T09:10:11Z") || !strings.Contains(out.String(), "facade-static") {
		t.Fatalf("static log options not applied: %q", out.String())
	}
}

func TestFacadeStaticLog(t *testing.T) {
	// smoke test: static log functions should not panic
	vlog.Trace("static trace")
	vlog.Debug("static debug")
	vlog.Info("static info")
	vlog.Warn("static warn")
	vlog.ErrorLog("static error")
	vlog.Tracef("formatted %s", "trace")
	vlog.Debugf("formatted %s", "debug")
	vlog.Infof("formatted %s", "info")
	vlog.Warnf("formatted %s", "warn")
	vlog.Errorf("formatted %s", "error")
}
