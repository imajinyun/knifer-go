package socket

import (
	"net"
	"runtime"
	"testing"
	"time"
)

func closeAndReport(t *testing.T, closeFn func() error) {
	t.Helper()
	if err := closeFn(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestSocketConfigDefaults(t *testing.T) {
	cfg := NewSocketConfig()
	if cfg.ThreadPoolSize != runtime.NumCPU() {
		t.Errorf("ThreadPoolSize 默认应为 CPU 核数，实际 %d", cfg.ThreadPoolSize)
	}
	if cfg.ReadBufferSize != DefaultBufferSize || cfg.WriteBufferSize != DefaultBufferSize {
		t.Errorf("默认缓冲区大小不正确：%d / %d", cfg.ReadBufferSize, cfg.WriteBufferSize)
	}

	cfg.SetThreadPoolSize(8).SetReadTimeout(100).SetWriteTimeout(200).
		SetReadBufferSize(1024).SetWriteBufferSize(2048)
	if cfg.ThreadPoolSize != 8 || cfg.ReadTimeout != 100 || cfg.WriteTimeout != 200 ||
		cfg.ReadBufferSize != 1024 || cfg.WriteBufferSize != 2048 {
		t.Errorf("链式 setter 未生效: %+v", cfg)
	}
}

func TestSocketUtilConnect(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, ln.Close)

	go func() {
		c, _ := ln.Accept()
		if c != nil {
			_ = c.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	conn, err := Connect("127.0.0.1", addr.Port, time.Second)
	if err != nil {
		t.Fatalf("Connect 失败: %v", err)
	}
	defer closeAndReport(t, conn.Close)

	if !IsConnected(conn) {
		t.Errorf("IsConnected 应返回 true")
	}
	if GetRemoteAddress(conn) == nil {
		t.Errorf("GetRemoteAddress 不应返回 nil")
	}
}

func TestSocketRuntimeError(t *testing.T) {
	base := net.ErrClosed
	e := WrapSocketError(base, "wrapped")
	if e == nil {
		t.Fatal("WrapSocketError 不应返回 nil")
	}
	if e.Unwrap() != base {
		t.Errorf("Unwrap 失败")
	}
	if e.Error() == "" {
		t.Errorf("Error 不应为空")
	}
	if WrapSocketError(nil, "x") != nil {
		t.Errorf("nil err 应返回 nil")
	}
	if NewSocketErrorf("hello %s", "world").Error() != "hello world" {
		t.Errorf("格式化失败")
	}
}
