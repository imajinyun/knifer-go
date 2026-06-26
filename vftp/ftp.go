package vftp

import (
	"context"

	ftpimpl "github.com/imajinyun/knifer-go/internal/ftp"
)

// EntryType describes the kind of remote FTP entry.
type EntryType = ftpimpl.EntryType

const (
	// EntryTypeUnknown marks an entry whose provider-specific type is not known.
	EntryTypeUnknown = ftpimpl.EntryTypeUnknown
	// EntryTypeFile marks a regular remote file.
	EntryTypeFile = ftpimpl.EntryTypeFile
	// EntryTypeDirectory marks a remote directory.
	EntryTypeDirectory = ftpimpl.EntryTypeDirectory
)

// Entry describes one remote FTP directory entry.
type Entry = ftpimpl.Entry

// ListRequest describes a remote directory listing request.
type ListRequest = ftpimpl.ListRequest

// ListResponse contains remote directory entries.
type ListResponse = ftpimpl.ListResponse

// DownloadRequest describes a remote file download request.
type DownloadRequest = ftpimpl.DownloadRequest

// DownloadResponse contains downloaded bytes and transfer metadata.
type DownloadResponse = ftpimpl.DownloadResponse

// UploadRequest describes an in-memory upload to a remote path.
type UploadRequest = ftpimpl.UploadRequest

// UploadResponse contains upload transfer metadata.
type UploadResponse = ftpimpl.UploadResponse

// Provider performs provider-specific FTP operations for validated requests.
type Provider = ftpimpl.Provider

// Client routes validated FTP requests to an injected provider.
type Client = ftpimpl.Client

// Option customizes a Client.
type Option = ftpimpl.Option

var (
	// ErrInvalidListRequest reports a malformed remote listing request.
	ErrInvalidListRequest = ftpimpl.ErrInvalidListRequest
	// ErrInvalidDownloadRequest reports a malformed remote download request.
	ErrInvalidDownloadRequest = ftpimpl.ErrInvalidDownloadRequest
	// ErrInvalidUploadRequest reports a malformed remote upload request.
	ErrInvalidUploadRequest = ftpimpl.ErrInvalidUploadRequest
	// ErrTransferLimitExceeded reports that a transfer exceeded its configured byte limit.
	ErrTransferLimitExceeded = ftpimpl.ErrTransferLimitExceeded
	// ErrMissingProvider reports that an FTP call has no configured provider.
	ErrMissingProvider = ftpimpl.ErrMissingProvider
)

// WithProvider sets the provider used by list, download, and upload calls.
func WithProvider(provider Provider) Option { return ftpimpl.WithProvider(provider) }

// New returns a provider-neutral FTP client.
func New(opts ...Option) *Client { return ftpimpl.New(opts...) }

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
