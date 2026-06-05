package vlog_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vlog"
)

func TestFacadeLogger(t *testing.T) {
	log := vlog.NewConsoleLog("test")
	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	// smoke test: log at each level should not panic
	log.Trace("trace")
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
}

func TestFacadeDefaultLogger(t *testing.T) {
	log := vlog.DefaultLogger()
	if log == nil {
		t.Fatal("expected non-nil default logger")
	}
	log.Info("default logger works")
}

func TestFacadeLoggerByName(t *testing.T) {
	log1 := vlog.Logger("foo")
	log2 := vlog.Logger("foo")
	if log1 == nil || log2 == nil {
		t.Fatal("expected non-nil loggers")
	}
}

func TestFacadeLogLevel(t *testing.T) {
	old := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelDebug)
	if vlog.GetLogLevel() != vlog.LogLevelDebug {
		t.Fatal("expected log level to be set to Debug")
	}
	vlog.SetLogLevel(old)
}

func TestFacadeConsoleLogOptions(t *testing.T) {
	old := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelDebug)
	defer vlog.SetLogLevel(old)

	out := &bytes.Buffer{}
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)
	log := vlog.NewConsoleLogWithOptions("facade.options",
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogOutput(out, &bytes.Buffer{}),
	)
	log.Info("hello")
	if !strings.Contains(out.String(), "2024-04-05T06:07:08Z") || !strings.Contains(out.String(), "hello") {
		t.Fatalf("console log options not applied: %q", out.String())
	}

	colorOut := &bytes.Buffer{}
	colorLog := vlog.NewConsoleColorLogWithOptions("facade.color",
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout("15:04"),
		vlog.WithLogOutput(colorOut, &bytes.Buffer{}),
	)
	colorLog.Info("color")
	if !strings.Contains(colorOut.String(), "06:07") || !strings.Contains(colorOut.String(), "color") {
		t.Fatalf("color log options not applied: %q", colorOut.String())
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
