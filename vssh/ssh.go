package vssh

import (
	"context"

	sshimpl "github.com/imajinyun/go-knifer/internal/ssh"
)

// EntryType describes the kind of remote SFTP entry.
type EntryType = sshimpl.EntryType

const (
	// EntryTypeUnknown marks an entry whose provider-specific type is not known.
	EntryTypeUnknown = sshimpl.EntryTypeUnknown
	// EntryTypeFile marks a regular remote file.
	EntryTypeFile = sshimpl.EntryTypeFile
	// EntryTypeDirectory marks a remote directory.
	EntryTypeDirectory = sshimpl.EntryTypeDirectory
	// EntryTypeSymlink marks a remote symbolic link.
	EntryTypeSymlink = sshimpl.EntryTypeSymlink
)

// Entry describes one remote SFTP directory entry.
type Entry = sshimpl.Entry

// CommandRequest describes a remote SSH command request.
type CommandRequest = sshimpl.CommandRequest

// CommandResponse contains remote command output and exit metadata.
type CommandResponse = sshimpl.CommandResponse

// ListRequest describes a remote SFTP directory listing request.
type ListRequest = sshimpl.ListRequest

// ListResponse contains remote SFTP directory entries.
type ListResponse = sshimpl.ListResponse

// DownloadRequest describes a remote SFTP file download request.
type DownloadRequest = sshimpl.DownloadRequest

// DownloadResponse contains downloaded bytes and transfer metadata.
type DownloadResponse = sshimpl.DownloadResponse

// UploadRequest describes an in-memory SFTP upload to a remote path.
type UploadRequest = sshimpl.UploadRequest

// UploadResponse contains upload transfer metadata.
type UploadResponse = sshimpl.UploadResponse

// Provider performs provider-specific SSH/SFTP operations for validated requests.
type Provider = sshimpl.Provider

// Client routes validated SSH/SFTP requests to an injected provider.
type Client = sshimpl.Client

// Option customizes a Client.
type Option = sshimpl.Option

var (
	// ErrInvalidCommandRequest reports a malformed remote command request.
	ErrInvalidCommandRequest = sshimpl.ErrInvalidCommandRequest
	// ErrInvalidListRequest reports a malformed remote listing request.
	ErrInvalidListRequest = sshimpl.ErrInvalidListRequest
	// ErrInvalidDownloadRequest reports a malformed remote download request.
	ErrInvalidDownloadRequest = sshimpl.ErrInvalidDownloadRequest
	// ErrInvalidUploadRequest reports a malformed remote upload request.
	ErrInvalidUploadRequest = sshimpl.ErrInvalidUploadRequest
	// ErrOutputLimitExceeded reports that command output exceeded its configured byte limit.
	ErrOutputLimitExceeded = sshimpl.ErrOutputLimitExceeded
	// ErrTransferLimitExceeded reports that a transfer exceeded its configured byte limit.
	ErrTransferLimitExceeded = sshimpl.ErrTransferLimitExceeded
	// ErrMissingProvider reports that an SSH/SFTP call has no configured provider.
	ErrMissingProvider = sshimpl.ErrMissingProvider
)

// WithProvider sets the provider used by command and transfer calls.
func WithProvider(provider Provider) Option { return sshimpl.WithProvider(provider) }

// New returns a provider-neutral SSH/SFTP client.
func New(opts ...Option) *Client { return sshimpl.New(opts...) }

// Run validates request and delegates to provider through a short-lived Client.
func Run(ctx context.Context, provider Provider, request CommandRequest) (CommandResponse, error) {
	return New(WithProvider(provider)).Run(ctx, request)
}

// List validates request and delegates to provider through a short-lived Client.
func List(ctx context.Context, provider Provider, request ListRequest) (ListResponse, error) {
	return New(WithProvider(provider)).List(ctx, request)
}

// Download validates request and delegates to provider through a short-lived Client.
func Download(ctx context.Context, provider Provider, request DownloadRequest) (DownloadResponse, error) {
	return New(WithProvider(provider)).Download(ctx, request)
}

// Upload validates request and delegates to provider through a short-lived Client.
func Upload(ctx context.Context, provider Provider, request UploadRequest) (UploadResponse, error) {
	return New(WithProvider(provider)).Upload(ctx, request)
}
