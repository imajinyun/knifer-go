package id

import (
	"sync"
	"testing"
)

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

func TestSnowflakeRuntimeOptionsBypassSingletonCache(t *testing.T) {
	t.Cleanup(func() { ConfigureDefaultSnowflake() })
	configured := ConfigureDefaultSnowflake(WithSnowflakeWorkerID(1), WithSnowflakeDatacenterID(1))

	now := int64(1288834974657)
	one := GetSnowflakeWithOptions(
		WithSnowflakeWorkerID(2),
		WithSnowflakeDatacenterID(3),
		WithSnowflakeTimeFunc(func() int64 { return now }),
	)
	two := GetSnowflakeWithOptions(
		WithSnowflakeWorkerID(2),
		WithSnowflakeDatacenterID(3),
		WithSnowflakeTimeFunc(func() int64 { return now }),
	)
	if one == configured || two == configured || one == two {
		t.Fatalf("runtime options should bypass default singleton/cache: configured=%p one=%p two=%p", configured, one, two)
	}
	if one.WorkerID() != 2 || one.DatacenterID() != 3 {
		t.Fatalf("runtime options ids = worker %d datacenter %d", one.WorkerID(), one.DatacenterID())
	}
}

func TestSnowflakeCacheOptionRetainsSingletonBehavior(t *testing.T) {
	now := int64(1288834974657)
	one := GetSnowflakeWithWorkerDataCenterWithOptions(9, 10,
		WithSnowflakeTimeFunc(func() int64 { return now }),
		WithSnowflakeCache(true),
	)
	two := GetSnowflakeWithWorkerDataCenterWithOptions(9, 10,
		WithSnowflakeTimeFunc(func() int64 { return now }),
		WithSnowflakeCache(true),
	)
	if one != two {
		t.Fatal("explicit cache option should retain singleton behavior")
	}

	isolated := GetSnowflakeWithWorkerDataCenterWithOptions(9, 10, WithSnowflakeCache(false))
	if isolated == one {
		t.Fatal("WithSnowflakeCache(false) should bypass cached worker/datacenter generator")
	}
}

func TestNewIsolatedSnowflake(t *testing.T) {
	t.Cleanup(func() { ConfigureDefaultSnowflake() })
	configured := ConfigureDefaultSnowflake(WithSnowflakeWorkerID(1), WithSnowflakeDatacenterID(1))
	isolated := NewIsolatedSnowflake(WithSnowflakeWorkerID(4), WithSnowflakeDatacenterID(5))
	if isolated == configured || isolated.WorkerID() != 4 || isolated.DatacenterID() != 5 {
		t.Fatalf("isolated snowflake = %p worker %d datacenter %d", isolated, isolated.WorkerID(), isolated.DatacenterID())
	}
}

func TestSnowflakeGlobalStateConcurrentConfigureAndUse(t *testing.T) {
	t.Cleanup(func() { ConfigureDefaultSnowflake() })

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		workerID := int64(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ConfigureDefaultSnowflake(WithSnowflakeWorkerID(workerID), WithSnowflakeDatacenterID(workerID))
				if id := GetSnowflakeNextID(); id <= 0 {
					t.Errorf("GetSnowflakeNextID() = %d, want positive", id)
				}
			}
		}()
	}
	for i := 0; i < 8; i++ {
		workerID := int64(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if id := GetSnowflakeWithWorkerDataCenter(workerID, workerID).NextID(); id <= 0 {
					t.Errorf("cached snowflake id = %d, want positive", id)
				}
				if id := NewIsolatedSnowflake(WithSnowflakeWorkerID(workerID), WithSnowflakeDatacenterID(workerID)).NextID(); id <= 0 {
					t.Errorf("isolated snowflake id = %d, want positive", id)
				}
			}
		}()
	}
	wg.Wait()
}
