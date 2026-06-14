package vid

import (
	"bytes"
	"strings"
	"testing"
)

func TestUUIDFacade(t *testing.T) {
	u1 := SimpleUUID()
	u2 := UUID()
	if len(u1) != 32 || len(u2) != 32 || u1 == u2 || u1[12] != '4' {
		t.Fatalf("uuid failed: %q %q", u1, u2)
	}
	if u := RandomUUID(); len(u) != 36 || strings.Count(u, "-") != 4 {
		t.Fatalf("RandomUUID failed: %q", u)
	}
	if fast := FastUUID(); len(fast) != 36 || strings.Count(fast, "-") != 4 {
		t.Fatalf("FastUUID failed: %q", fast)
	}
	if u := FastSimpleUUID(); len(u) != 32 || strings.Contains(u, "-") {
		t.Fatalf("FastSimpleUUID failed: %q", u)
	}
}

func TestUUIDFacadeOptions(t *testing.T) {
	reader := bytes.NewReader(bytes.Repeat([]byte{0x11}, 32))
	u := SimpleUUIDWithOptions(WithRandomReader(reader))
	if len(u) != 32 || u[12] != '4' || u[16] != '9' {
		t.Fatalf("SimpleUUIDWithOptions format: %s", u)
	}
}
