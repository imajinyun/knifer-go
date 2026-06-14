package vlog_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vlog"
)

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

	customColorOut := &bytes.Buffer{}
	customColorLog := vlog.NewConsoleColorLogWithOptions("facade.color.custom",
		vlog.WithLogOutput(customColorOut, &bytes.Buffer{}),
		vlog.WithLogColorFactory(func(vlog.Level) string { return "\033[36m" }),
	)
	customColorLog.Info("custom-color")
	if !strings.Contains(customColorOut.String(), "\033[36m") || !strings.Contains(customColorOut.String(), "custom-color") {
		t.Fatalf("color factory option not applied: %q", customColorOut.String())
	}
}
