package vid

import (
	"strings"
	"testing"
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
