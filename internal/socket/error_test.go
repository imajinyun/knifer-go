package socket

import (
	"errors"
	"net"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

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
	if e.ErrorCode() != knifer.ErrCodeInternal {
		t.Errorf("ErrorCode() = %v, want %v", e.ErrorCode(), knifer.ErrCodeInternal)
	}
	if !errors.Is(e, knifer.ErrCodeInternal) {
		t.Errorf("SocketRuntimeError 应匹配 ErrCodeInternal")
	}
	if !errors.Is(e, base) {
		t.Errorf("SocketRuntimeError 应保留 cause 链")
	}
	if WrapSocketError(nil, "x") != nil {
		t.Errorf("nil err 应返回 nil")
	}
	if NewSocketErrorf("hello %s", "world").Error() != "hello world" {
		t.Errorf("格式化失败")
	}
}

func TestSocketRuntimeErrorEmptyMsg(t *testing.T) {
	e := &SocketRuntimeError{Code: knifer.ErrCodeInternal, Cause: net.ErrClosed}
	if e.Error() != net.ErrClosed.Error() {
		t.Fatalf("Error with empty msg = %q, want %q", e.Error(), net.ErrClosed.Error())
	}
}

func TestSocketRuntimeErrorNilReceiver(t *testing.T) {
	var e *SocketRuntimeError
	if e.Error() != "" {
		t.Fatal("nil receiver Error should be empty")
	}
	if e.ErrorCode() != "" {
		t.Fatal("nil receiver ErrorCode should be empty")
	}
	if e.Unwrap() != nil {
		t.Fatal("nil receiver Unwrap should return nil")
	}
}

func TestNewSocketError(t *testing.T) {
	if e := NewSocketError(nil); e != nil {
		t.Fatal("NewSocketError(nil) should return nil")
	}
	if e := NewSocketError(net.ErrClosed); e == nil || e.Error() == "" {
		t.Fatal("NewSocketError with error should return non-nil error")
	}
}

func TestNewSocketErrorMsg(t *testing.T) {
	e := NewSocketErrorMsg("test error")
	if e.Error() != "test error" {
		t.Fatalf("NewSocketErrorMsg = %q, want %q", e.Error(), "test error")
	}
	if e.ErrorCode() != knifer.ErrCodeInternal {
		t.Fatalf("ErrorCode = %v, want %v", e.ErrorCode(), knifer.ErrCodeInternal)
	}
}

func TestSocketRuntimeErrorIsNilTarget(t *testing.T) {
	e := NewSocketErrorMsg("err")
	if errors.Is(e, nil) {
		t.Fatal("SocketRuntimeError should not match nil target")
	}
}
