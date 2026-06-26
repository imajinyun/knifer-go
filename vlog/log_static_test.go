package vlog_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vlog"
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

func TestFacadeStaticLogAllOptions(t *testing.T) {
	old := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelAll)
	t.Cleanup(func() { vlog.SetLogLevel(old) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	opts := []vlog.LoggerOption{vlog.WithLoggerConsoleOptions(
		vlog.WithLogOutput(out, errOut),
		vlog.WithLogClock(func() time.Time { return time.Date(2024, 8, 9, 10, 11, 12, 0, time.UTC) }),
		vlog.WithLogTimeLayout("15:04:05"),
	)}
	vlog.TraceWithOptions(opts, "trace option")
	vlog.TracefWithOptions(opts, "trace %s", "format")
	vlog.DebugWithOptions(opts, "debug option")
	vlog.DebugfWithOptions(opts, "debug %s", "format")
	vlog.InfofWithOptions(opts, "info %s", "format")
	vlog.WarnWithOptions(opts, "warn option")
	vlog.WarnfWithOptions(opts, "warn %s", "format")
	vlog.ErrorLogWithOptions(opts, "error option")
	vlog.ErrorfWithOptions(opts, "error %s", "format")
	vlog.LogAt(vlog.LogLevelInfo, "logat %s", "global")
	vlog.LogAtWithOptions(opts, vlog.LogLevelInfo, "logat %s", "option")
	vlog.LogAtE(vlog.LogLevelError, errors.New("global boom"), "loge %s", "global")
	vlog.LogAtEWithOptions(opts, vlog.LogLevelError, errors.New("option boom"), "loge %s", "option")

	stdout := out.String()
	for _, want := range []string{"trace option", "trace format", "debug option", "debug format", "info format", "logat option", "10:11:12"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q: %q", want, stdout)
		}
	}
	stderr := errOut.String()
	for _, want := range []string{"warn option", "warn format", "error option", "error format", "loge option", "option boom"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q: %q", want, stderr)
		}
	}
}
