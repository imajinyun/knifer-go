package vid_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/imajinyun/knifer-go/vid"
)

func ExampleSimpleUUID() {
	// A simple UUID is a 32-character hex string without dashes.
	id := vid.SimpleUUID()
	fmt.Println(len(id))
	// Output: 32
}

func ExampleRandomUUID() {
	// A standard UUID is 36 characters including dashes.
	id := vid.RandomUUID()
	fmt.Println(len(id))
	// Output: 36
}

func ExampleNanoIdN() {
	id := vid.NanoIdN(10)
	fmt.Println(len(id))
	// Output: 10
}

func ExampleGetSnowflakeNextID() {
	// Snowflake IDs are monotonically increasing positive int64 values.
	id := vid.GetSnowflakeNextID()
	fmt.Println(id > 0)
	// Output: true
}

func ExampleObjectId() {
	id := vid.ObjectId()
	fmt.Println(len(id))
	// Output: 24
}

func ExampleRandomUUIDWithOptions() {
	id := vid.RandomUUIDWithOptions(vid.WithRandomReader(bytes.NewReader(make([]byte, 16))))
	fmt.Println(id)
	// Output: 00000000-0000-4000-8000-000000000000
}

func ExampleObjectIdWithOptions() {
	id := vid.ObjectIdWithOptions(
		vid.WithObjectIDTimeFunc(func() time.Time { return time.Unix(1, 0) }),
		vid.WithObjectIDRandomReader(bytes.NewReader([]byte{1, 2, 3, 4, 5})),
		vid.WithObjectIDCounter(func() uint32 { return 2 }),
	)
	fmt.Println(id)
	// Output: 000000010102030405000002
}

func ExampleNanoIdWithOptions() {
	id := vid.NanoIdWithOptions(
		vid.WithNanoIDAlphabet("ab"),
		vid.WithNanoIDLength(4),
		vid.WithNanoIDRandomReader(bytes.NewReader([]byte{0, 1, 0, 1})),
	)
	fmt.Println(id)
	// Output: abab
}

func ExampleCreateSnowflakeWithOptions() {
	sf := vid.CreateSnowflakeWithOptions(
		vid.WithSnowflakeWorkerID(1),
		vid.WithSnowflakeDatacenterID(2),
		vid.WithSnowflakeTimeFunc(func() int64 { return 1288834974658 }),
	)

	fmt.Println(sf.NextID() > 0)
	// Output: true
}

func ExampleCreateSnowflake() {
	sf := vid.CreateSnowflake(1, 2)
	fmt.Println(sf.WorkerID(), sf.DatacenterID(), sf.NextID() > 0)
	// Output: 1 2 true
}

func ExampleConfigureDefaultSnowflake() {
	sf := vid.ConfigureDefaultSnowflake(
		vid.WithSnowflakeWorkerID(5),
		vid.WithSnowflakeDatacenterID(6),
		vid.WithSnowflakeTimeFunc(func() int64 { return 1288834974658 }),
	)
	defer vid.ConfigureDefaultSnowflake()

	fmt.Println(sf.WorkerID(), sf.DatacenterID(), vid.GetSnowflakeNextID() > 0)
	// Output: 5 6 true
}

func ExampleNewIsolatedSnowflake() {
	sf := vid.NewIsolatedSnowflake(
		vid.WithSnowflakeWorkerID(4),
		vid.WithSnowflakeDatacenterID(5),
		vid.WithSnowflakeTimeFunc(func() int64 { return 1288834974658 }),
	)
	fmt.Println(sf.WorkerID(), sf.DatacenterID(), sf.NextID() > 0)
	// Output: 4 5 true
}

func ExampleGetSnowflakeNextIDWithOptions() {
	id := vid.GetSnowflakeNextIDWithOptions(
		vid.WithSnowflakeWorkerID(1),
		vid.WithSnowflakeDatacenterID(2),
		vid.WithSnowflakeTimeFunc(func() int64 { return 1288834974658 }),
		vid.WithSnowflakeCache(false),
	)
	fmt.Println(id > 0)
	// Output: true
}

func ExampleGetSnowflakeNextIDStr() {
	id := vid.GetSnowflakeNextIDStr()
	fmt.Println(id != "")
	// Output: true
}

func ExampleNanoId() {
	fmt.Println(len(vid.NanoId()))
	// Output: 21
}

func ExampleFastSimpleUUID() {
	fmt.Println(len(vid.FastSimpleUUID()))
	// Output: 32
}

func ExampleSimpleUUIDWithOptions() {
	id := vid.SimpleUUIDWithOptions(vid.WithRandomReader(bytes.NewReader(make([]byte, 16))))
	fmt.Println(id)
	// Output: 00000000000040008000000000000000
}

func ExampleConfigureDefaultFallbackRandomSourceProvider() {
	vid.ResetDefaultFallbackRandomSource()
	defer vid.ResetDefaultFallbackRandomSource()

	vid.ConfigureDefaultFallbackRandomSourceProvider(func() *rand.Rand {
		return rand.New(rand.NewSource(13))
	})
	id := vid.SimpleUUIDWithOptions(vid.WithRandomReader(failingReader{}))
	fmt.Println(len(id))
	// Output: 32
}

func ExampleResetDefaultFallbackRandomSource() {
	vid.ResetDefaultFallbackRandomSource()
	fmt.Println("reset")
	// Output: reset
}

func ExampleNanoIdNWithOptions() {
	id := vid.NanoIdNWithOptions(4,
		vid.WithNanoIDAlphabet("ab"),
		vid.WithNanoIDRandomReader(bytes.NewReader([]byte{0, 1, 0, 1})),
	)
	fmt.Println(id)
	// Output: abab
}

func ExampleGetWorkerIDWithOptions() {
	id := vid.GetWorkerIDWithOptions(1, 31, vid.WithSnowflakePIDFunc(func() int { return 8 }))
	fmt.Println(id >= 0 && id <= 31)
	// Output: true
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
