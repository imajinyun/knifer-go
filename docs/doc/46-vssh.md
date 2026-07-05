# vssh: SSH/SFTP adapter helpers

`vssh` provides provider-neutral SSH command and SFTP-style transfer helpers. It defines a small interface for callers to inject their own SSH/SFTP providers while keeping `knifer-go` free of network-client, key-parsing, and credential dependencies.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Download`
- `New`
- `WithProvider`
- `List`
- `Run`

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Keep a reusable injected adapter | `New` with `WithProvider` | Use when application code performs multiple SSH/SFTP-style operations through one provider. |
| Run a one-off remote command | `Run` | Validates `CommandRequest`, enforces `MaxOutputBytes`, then delegates to the provider. |
| List a remote directory | `List` | Validates `ListRequest.RemoteDir` before provider delegation. |
| Download small in-memory content | `Download` | Uses `DownloadRequest.MaxBytes` to bound provider-returned content. |
| Upload small in-memory content | `Upload` | Validates request content size before calling the provider and validates provider-reported size after upload. |
| Classify remote entries | `EntryTypeFile`, `EntryTypeDirectory`, `EntryTypeSymlink` | Entry types are provider-neutral metadata; provider-specific fields stay outside the facade. |
| Check invalid requests and limits | `ErrInvalid*`, `ErrOutputLimitExceeded`, `ErrTransferLimitExceeded`, `ErrMissingProvider` | Use `errors.Is` when callers need a stable error contract. |

## SSH/SFTP safety checklist

- Always inject a provider; the facade intentionally has no built-in SSH client, socket, credential, or key-loading behavior.
- Pass cancellable contexts to every operation so providers can stop dials, commands, and transfers during shutdown or timeout.
- Set `MaxOutputBytes` for commands that could produce unbounded stdout/stderr.
- Set `MaxBytes` for transfers and keep this facade for small in-memory payloads rather than streaming large files.
- Treat command strings and args as provider inputs, not shell-quoted safe strings. The provider decides whether and how a shell is used.
- Validate remote paths at the application/provider boundary; `vssh` rejects blank/NUL paths but does not normalize server paths.
- Keep host-key verification, authentication, retry policy, logging, and metrics in the provider where connection details are known.

## When to use

Use `vssh` when application code needs a stable internal contract for remote command execution or in-memory SFTP-style transfer operations, but connection setup, authentication, host-key verification, retries, and provider-specific behavior belong to the application boundary.

Use a dedicated SSH/SFTP client directly when you need streaming transfers, PTY/session controls, port forwarding, SCP, filesystem paths, known-hosts parsing, or connection-pool lifecycle management that is not part of the `vssh` MVP.

## Provider injection

`vssh` has no built-in network provider. It does not read usernames, passwords, private keys, environment variables, or local files. Tests and applications provide behavior by implementing `Provider`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/knifer-go/vssh"
)

type sshProvider struct{}

func (sshProvider) Run(ctx context.Context, request vssh.CommandRequest) (vssh.CommandResponse, error) {
	return vssh.CommandResponse{ExitCode: 0, Stdout: []byte("hello gopher")}, nil
}

func (sshProvider) List(ctx context.Context, request vssh.ListRequest) (vssh.ListResponse, error) {
	return vssh.ListResponse{Entries: []vssh.Entry{{Name: "report.csv", Path: request.RemoteDir + "/report.csv", Type: vssh.EntryTypeFile}}}, nil
}

func (sshProvider) Download(ctx context.Context, request vssh.DownloadRequest) (vssh.DownloadResponse, error) {
	return vssh.DownloadResponse{RemotePath: request.RemotePath, Content: []byte("id,total\n1,42\n"), Size: 14}, nil
}

func (sshProvider) Upload(ctx context.Context, request vssh.UploadRequest) (vssh.UploadResponse, error) {
	return vssh.UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
}

func main() {
	client := vssh.New(vssh.WithProvider(sshProvider{}))
	response, err := client.Run(context.Background(), vssh.CommandRequest{Command: "printf"})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(response.Stdout))
}
```

## Command example

For one-off calls, use `vssh.Run` with an injected provider:

```go
response, err := vssh.Run(context.Background(), sshProvider{}, vssh.CommandRequest{
	Command:        "printf",
	Args:           []string{"hello"},
	MaxOutputBytes: 1024,
})
if err != nil {
	panic(err)
}
fmt.Println(response.ExitCode, string(response.Stdout))
```

`CommandRequest.Validate` requires a non-blank command, rejects NUL bytes in command text or args, and requires `MaxOutputBytes` to be non-negative. When `MaxOutputBytes` is greater than zero, `Run` rejects provider responses whose combined stdout and stderr length exceeds the limit with `ErrOutputLimitExceeded`.

## Listing example

```go
response, err := vssh.List(context.Background(), sshProvider{}, vssh.ListRequest{RemoteDir: "/pub"})
if err != nil {
	panic(err)
}
fmt.Println(len(response.Entries), response.Entries[0].Type)
```

`ListRequest.Validate` requires a non-blank remote directory and rejects NUL bytes.

## Download example

```go
response, err := vssh.Download(context.Background(), sshProvider{}, vssh.DownloadRequest{
	RemotePath: "/pub/report.csv",
	MaxBytes:   1024,
})
if err != nil {
	panic(err)
}
fmt.Println(response.RemotePath, len(response.Content))
```

`DownloadRequest.Validate` requires a non-blank remote path, rejects NUL bytes, and requires `MaxBytes` to be non-negative. When `MaxBytes` is greater than zero, `Download` rejects provider responses whose content length exceeds the limit with `ErrTransferLimitExceeded`.

## Upload example

```go
response, err := vssh.Upload(context.Background(), sshProvider{}, vssh.UploadRequest{
	RemotePath: "/pub/out.csv",
	Content:    []byte("ok\n"),
	MaxBytes:   1024,
})
if err != nil {
	panic(err)
}
fmt.Println(response.RemotePath, response.Size)
```

`UploadRequest.Validate` requires a non-blank remote path, rejects NUL bytes, requires `MaxBytes` to be non-negative, and rejects in-memory content that already exceeds the configured limit. `Upload` also checks the provider-reported uploaded size against `MaxBytes`.

## Security boundary

`vssh` treats commands, args, remote paths, output sizes, and transfer sizes as validation inputs only. It does not quote commands, invoke shells, normalize server paths, join local filesystem paths, discover credentials, parse private keys, validate host keys, log command output, or open sockets. Providers are responsible for SSH connection security, authentication, host-key verification, TLS-equivalent policy, retry policy, streaming behavior, and provider-specific path rules.

Requests and responses are defensively copied around provider calls so callers and providers can mutate their own values without sharing slices or maps unexpectedly.

## When not to use vssh

- Use a concrete SSH/SFTP library directly when you need real network connections, host-key stores, private-key parsing, PTY/session controls, port forwarding, SCP, or streaming transfers.
- Use a job runner or orchestration tool when command execution needs scheduling, retries, audit logs, or fleet-level state.
- Use local filesystem helpers when paths are local; `vssh` does not join, clean, or open local paths.
- Avoid this facade for large file transfers because download and upload payloads are represented in memory.

## Out of scope

- Built-in SSH, SFTP, SCP, or FTP clients.
- Credential discovery, private-key parsing, or environment-variable loading.
- Real shell execution, quoting, PTY allocation, port forwarding, or session lifecycle.
- Streaming transfers or local filesystem reads/writes.
- Directory creation, deletion, rename, chmod, or server feature negotiation.
- Retry, rate limiting, tracing, logging, metrics, or connection pooling.

## Validation

Focused checks:

```bash
go test ./internal/ssh ./vssh
go test -bench=. -benchmem -run=^$ ./internal/ssh ./vssh
```

Governance checks for public API and catalog changes:

```bash
UPDATE_API=1 make api-check
make docs-gen
make docs-check
make tools-check
make agent-check
make agent-security-check
```

## Related packages

- Use `vftp` when file transfer should use FTP/FTPS rather than SSH/SFTP.
- Use `vcli` when local command execution is enough and no remote SSH transport is required.
- Use `vfile` when transferred files need local path, temp-file, or filesystem policy checks.

## Benchmarks and trade-offs

- The facade adds request validation, defensive copying, and limit checks around provider calls. That overhead is small for network-bound operations but visible in microbenchmarks.
- In-memory transfers simplify testing and provider contracts, but they are not appropriate for multi-gigabyte streams.
- `Run`, `List`, `Download`, and `Upload` create short-lived clients for convenience. Reuse `New(WithProvider(...))` when multiple calls share one provider.
- Provider-neutral structs keep knifer-go dependency-light while shifting connection lifecycle and protocol-specific tuning to the application.

## FAQ

### Does `vssh` open SSH connections?

No. It validates provider-neutral requests and delegates to an injected provider. The application owns real SSH/SFTP clients, credentials, and host-key verification.

### Does `CommandRequest` protect me from shell injection?

No. The facade rejects malformed command text such as blank or NUL-containing values, but quoting and shell selection are provider responsibilities.

### Why are transfers in memory?

The MVP contract favors deterministic tests and simple adapters. Use a concrete SFTP client directly when streaming or local file paths are required.

### Why does the facade defensively copy data?

Copies prevent callers and providers from accidentally sharing mutable slices or maps after validation and limit checks.
