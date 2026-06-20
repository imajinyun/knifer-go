# vftp: FTP adapter helpers

`vftp` provides provider-neutral list, download, and upload helpers for FTP-style remote file transfer workflows. It defines small interfaces for callers to inject their own FTP providers while keeping `go-knifer` free of network-client and credential dependencies.

## When to use

Use `vftp` when application code needs a stable internal contract for FTP listing or in-memory transfer operations, but connection setup, authentication, retries, TLS mode, and provider-specific behavior belong to the application boundary.

Use a dedicated FTP client directly when you need streaming transfers, filesystem paths, active/passive mode controls, provider-specific extensions, or connection-pool lifecycle management that is not part of the `vftp` MVP.

## Provider injection

`vftp` has no built-in network provider. It does not read usernames, passwords, environment variables, or local files. Tests and applications provide behavior by implementing `Provider`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vftp"
)

type ftpProvider struct{}

func (ftpProvider) List(ctx context.Context, request vftp.ListRequest) (vftp.ListResponse, error) {
	return vftp.ListResponse{Entries: []vftp.Entry{{Name: "report.csv", Path: request.RemoteDir + "/report.csv", Type: vftp.EntryTypeFile}}}, nil
}

func (ftpProvider) Download(ctx context.Context, request vftp.DownloadRequest) (vftp.DownloadResponse, error) {
	return vftp.DownloadResponse{RemotePath: request.RemotePath, Content: []byte("id,total\n1,42\n"), Size: 14}, nil
}

func (ftpProvider) Upload(ctx context.Context, request vftp.UploadRequest) (vftp.UploadResponse, error) {
	return vftp.UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
}

func main() {
	client := vftp.New(vftp.WithProvider(ftpProvider{}))
	response, err := client.List(context.Background(), vftp.ListRequest{RemoteDir: "/pub"})
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Entries[0].Name)
}
```

## Listing example

For one-off calls, use `vftp.List` with an injected provider:

```go
response, err := vftp.List(context.Background(), ftpProvider{}, vftp.ListRequest{RemoteDir: "/pub"})
if err != nil {
	panic(err)
}
fmt.Println(len(response.Entries), response.Entries[0].Type)
```

`ListRequest.Validate` requires a non-blank remote directory and rejects NUL bytes.

## Download example

```go
response, err := vftp.Download(context.Background(), ftpProvider{}, vftp.DownloadRequest{
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
response, err := vftp.Upload(context.Background(), ftpProvider{}, vftp.UploadRequest{
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

`vftp` treats remote paths and transfer sizes as validation inputs only. It does not normalize server paths, join local filesystem paths, discover credentials, log transfer data, or open sockets. Providers are responsible for FTP connection security, authentication, TLS configuration, retry policy, streaming behavior, and any provider-specific path rules.

Requests and responses are defensively copied around provider calls so callers and providers can mutate their own values without sharing slices or maps unexpectedly.

## Out of scope

- Built-in FTP, FTPS, SSH, or SFTP clients.
- Credential discovery or environment-variable loading.
- Streaming transfers or local filesystem reads/writes.
- Directory creation, deletion, rename, chmod, or server feature negotiation.
- Retry, rate limiting, tracing, logging, metrics, or connection pooling.

## Validation

Focused checks:

```bash
go test ./internal/ftp ./vftp
go test -bench=. -benchmem -run=^$ ./internal/ftp ./vftp
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

