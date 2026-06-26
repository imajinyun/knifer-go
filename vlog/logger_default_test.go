package vlog_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vlog"
)

func TestFacadeDefaultLoggerWithOptions(t *testing.T) {
	log := vlog.DefaultLoggerWithOptions(vlog.WithLoggerFactory(vlog.LogFactoryFunc(func(name string) vlog.Log {
		return vlog.NewConsoleLog("default-opt:" + name)
	})))
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	if log.GetName() != "default-opt:default" {
		t.Fatalf("DefaultLoggerWithOptions name = %q", log.GetName())
	}
}
