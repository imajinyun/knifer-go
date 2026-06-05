package vid

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestIDFacade(t *testing.T) {
	u1 := SimpleUUID()
	u2 := UUID()
	if len(u1) != 32 || len(u2) != 32 || u1 == u2 || u1[12] != '4' {
		t.Fatalf("uuid failed: %q %q", u1, u2)
	}
	if fast := FastUUID(); len(fast) != 36 || strings.Count(fast, "-") != 4 {
		t.Fatalf("FastUUID failed: %q", fast)
	}
	if oid := ObjectId(); len(oid) != 24 {
		t.Fatalf("ObjectId failed: %q", oid)
	}
	if nid := NanoId(); len(nid) != 21 {
		t.Fatalf("NanoId failed: %q", nid)
	}
	if nid := NanoIdN(10); len(nid) != 10 {
		t.Fatalf("NanoIdN failed: %q", nid)
	}
}

func TestIDFacadeExtended(t *testing.T) {
	if u := RandomUUID(); len(u) != 36 || strings.Count(u, "-") != 4 {
		t.Fatalf("RandomUUID failed: %q", u)
	}
	if u := FastSimpleUUID(); len(u) != 32 || strings.Contains(u, "-") {
		t.Fatalf("FastSimpleUUID failed: %q", u)
	}
	sf := CreateSnowflake(1, 2)
	if sf.WorkerID() != 1 || sf.DatacenterID() != 2 || sf.NextID() <= 0 || sf.NextIDStr() == "" {
		t.Fatal("snowflake facade failed")
	}
	if GetSnowflake() == nil || GetSnowflakeWithWorker(1) == nil || GetSnowflakeWithWorkerDataCenter(1, 2) == nil {
		t.Fatal("snowflake singleton facade failed")
	}
	if dc := GetDataCenterID(31); dc < 0 || dc > 31 {
		t.Fatalf("datacenter id out of range: %d", dc)
	}
	if worker := GetWorkerID(1, 31); worker < 0 || worker > 31 {
		t.Fatalf("worker id out of range: %d", worker)
	}
	if GetSnowflakeNextID() <= 0 || GetSnowflakeNextIDStr() == "" {
		t.Fatal("snowflake next id facade failed")
	}
}

func TestSnowflakeFacadeOptions(t *testing.T) {
	now := int64(1288834974657)
	sf := CreateSnowflakeWithOptions(
		WithSnowflakeWorkerID(3),
		WithSnowflakeDatacenterID(4),
		WithSnowflakeTimeFunc(func() int64 { return now }),
	)
	if sf.WorkerID() != 3 || sf.DatacenterID() != 4 {
		t.Fatalf("snowflake option ids: worker=%d datacenter=%d", sf.WorkerID(), sf.DatacenterID())
	}
	id1 := sf.NextID()
	id2 := sf.NextID()
	if id1 <= 0 || id2 <= id1 {
		t.Fatalf("snowflake option IDs should be positive and increasing: %d %d", id1, id2)
	}
}

func TestIDFacadeOptions(t *testing.T) {
	reader := bytes.NewReader(bytes.Repeat([]byte{0x11}, 32))
	u := SimpleUUIDWithOptions(WithRandomReader(reader))
	if len(u) != 32 || u[12] != '4' || u[16] != '9' {
		t.Fatalf("SimpleUUIDWithOptions format: %s", u)
	}

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

	nid := NanoIdWithOptions(
		WithNanoIDLength(5),
		WithNanoIDAlphabet("ab"),
		WithNanoIDRandomReader(bytes.NewReader([]byte{0, 1, 0, 1, 1})),
	)
	if nid != "ababb" {
		t.Fatalf("NanoIdWithOptions = %q", nid)
	}
}
