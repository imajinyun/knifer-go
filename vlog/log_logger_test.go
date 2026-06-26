package vlog_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vlog"
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

func TestFacadeLoggerGeneratedDelegates(t *testing.T) {
	vlog.SetFactory(vlog.LogFactoryFunc(func(name string) vlog.Log { return vlog.NewConsoleLog("delegate:" + name) }))
	t.Cleanup(func() {
		vlog.SetFactory(vlog.LogFactoryFunc(func(name string) vlog.Log { return vlog.NewConsoleLog(name) }))
	})

	if vlog.GetFactory() == nil {
		t.Fatal("GetFactory should return configured factory")
	}
	if got := vlog.Get("facade.delegate"); got.GetName() != "delegate:facade.delegate" {
		t.Fatalf("Get delegate name = %q", got.GetName())
	}
	if got := vlog.GetWithOptions("facade.direct", vlog.WithLoggerFactory(vlog.LogFactoryFunc(func(name string) vlog.Log {
		return vlog.NewConsoleLog("direct:" + name)
	}))); got.GetName() != "direct:facade.direct" {
		t.Fatalf("GetWithOptions delegate name = %q", got.GetName())
	}
	if got := vlog.GetDefault(); got.GetName() != "delegate:default" {
		t.Fatalf("GetDefault delegate name = %q", got.GetName())
	}
	if got := vlog.GetDefaultWithOptions(vlog.WithLoggerFactory(vlog.LogFactoryFunc(func(name string) vlog.Log {
		return vlog.NewConsoleLog("default-option:" + name)
	}))); got.GetName() != "default-option:default" {
		t.Fatalf("GetDefaultWithOptions name = %q", got.GetName())
	}
}

func TestFacadeLoggerCacheOption(t *testing.T) {
	created := 0
	vlog.SetFactory(vlog.LogFactoryFunc(func(name string) vlog.Log {
		created++
		return vlog.NewConsoleLog(name)
	}))
	t.Cleanup(func() {
		vlog.SetFactory(vlog.LogFactoryFunc(func(name string) vlog.Log { return vlog.NewConsoleLog(name) }))
	})

	_ = vlog.LoggerWithOptions("facade.no-cache", vlog.WithLoggerCache(false))
	_ = vlog.LoggerWithOptions("facade.no-cache", vlog.WithLoggerCache(false))
	if created != 2 {
		t.Fatalf("WithLoggerCache(false) created %d loggers, want 2", created)
	}
}
