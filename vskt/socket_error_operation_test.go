package vskt_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vskt"
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
