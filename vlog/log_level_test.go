package vlog_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vlog"
)

func TestFacadeLogLevel(t *testing.T) {
	old := vlog.GetLogLevel()
	vlog.SetLogLevel(vlog.LogLevelDebug)
	if vlog.GetLogLevel() != vlog.LogLevelDebug {
		t.Fatal("expected log level to be set to Debug")
	}
	vlog.SetLogLevel(old)
}
