package vskt_test

import (
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/vskt"
)

func TestFacadeSocketError(t *testing.T) {
	err := vskt.NewSocketErrorMsg("test error")
	if err == nil {
		t.Fatal("expected non-nil socket error")
	}
	if err.Error() != "test error" {
		t.Fatalf("expected 'test error', got %q", err.Error())
	}
}

func TestFacadeOperations(t *testing.T) {
	// verify operation constants are accessible
	_ = vskt.OpRead
	_ = vskt.OpWrite
	_ = vskt.OpConnect
	_ = vskt.OpAccept
}

func TestFacadeSocketErrorConstructors(t *testing.T) {
	cause := errors.New("dial failed")
	if got := vskt.NewSocketError(cause); got == nil || got.Cause != cause || got.Msg != "dial failed" {
		t.Fatalf("NewSocketError() = %#v", got)
	}
	if got := vskt.NewSocketError(nil); got != nil {
		t.Fatalf("NewSocketError(nil) = %#v, want nil", got)
	}
	if got := vskt.NewSocketErrorf("socket %s", "closed"); got == nil || got.Error() != "socket closed" {
		t.Fatalf("NewSocketErrorf() = %#v", got)
	}
	if got := vskt.WrapSocketError(cause, "connect"); got == nil || !errors.Is(got, cause) || got.Error() != "connect: dial failed" {
		t.Fatalf("WrapSocketError() = %#v", got)
	}
	if got := vskt.WrapSocketError(nil, "ignored"); got != nil {
		t.Fatalf("WrapSocketError(nil) = %#v, want nil", got)
	}
}
