package vlog_test

import (
	"testing"

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

func TestFacadeLoggerWithOptions(t *testing.T) {
	log := vlog.LoggerWithOptions("facade.logger.option", vlog.WithLoggerFactory(vlog.LogFactoryFunc(func(name string) vlog.Log {
		return vlog.NewConsoleLog("facade:" + name)
	})))
	if log.GetName() != "facade:facade.logger.option" {
		t.Fatalf("LoggerWithOptions name = %q", log.GetName())
	}

	vlog.SetLogFactory(vlog.LogFactoryFunc(func(name string) vlog.Log { return vlog.NewConsoleLog("global:" + name) }))
	defer vlog.SetLogFactory(vlog.LogFactoryFunc(func(name string) vlog.Log { return vlog.NewConsoleLog(name) }))
	isolated := vlog.NewIsolatedLogger("facade.isolated")
	if isolated.GetName() != "facade.isolated" {
		t.Fatalf("NewIsolatedLogger leaked global factory: %q", isolated.GetName())
	}
}
