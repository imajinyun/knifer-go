# vftp: FTP adapter helpers

`vftp` provides provider-neutral list, download, and upload helpers for FTP-style remote file transfer workflows. It defines small interfaces for callers to inject their own FTP providers while keeping `go-knifer` free of network-client and credential dependencies.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Keep a reusable injected adapter | `New` with `WithProvider` | Use when a workflow performs multiple FTP-style operations through one provider. |
| List a remote directory once | `List` | Validates `ListRequest.RemoteDir`, rejects NUL bytes, then delegates to the provider. |
| Download small in-memory content | `Download` | Uses `DownloadRequest.MaxBytes` to bound provider-returned content. |
| Upload small in-memory content | `Upload` | Validates content size before provider delegation and provider-reported size after upload. |
| Classify entries | `EntryTypeFile`, `EntryTypeDirectory`, `EntryTypeUnknown` | Keeps directory metadata provider-neutral. |
| Handle stable errors | `ErrInvalidListRequest`, `ErrInvalidDownloadRequest`, `ErrInvalidUploadRequest`, `ErrTransferLimitExceeded`, `ErrMissingProvider` | Use `errors.Is` when branching on validation or limit failures. |

## FTP safety checklist

- Always provide a `Provider`; the facade does not open sockets, read credentials, or configure FTP/FTPS sessions.
- Pass cancellable contexts so provider dials, listings, downloads, and uploads can stop on timeout or shutdown.
- Set `MaxBytes` for downloads and uploads to keep in-memory transfer size bounded.
- Validate remote path policy in the provider or application; `vftp` rejects blank/NUL values but does not normalize server paths.
- Keep TLS mode, passive/active mode, authentication, retries, logging, and metrics in the provider where server details are known.
- Do not log transfer content or credentials in provider error handling.

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

## When not to use vftp

- Use a concrete FTP/FTPS client directly when you need real network sessions, passive/active mode controls, TLS negotiation, streaming, or provider extensions.
- Use `vssh` or an SFTP client when the server exposes SSH/SFTP rather than FTP/FTPS.
- Avoid this facade for large transfers because download and upload content is represented in memory.
- Use local file helpers when the source or destination path is local; `vftp` does not read or write local files.

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

## Related packages

- Use `vssh` when file transfer should run over SSH/SFTP instead of FTP or FTPS.
- Use `vfile` when downloaded paths, temporary files, or local filesystem policy need validation.
- Use `vlog` and `verr` when transfer failures need structured diagnostics and error wrapping.

## Benchmarks and trade-offs

- Request validation and defensive copying make provider boundaries safer and more testable, with small overhead compared with real network calls.
- The one-off helpers create short-lived clients for concise call sites. Reuse `New(WithProvider(...))` when a workflow shares provider setup.
- In-memory transfers make limit checks straightforward but move streaming and backpressure out of scope.
- Provider neutrality keeps go-knifer dependency-light while requiring applications to document their own FTP security and retry behavior.

## FAQ

### Does `vftp` include an FTP client?

No. It only defines provider-neutral request/response types and delegates to an injected provider.

### Where should credentials and TLS settings live?

In the provider or application configuration. The facade deliberately does not discover credentials, configure TLS, or open network connections.

### Why does `Download` return bytes instead of a stream?

The current facade is optimized for small, deterministic transfer adapters. Use a concrete FTP client directly for streaming large files.

### What path validation does `vftp` perform?

It rejects blank and NUL-containing remote paths. Provider-specific normalization, chroot rules, and path allowlists belong to the provider/application boundary.
