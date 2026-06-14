package vskt_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vskt"
)

func TestFacadeSocketIsConnected(t *testing.T) {
	// nil conn should not be connected
	if vskt.SocketIsConnected(nil) {
		t.Fatal("expected nil conn to be disconnected")
	}
}
