package vid_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vid"
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
