package id

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestSimpleUUID(t *testing.T) {
	u1 := SimpleUUID()
	u2 := SimpleUUID()
	if len(u1) != 32 || len(u2) != 32 {
		t.Fatalf("UUID length wrong")
	}
	if u1 == u2 {
		t.Fatalf("UUID collision")
	}
	// Version 4 marker: the 13th character is '4'.
	if u1[12] != '4' {
		t.Fatalf("UUID version: %s", u1)
	}
}

func TestRandomUUIDAndFastSimpleUUID(t *testing.T) {
	u := RandomUUID()
	if len(u) != 36 || strings.Count(u, "-") != 4 {
		t.Fatalf("RandomUUID format: %s", u)
	}
	s := FastSimpleUUID()
	if len(s) != 32 || strings.Contains(s, "-") || s[12] != '4' {
		t.Fatalf("FastSimpleUUID format: %s", s)
	}
}

func TestFastUUID(t *testing.T) {
	u := FastUUID()
	if len(u) != 36 || strings.Count(u, "-") != 4 {
		t.Fatalf("FastUUID format: %s", u)
	}
}

func TestObjectId(t *testing.T) {
	o := ObjectId()
	if len(o) != 24 {
		t.Fatalf("ObjectId length: %s", o)
	}
}

func TestNanoId(t *testing.T) {
	id := NanoId()
	if len(id) != 21 {
		t.Fatalf("NanoId default len: %s", id)
	}
	id = NanoIdN(10)
	if len(id) != 10 {
		t.Fatalf("NanoIdN len: %s", id)
	}
}

func TestSnowflake(t *testing.T) {
	sf := CreateSnowflake(1, 2)
	if sf.WorkerID() != 1 || sf.DatacenterID() != 2 {
		t.Fatalf("snowflake ids: worker=%d datacenter=%d", sf.WorkerID(), sf.DatacenterID())
	}
	id1 := sf.NextID()
	id2 := sf.NextID()
	if id1 <= 0 || id2 <= id1 {
		t.Fatalf("snowflake should be positive and increasing: %d %d", id1, id2)
	}
	if sf.NextIDStr() == "" {
		t.Fatal("snowflake string id should not be empty")
	}
	first := GetSnowflakeWithWorkerDataCenter(1, 2)
	second := GetSnowflakeWithWorkerDataCenter(1, 2)
	if first != second {
		t.Fatal("same worker/datacenter pair should return singleton")
	}
	if GetSnowflakeWithWorker(3) == nil || GetSnowflake() == nil {
		t.Fatal("snowflake singleton helpers should not return nil")
	}
	if GetSnowflakeNextID() <= 0 || GetSnowflakeNextIDStr() == "" {
		t.Fatal("default snowflake next id helpers failed")
	}
}

func TestSnowflakeOptions(t *testing.T) {
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

func TestDefaultSnowflakeOptions(t *testing.T) {
	t.Cleanup(func() { ConfigureDefaultSnowflake() })
	now := int64(1288834974657)
	sf := ConfigureDefaultSnowflake(
		WithSnowflakeWorkerID(5),
		WithSnowflakeDatacenterID(6),
		WithSnowflakeTimeFunc(func() int64 { return now }),
	)
	if sf.WorkerID() != 5 || sf.DatacenterID() != 6 {
		t.Fatalf("default snowflake option ids: worker=%d datacenter=%d", sf.WorkerID(), sf.DatacenterID())
	}
	if GetSnowflake() != sf || GetSnowflakeWithOptions(WithSnowflakeWorkerID(7)) != sf {
		t.Fatal("default snowflake singleton should keep configured instance")
	}
	first := sf.NextID()
	second := GetSnowflakeNextID()
	if first <= 0 || second <= first {
		t.Fatalf("configured default snowflake should generate increasing ids: %d %d", first, second)
	}
	if got := GetSnowflakeNextIDStr(); got == "" {
		t.Fatal("configured default snowflake string id should not be empty")
	}
}

func TestNormalizeSnowflakeIDUsesProvidedMax(t *testing.T) {
	if got := normalizeSnowflakeID(14, 5); got != 2 {
		t.Fatalf("normalizeSnowflakeID should use provided max: got %d", got)
	}
	if got := normalizeSnowflakeID(-14, 5); got != 2 {
		t.Fatalf("normalizeSnowflakeID should normalize negative values with provided max: got %d", got)
	}
	if got := normalizeSnowflakeID(14, 0); got != 0 {
		t.Fatalf("normalizeSnowflakeID should return 0 when max is not positive: got %d", got)
	}
}

func TestWorkerAndDatacenterID(t *testing.T) {
	if dc := GetDataCenterID(31); dc < 0 || dc > 31 {
		t.Fatalf("datacenter id out of range: %d", dc)
	}
	if worker := GetWorkerID(1, 31); worker < 0 || worker > 31 {
		t.Fatalf("worker id out of range: %d", worker)
	}
}

func TestIDOptions(t *testing.T) {
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
