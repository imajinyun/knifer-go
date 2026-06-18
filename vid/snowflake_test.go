package vid

import (
	"net"
	"testing"
)

func TestSnowflakeFacade(t *testing.T) {
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

func TestDefaultSnowflakeFacadeOptions(t *testing.T) {
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
		t.Fatal("default snowflake singleton facade should keep configured instance")
	}
	first := sf.NextID()
	second := GetSnowflakeNextID()
	if first <= 0 || second <= first {
		t.Fatalf("configured default snowflake facade should generate increasing ids: %d %d", first, second)
	}
}

func TestSnowflakeFacadeRuntimeOptionsBypassCache(t *testing.T) {
	t.Cleanup(func() { ConfigureDefaultSnowflake() })
	configured := ConfigureDefaultSnowflake(WithSnowflakeWorkerID(1), WithSnowflakeDatacenterID(1))
	now := int64(1288834974657)
	one := GetSnowflakeWithOptions(WithSnowflakeWorkerID(2), WithSnowflakeDatacenterID(3), WithSnowflakeTimeFunc(func() int64 { return now }))
	two := GetSnowflakeWithOptions(WithSnowflakeWorkerID(2), WithSnowflakeDatacenterID(3), WithSnowflakeTimeFunc(func() int64 { return now }))
	if one == configured || two == configured || one == two {
		t.Fatal("facade runtime snowflake options should bypass singleton/cache")
	}

	isolated := NewIsolatedSnowflake(WithSnowflakeWorkerID(4), WithSnowflakeDatacenterID(5))
	if isolated.WorkerID() != 4 || isolated.DatacenterID() != 5 {
		t.Fatalf("isolated snowflake ids: worker=%d datacenter=%d", isolated.WorkerID(), isolated.DatacenterID())
	}
}

func TestSnowflakeFacadeOptionSetters(t *testing.T) {
	if WithSnowflakeWaitFunc(func(last int64, now func() int64) int64 { return now() }) == nil {
		t.Fatal("WithSnowflakeWaitFunc returned nil")
	}
	if WithSnowflakeCache(false) == nil {
		t.Fatal("WithSnowflakeCache returned nil")
	}
	if WithSnowflakeInterfacesFunc(func() ([]net.Interface, error) { return nil, nil }) == nil {
		t.Fatal("WithSnowflakeInterfacesFunc returned nil")
	}
	if WithSnowflakePIDFunc(func() int { return 99 }) == nil {
		t.Fatal("WithSnowflakePIDFunc returned nil")
	}
}
