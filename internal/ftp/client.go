package ftp

import (
	"context"
	"errors"
	"fmt"
)

// ErrMissingProvider reports that an FTP call has no configured provider.
var ErrMissingProvider = errors.New("ftp: missing provider")

// Provider performs provider-specific FTP operations for validated requests.
type Provider interface {
	List(ctx context.Context, request ListRequest) (ListResponse, error)
	Download(ctx context.Context, request DownloadRequest) (DownloadResponse, error)
	Upload(ctx context.Context, request UploadRequest) (UploadResponse, error)
}

// Client routes validated FTP requests to an injected provider.
type Client struct {
	provider Provider
}

// Option customizes a Client.
type Option func(*Client)

// WithProvider sets the provider used by list, download, and upload calls.
func WithProvider(provider Provider) Option {
	return func(c *Client) {
		if provider != nil {
			c.provider = provider
		}
	}
}

// New returns a provider-neutral FTP client.
func New(opts ...Option) *Client {
	client := &Client{}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

// List validates request and delegates to the configured provider.
func (c *Client) List(ctx context.Context, request ListRequest) (ListResponse, error) {
	if err := request.Validate(); err != nil {
		return ListResponse{}, err
	}
	if c == nil || c.provider == nil {
		return ListResponse{}, ErrMissingProvider
	}
	response, err := c.provider.List(ctx, request.Clone())
	if err != nil {
		return ListResponse{}, fmt.Errorf("ftp list provider: %w", err)
	}
	return response.Clone(), nil
}

// Download validates request and delegates to the configured provider.
func (c *Client) Download(ctx context.Context, request DownloadRequest) (DownloadResponse, error) {
	if err := request.Validate(); err != nil {
		return DownloadResponse{}, err
	}
	if c == nil || c.provider == nil {
		return DownloadResponse{}, ErrMissingProvider
	}
	response, err := c.provider.Download(ctx, request.Clone())
	if err != nil {
		return DownloadResponse{}, fmt.Errorf("ftp download provider: %w", err)
	}
	if request.MaxBytes > 0 && int64(len(response.Content)) > request.MaxBytes {
		return DownloadResponse{}, fmt.Errorf("%w: download content has %d bytes, max %d", ErrTransferLimitExceeded, len(response.Content), request.MaxBytes)
	}
	return response.Clone(), nil
}

// Upload validates request and delegates to the configured provider.
func (c *Client) Upload(ctx context.Context, request UploadRequest) (UploadResponse, error) {
	if err := request.Validate(); err != nil {
		return UploadResponse{}, err
	}
	if c == nil || c.provider == nil {
		return UploadResponse{}, ErrMissingProvider
	}
	response, err := c.provider.Upload(ctx, request.Clone())
	if err != nil {
		return UploadResponse{}, fmt.Errorf("ftp upload provider: %w", err)
	}
	if request.MaxBytes > 0 && response.Size > request.MaxBytes {
		return UploadResponse{}, fmt.Errorf("%w: uploaded %d bytes, max %d", ErrTransferLimitExceeded, response.Size, request.MaxBytes)
	}
	return response.Clone(), nil
}
