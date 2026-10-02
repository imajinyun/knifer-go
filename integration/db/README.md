# Database integration contracts

This standalone Go module imports the public `vdb` API and pins its own SQLite,
PostgreSQL and MySQL drivers. The main library module does not depend on them.
Root `go test ./...` does not traverse this module; CI runs it explicitly.

For SQLite only:

```sh
DB_TEST_ENGINES=sqlite make db-integration-check
```

For all engines, provision dedicated test containers:

```sh
docker compose -p knifer-db-test -f integration/db/compose.yml up -d --wait
export KNIFER_TEST_POSTGRES_DSN='postgres://knifer:fixture-only@127.0.0.1:15432/knifer?sslmode=disable'
export KNIFER_TEST_MYSQL_DSN='knifer:fixture-only@tcp(127.0.0.1:13306)/knifer?parseTime=true&loc=UTC&timeout=5s'
make db-integration-check
docker compose -p knifer-db-test -f integration/db/compose.yml down -v
```

Use these DSNs only with the dedicated fixture databases. The credentials are
public test fixtures. Ports can be changed with KNIFER_PG_PORT/KNIFER_MYSQL_PORT;
update the DSNs accordingly. The tests create uniquely named tables and clean
up only those tables. Never point the suite at production data.

The default Make target requires all three engines and fails if a requested DSN
is missing. Direct `go test` in this module defaults to SQLite for local use.
The CI database job requests all engines, uses pinned service images, records
its exit through the CI recorder and is required by final admission.

Contracts cover named/positional parameter binding, repeated parameters,
injection-shaped values, upsert state and affected rows, pagination/count,
NULL/binary/time/numeric scans, overflow rejection, row connection release,
missing rows, callback/panic rollback, commit/savepoints, batch partial failure
inside and outside transactions, cancellation and connection recovery.

MySQL update-upsert normally reports two affected rows while PostgreSQL/SQLite
report one; this suite uses the driver's default clientFoundRows setting.
PostgreSQL additionally tests a real deferred constraint failure at COMMIT.
Driver failures that require injection (such as rollback transport failures)
remain explicitly covered by the existing unit-test fakes; they are not claimed
as naturally triggered on every engine here.

Raw `go test -json` evidence is saved to `.aiflow/db-integration/results.jsonl`.
For dependency hygiene, run `go -C integration/db mod tidy -diff` and keep this
module's go.mod/go.sum separate from the root module.
