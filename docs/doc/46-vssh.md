# vssh: SSH/SFTP adapter helpers

`vssh` provides provider-neutral SSH command and SFTP-style transfer helpers. It defines a small interface for callers to inject their own SSH/SFTP providers while keeping `go-knifer` free of network-client, key-parsing, and credential dependencies.

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

	"github.com/imajinyun/go-knifer/vssh"
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

