package vid

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"
)

func TestObjectIDFacade(t *testing.T) {
	if oid := ObjectId(); len(oid) != 24 {
		t.Fatalf("ObjectId failed: %q", oid)
	}
}

func TestObjectIDFacadeOptions(t *testing.T) {
	obj := ObjectIdWithOptions(
		WithObjectIDTimeFunc(func() time.Time { return time.Unix(1, 0) }),
		WithObjectIDRandomReader(bytes.NewReader([]byte{1, 2, 3, 4, 5})),
		WithObjectIDCounter(func() uint32 { return 0xabcdef }),
	)
	if obj != "000000010102030405abcdef" {
		t.Fatalf("ObjectIdWithOptions = %s", obj)
	}
	if _, err := hex.DecodeString(obj); err != nil {
		t.Fatalf("ObjectIdWithOptions is not hex: %v", err)
	}
}
