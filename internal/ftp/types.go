package ftp

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidListRequest reports a malformed remote listing request.
	ErrInvalidListRequest = errors.New("ftp: invalid list request")
	// ErrInvalidDownloadRequest reports a malformed remote download request.
	ErrInvalidDownloadRequest = errors.New("ftp: invalid download request")
	// ErrInvalidUploadRequest reports a malformed remote upload request.
	ErrInvalidUploadRequest = errors.New("ftp: invalid upload request")
	// ErrTransferLimitExceeded reports that a transfer exceeded its configured byte limit.
	ErrTransferLimitExceeded = errors.New("ftp: transfer limit exceeded")
)

// EntryType describes the kind of remote FTP entry.
type EntryType string

const (
	// EntryTypeUnknown marks an entry whose provider-specific type is not known.
	EntryTypeUnknown EntryType = "unknown"
	// EntryTypeFile marks a regular remote file.
	EntryTypeFile EntryType = "file"
	// EntryTypeDirectory marks a remote directory.
	EntryTypeDirectory EntryType = "directory"
)

// Entry describes one remote FTP directory entry.
type Entry struct {
	Name     string
	Path     string
	Type     EntryType
	Size     int64
	Modified time.Time
	Metadata map[string]string
}

// ListRequest describes a remote directory listing request.
type ListRequest struct {
	RemoteDir string
	Recursive bool
	Metadata  map[string]string
}

// ListResponse contains remote directory entries.
type ListResponse struct {
	Entries  []Entry
	Metadata map[string]string
}

// DownloadRequest describes a remote file download request.
type DownloadRequest struct {
	RemotePath string
	MaxBytes   int64
	Metadata   map[string]string
}

// DownloadResponse contains downloaded bytes and transfer metadata.
type DownloadResponse struct {
	RemotePath string
	Content    []byte
	Size       int64
	Metadata   map[string]string
}

// UploadRequest describes an in-memory upload to a remote path.
type UploadRequest struct {
	RemotePath string
	Content    []byte
	MaxBytes   int64
	Metadata   map[string]string
}

// UploadResponse contains upload transfer metadata.
type UploadResponse struct {
	RemotePath string
	Size       int64
	Metadata   map[string]string
}

// Validate checks whether r has a valid remote directory.
func (r ListRequest) Validate() error {
	if err := validateRemotePath(r.RemoteDir, ErrInvalidListRequest); err != nil {
		return err
	}
	return nil
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r ListRequest) Clone() ListRequest {
	return ListRequest{RemoteDir: r.RemoteDir, Recursive: r.Recursive, Metadata: cloneStringMap(r.Metadata)}
}

// Clone returns a response copy that callers can mutate independently.
func (r ListResponse) Clone() ListResponse {
	return ListResponse{Entries: cloneEntries(r.Entries), Metadata: cloneStringMap(r.Metadata)}
}

// Validate checks whether r has a valid remote file and byte limit.
func (r DownloadRequest) Validate() error {
	if err := validateRemotePath(r.RemotePath, ErrInvalidDownloadRequest); err != nil {
		return err
	}
	return validateLimit(r.MaxBytes, ErrInvalidDownloadRequest)
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r DownloadRequest) Clone() DownloadRequest {
	return DownloadRequest{RemotePath: r.RemotePath, MaxBytes: r.MaxBytes, Metadata: cloneStringMap(r.Metadata)}
}

// Clone returns a response copy that callers can mutate independently.
func (r DownloadResponse) Clone() DownloadResponse {
	return DownloadResponse{
		RemotePath: r.RemotePath,
		Content:    append([]byte(nil), r.Content...),
		Size:       r.Size,
		Metadata:   cloneStringMap(r.Metadata),
	}
}

// Validate checks whether r has a valid remote file, byte limit, and content size.
func (r UploadRequest) Validate() error {
	if err := validateRemotePath(r.RemotePath, ErrInvalidUploadRequest); err != nil {
		return err
	}
	if err := validateLimit(r.MaxBytes, ErrInvalidUploadRequest); err != nil {
		return err
	}
	if r.MaxBytes > 0 && int64(len(r.Content)) > r.MaxBytes {
		return fmt.Errorf("%w: upload content has %d bytes, max %d", ErrTransferLimitExceeded, len(r.Content), r.MaxBytes)
	}
	return nil
}

// Clone returns a request copy that callers and providers can mutate independently.
func (r UploadRequest) Clone() UploadRequest {
	return UploadRequest{
		RemotePath: r.RemotePath,
		Content:    append([]byte(nil), r.Content...),
		MaxBytes:   r.MaxBytes,
		Metadata:   cloneStringMap(r.Metadata),
	}
}

// Clone returns a response copy that callers can mutate independently.
func (r UploadResponse) Clone() UploadResponse {
	return UploadResponse{RemotePath: r.RemotePath, Size: r.Size, Metadata: cloneStringMap(r.Metadata)}
}

func validateRemotePath(path string, sentinel error) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("%w: remote path is required", sentinel)
	}
	if strings.ContainsRune(path, '\x00') {
		return fmt.Errorf("%w: remote path contains nul byte", sentinel)
	}
	return nil
}

func validateLimit(maxBytes int64, sentinel error) error {
	if maxBytes < 0 {
		return fmt.Errorf("%w: max bytes must be non-negative", sentinel)
	}
	return nil
}

func cloneEntries(entries []Entry) []Entry {
	if len(entries) == 0 {
		return nil
	}
	clone := make([]Entry, len(entries))
	for i, entry := range entries {
		clone[i] = entry
		clone[i].Metadata = cloneStringMap(entry.Metadata)
	}
	return clone
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	clone := make(map[string]string, len(values))
	for k, v := range values {
		clone[k] = v
	}
	return clone
}
