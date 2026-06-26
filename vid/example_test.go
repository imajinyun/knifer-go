package vid_test

import (
	"bytes"
	"fmt"
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
