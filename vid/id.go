package vid

import idimpl "github.com/imajinyun/go-knifer/internal/id"

type Snowflake = idimpl.Snowflake

func RandomUUID() string     { return idimpl.RandomUUID() }
func SimpleUUID() string     { return idimpl.SimpleUUID() }
func FastUUID() string       { return idimpl.FastUUID() }
func FastSimpleUUID() string { return idimpl.FastSimpleUUID() }
func UUID() string           { return idimpl.SimpleUUID() }
func ObjectId() string       { return idimpl.ObjectId() }

func CreateSnowflake(workerID, datacenterID int64) *Snowflake {
	return idimpl.CreateSnowflake(workerID, datacenterID)
}

func GetSnowflake() *Snowflake { return idimpl.GetSnowflake() }

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

func GetSnowflakeNextID() int64     { return idimpl.GetSnowflakeNextID() }
func GetSnowflakeNextIDStr() string { return idimpl.GetSnowflakeNextIDStr() }
