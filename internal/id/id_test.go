package id

import (
	"strings"
	"testing"
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
