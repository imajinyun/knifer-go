package vid

import (
	"io"
	mathrand "math/rand"
	"net"
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

// RandomUUID creates an RFC 4122 UUID using the default entropy source.
func RandomUUID() string { return RandomUUIDWithOptions() }

// SimpleUUID creates a UUID string without hyphens using the default entropy source.
func SimpleUUID() string { return SimpleUUIDWithOptions() }

// FastUUID creates an RFC 4122 UUID through the compatibility fast UUID alias.
func FastUUID() string { return FastUUIDWithOptions() }

// FastSimpleUUID creates a hyphen-free UUID through the compatibility fast UUID alias.
func FastSimpleUUID() string { return FastSimpleUUIDWithOptions() }

// UUID creates a hyphen-free UUID string for compatibility with legacy callers.
func UUID() string { return SimpleUUIDWithOptions() }

// ObjectId creates a MongoDB-style ObjectId using the default timestamp, random bytes, and counter sources.
func ObjectId() string { return ObjectIdWithOptions() }

// WithRandomReader sets the entropy source used by UUID helpers.
func WithRandomReader(reader io.Reader) RandomOption { return idimpl.WithRandomReader(reader) }

// WithFallbackRandomSource sets the fallback PRNG used when UUID random reads
// fail. It is intended for compatibility and deterministic tests, not for
// security-sensitive identifiers.
func WithFallbackRandomSource(source *mathrand.Rand) RandomOption {
	return idimpl.WithFallbackRandomSource(source)
}

// ConfigureDefaultFallbackRandomSourceProvider sets the provider used to lazily create the package-level fallback PRNG.
func ConfigureDefaultFallbackRandomSourceProvider(provider func() *mathrand.Rand) {
	idimpl.ConfigureDefaultFallbackRandomSourceProvider(provider)
}

// ResetDefaultFallbackRandomSource restores the fallback PRNG provider and clears cached state.
func ResetDefaultFallbackRandomSource() { idimpl.ResetDefaultFallbackRandomSource() }

// SetFallbackRandomSeed resets the package-level fallback PRNG to a deterministic seed.
// It is intended for tests and reproducible non-security fallback behavior only.
func SetFallbackRandomSeed(seed int64) { idimpl.SetFallbackRandomSeed(seed) }

// RandomUUIDWithOptions creates an RFC 4122 UUID with random options.
func RandomUUIDWithOptions(opts ...RandomOption) string { return idimpl.RandomUUIDWithOptions(opts...) }

// SimpleUUIDWithOptions creates a UUID without hyphens with random options.
func SimpleUUIDWithOptions(opts ...RandomOption) string { return idimpl.SimpleUUIDWithOptions(opts...) }

// FastUUIDWithOptions creates a UUID alias with random options.
func FastUUIDWithOptions(opts ...RandomOption) string { return idimpl.FastUUIDWithOptions(opts...) }

// FastSimpleUUIDWithOptions creates a simple UUID alias with random options.
func FastSimpleUUIDWithOptions(opts ...RandomOption) string {
	return idimpl.FastSimpleUUIDWithOptions(opts...)
}

// WithObjectIDRandomReader sets the random source used by ObjectIdWithOptions.
func WithObjectIDRandomReader(reader io.Reader) ObjectIDOption {
	return idimpl.WithObjectIDRandomReader(reader)
}

// WithObjectIDFallbackRandomSource sets the fallback PRNG used when ObjectId
// random reads fail. It is intended for compatibility and deterministic tests,
// not for security-sensitive identifiers.
func WithObjectIDFallbackRandomSource(source *mathrand.Rand) ObjectIDOption {
	return idimpl.WithObjectIDFallbackRandomSource(source)
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

// CreateSnowflake creates a Snowflake generator for an explicit worker and datacenter pair.
func CreateSnowflake(workerID, datacenterID int64) *Snowflake {
	return CreateSnowflakeWithOptions(WithSnowflakeWorkerID(workerID), WithSnowflakeDatacenterID(datacenterID))
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

// WithSnowflakeCache controls whether singleton helper variants may reuse cached generators.
func WithSnowflakeCache(enabled bool) SnowflakeOption { return idimpl.WithSnowflakeCache(enabled) }

// WithSnowflakeInterfacesFunc sets the network interface provider used to derive the default datacenter ID.
func WithSnowflakeInterfacesFunc(interfaces func() ([]net.Interface, error)) SnowflakeOption {
	return idimpl.WithSnowflakeInterfacesFunc(interfaces)
}

// WithSnowflakePIDFunc sets the process-id provider used to derive the default worker ID.
func WithSnowflakePIDFunc(pid func() int) SnowflakeOption {
	return idimpl.WithSnowflakePIDFunc(pid)
}

// CreateSnowflakeWithOptions creates a Snowflake generator from options.
func CreateSnowflakeWithOptions(opts ...SnowflakeOption) *Snowflake {
	return idimpl.CreateSnowflakeWithOptions(opts...)
}

// NewIsolatedSnowflake creates a standalone Snowflake generator without singleton/cache lookup.
func NewIsolatedSnowflake(opts ...SnowflakeOption) *Snowflake {
	return idimpl.NewIsolatedSnowflake(opts...)
}

// GetSnowflake returns the package-level default Snowflake generator.
func GetSnowflake() *Snowflake { return GetSnowflakeWithOptions() }

// GetSnowflakeWithOptions returns the default singleton Snowflake generator, creating it with options if needed.
func GetSnowflakeWithOptions(opts ...SnowflakeOption) *Snowflake {
	return idimpl.GetSnowflakeWithOptions(opts...)
}

// ConfigureDefaultSnowflake replaces the default singleton Snowflake generator with options.
func ConfigureDefaultSnowflake(opts ...SnowflakeOption) *Snowflake {
	return idimpl.ConfigureDefaultSnowflake(opts...)
}

// GetSnowflakeWithWorker returns a cached Snowflake generator for workerID using default datacenter settings.
func GetSnowflakeWithWorker(workerID int64) *Snowflake {
	return GetSnowflakeWithWorkerWithOptions(workerID)
}

// GetSnowflakeWithWorkerWithOptions returns a singleton Snowflake generator for workerID using custom defaults.
func GetSnowflakeWithWorkerWithOptions(workerID int64, opts ...SnowflakeOption) *Snowflake {
	return idimpl.GetSnowflakeWithWorkerWithOptions(workerID, opts...)
}

// GetSnowflakeWithWorkerDataCenter returns a cached Snowflake generator for an explicit worker/datacenter pair.
func GetSnowflakeWithWorkerDataCenter(workerID, datacenterID int64) *Snowflake {
	return GetSnowflakeWithWorkerDataCenterWithOptions(workerID, datacenterID)
}

// GetSnowflakeWithWorkerDataCenterWithOptions returns a singleton Snowflake generator for worker/datacenter pair using custom clock options.
func GetSnowflakeWithWorkerDataCenterWithOptions(workerID, datacenterID int64, opts ...SnowflakeOption) *Snowflake {
	return idimpl.GetSnowflakeWithWorkerDataCenterWithOptions(workerID, datacenterID, opts...)
}

// GetDataCenterID derives a datacenter ID within maxDatacenterID from host network interfaces.
func GetDataCenterID(maxDatacenterID int64) int64 { return GetDataCenterIDWithOptions(maxDatacenterID) }

// GetDataCenterIDWithOptions derives a datacenter id using custom Snowflake providers.
func GetDataCenterIDWithOptions(maxDatacenterID int64, opts ...SnowflakeOption) int64 {
	return idimpl.GetDataCenterIDWithOptions(maxDatacenterID, opts...)
}

// GetWorkerID derives a worker ID within maxWorkerID from the process ID and datacenter ID.
func GetWorkerID(datacenterID, maxWorkerID int64) int64 {
	return GetWorkerIDWithOptions(datacenterID, maxWorkerID)
}

// GetWorkerIDWithOptions derives a worker id using custom Snowflake providers.
func GetWorkerIDWithOptions(datacenterID, maxWorkerID int64, opts ...SnowflakeOption) int64 {
	return idimpl.GetWorkerIDWithOptions(datacenterID, maxWorkerID, opts...)
}

// NanoId creates a NanoId using the default alphabet and length.
func NanoId() string { return NanoIdWithOptions() }

// NanoIdN creates a NanoId with an explicit length and the default alphabet.
func NanoIdN(n int) string { return NanoIdNWithOptions(n) }

// WithNanoIDRandomReader sets the entropy source used by NanoIdWithOptions.
func WithNanoIDRandomReader(reader io.Reader) NanoIDOption {
	return idimpl.WithNanoIDRandomReader(reader)
}

// WithNanoIDFallbackRandomSource sets the fallback PRNG used when NanoId random reads fail.
func WithNanoIDFallbackRandomSource(source *mathrand.Rand) NanoIDOption {
	return idimpl.WithNanoIDFallbackRandomSource(source)
}

// WithNanoIDAlphabet sets the alphabet used by NanoIdWithOptions.
func WithNanoIDAlphabet(alphabet string) NanoIDOption { return idimpl.WithNanoIDAlphabet(alphabet) }

// WithNanoIDLength sets the output length used by NanoIdWithOptions.
func WithNanoIDLength(length int) NanoIDOption { return idimpl.WithNanoIDLength(length) }

// NanoIdWithOptions creates a NanoId with custom generation options.
func NanoIdWithOptions(opts ...NanoIDOption) string { return idimpl.NanoIdWithOptions(opts...) }

// NanoIdNWithOptions creates a NanoId with explicit length and custom options.
func NanoIdNWithOptions(n int, opts ...NanoIDOption) string {
	return idimpl.NanoIdNWithOptions(n, opts...)
}

// GetSnowflakeNextID returns the next numeric ID from the default Snowflake generator.
func GetSnowflakeNextID() int64 { return GetSnowflakeNextIDWithOptions() }

// GetSnowflakeNextIDWithOptions returns the next ID from the default singleton Snowflake generator.
func GetSnowflakeNextIDWithOptions(opts ...SnowflakeOption) int64 {
	return idimpl.GetSnowflakeNextIDWithOptions(opts...)
}

// GetSnowflakeNextIDStr returns the next Snowflake ID from the default generator as a string.
func GetSnowflakeNextIDStr() string { return GetSnowflakeNextIDStrWithOptions() }

// GetSnowflakeNextIDStrWithOptions returns the next ID string from the default singleton Snowflake generator.
func GetSnowflakeNextIDStrWithOptions(opts ...SnowflakeOption) string {
	return idimpl.GetSnowflakeNextIDStrWithOptions(opts...)
}
