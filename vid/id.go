package vid

import (
	"io"
	"time"

	idimpl "github.com/imajinyun/go-knifer/internal/id"
)

type (
	// Snowflake generates distributed unique IDs.
	Snowflake = idimpl.Snowflake
	// RandomOption customizes UUID random sources.
	RandomOption = idimpl.RandomOption
	// ObjectIDOption customizes ObjectId generation.
	ObjectIDOption = idimpl.ObjectIDOption
	// NanoIDOption customizes NanoId generation.
	NanoIDOption = idimpl.NanoIDOption
	// SnowflakeOption customizes Snowflake construction.
	SnowflakeOption = idimpl.SnowflakeOption
)

func RandomUUID() string     { return idimpl.RandomUUID() }
func SimpleUUID() string     { return idimpl.SimpleUUID() }
func FastUUID() string       { return idimpl.FastUUID() }
func FastSimpleUUID() string { return idimpl.FastSimpleUUID() }
func UUID() string           { return idimpl.SimpleUUID() }
func ObjectId() string       { return idimpl.ObjectId() }

// WithRandomReader sets the entropy source used by UUID helpers.
func WithRandomReader(reader io.Reader) RandomOption { return idimpl.WithRandomReader(reader) }

// RandomUUIDWithOptions creates an RFC 4122 UUID with random options.
func RandomUUIDWithOptions(opts ...RandomOption) string { return idimpl.RandomUUIDWithOptions(opts...) }

// SimpleUUIDWithOptions creates a UUID without hyphens with random options.
func SimpleUUIDWithOptions(opts ...RandomOption) string { return idimpl.SimpleUUIDWithOptions(opts...) }

// WithObjectIDRandomReader sets the random source used by ObjectIdWithOptions.
func WithObjectIDRandomReader(reader io.Reader) ObjectIDOption {
	return idimpl.WithObjectIDRandomReader(reader)
}

// WithObjectIDTimeFunc sets the timestamp source used by ObjectIdWithOptions.
func WithObjectIDTimeFunc(now func() time.Time) ObjectIDOption {
	return idimpl.WithObjectIDTimeFunc(now)
}

// WithObjectIDCounter sets the counter source used by ObjectIdWithOptions.
func WithObjectIDCounter(counter func() uint32) ObjectIDOption {
	return idimpl.WithObjectIDCounter(counter)
}

// ObjectIdWithOptions creates an ObjectId with deterministic/custom generation options.
func ObjectIdWithOptions(opts ...ObjectIDOption) string { return idimpl.ObjectIdWithOptions(opts...) }

func CreateSnowflake(workerID, datacenterID int64) *Snowflake {
	return idimpl.CreateSnowflake(workerID, datacenterID)
}

// WithSnowflakeWorkerID sets the Snowflake worker ID.
func WithSnowflakeWorkerID(workerID int64) SnowflakeOption {
	return idimpl.WithSnowflakeWorkerID(workerID)
}

// WithSnowflakeDatacenterID sets the Snowflake datacenter ID.
func WithSnowflakeDatacenterID(datacenterID int64) SnowflakeOption {
	return idimpl.WithSnowflakeDatacenterID(datacenterID)
}

// WithSnowflakeTimeFunc sets the millisecond time source used by Snowflake.
func WithSnowflakeTimeFunc(timeFunc func() int64) SnowflakeOption {
	return idimpl.WithSnowflakeTimeFunc(timeFunc)
}

// WithSnowflakeWaitFunc sets the wait strategy used when Snowflake sequence overflows in one millisecond.
func WithSnowflakeWaitFunc(waitFunc func(lastTimestamp int64, now func() int64) int64) SnowflakeOption {
	return idimpl.WithSnowflakeWaitFunc(waitFunc)
}

// CreateSnowflakeWithOptions creates a Snowflake generator from options.
func CreateSnowflakeWithOptions(opts ...SnowflakeOption) *Snowflake {
	return idimpl.CreateSnowflakeWithOptions(opts...)
}

func GetSnowflake() *Snowflake { return idimpl.GetSnowflake() }

// GetSnowflakeWithOptions returns the default singleton Snowflake generator, creating it with options if needed.
func GetSnowflakeWithOptions(opts ...SnowflakeOption) *Snowflake {
	return idimpl.GetSnowflakeWithOptions(opts...)
}

// ConfigureDefaultSnowflake replaces the default singleton Snowflake generator with options.
func ConfigureDefaultSnowflake(opts ...SnowflakeOption) *Snowflake {
	return idimpl.ConfigureDefaultSnowflake(opts...)
}

func GetSnowflakeWithWorker(workerID int64) *Snowflake {
	return idimpl.GetSnowflakeWithWorker(workerID)
}

func GetSnowflakeWithWorkerDataCenter(workerID, datacenterID int64) *Snowflake {
	return idimpl.GetSnowflakeWithWorkerDataCenter(workerID, datacenterID)
}

func GetDataCenterID(maxDatacenterID int64) int64 { return idimpl.GetDataCenterID(maxDatacenterID) }
func GetWorkerID(datacenterID, maxWorkerID int64) int64 {
	return idimpl.GetWorkerID(datacenterID, maxWorkerID)
}

func NanoId() string       { return idimpl.NanoId() }
func NanoIdN(n int) string { return idimpl.NanoIdN(n) }

// WithNanoIDRandomReader sets the entropy source used by NanoIdWithOptions.
func WithNanoIDRandomReader(reader io.Reader) NanoIDOption {
	return idimpl.WithNanoIDRandomReader(reader)
}

// WithNanoIDAlphabet sets the alphabet used by NanoIdWithOptions.
func WithNanoIDAlphabet(alphabet string) NanoIDOption { return idimpl.WithNanoIDAlphabet(alphabet) }

// WithNanoIDLength sets the output length used by NanoIdWithOptions.
func WithNanoIDLength(length int) NanoIDOption { return idimpl.WithNanoIDLength(length) }

// NanoIdWithOptions creates a NanoId with custom generation options.
func NanoIdWithOptions(opts ...NanoIDOption) string { return idimpl.NanoIdWithOptions(opts...) }

func GetSnowflakeNextID() int64     { return idimpl.GetSnowflakeNextID() }
func GetSnowflakeNextIDStr() string { return idimpl.GetSnowflakeNextIDStr() }
