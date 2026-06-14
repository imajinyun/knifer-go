package vid

import (
	"bytes"
	"testing"
)

func TestNanoIDFacade(t *testing.T) {
	if nid := NanoId(); len(nid) != 21 {
		t.Fatalf("NanoId failed: %q", nid)
	}
	if nid := NanoIdN(10); len(nid) != 10 {
		t.Fatalf("NanoIdN failed: %q", nid)
	}
}

func TestNanoIDFacadeOptions(t *testing.T) {
	nid := NanoIdWithOptions(
		WithNanoIDLength(5),
		WithNanoIDAlphabet("ab"),
		WithNanoIDRandomReader(bytes.NewReader([]byte{0, 1, 0, 1, 1})),
	)
	if nid != "ababb" {
		t.Fatalf("NanoIdWithOptions = %q", nid)
	}
}
