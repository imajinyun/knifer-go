package ssh

import (
	"context"
	"errors"
	"fmt"
)

// ErrMissingProvider reports that an SSH/SFTP call has no configured provider.
var ErrMissingProvider = errors.New("ssh: missing provider")

// Provider performs provider-specific SSH/SFTP operations for validated requests.
type Provider interface {
	Run(ctx context.Context, request CommandRequest) (CommandResponse, error)
	List(ctx context.Context, request ListRequest) (ListResponse, error)
	Download(ctx context.Context, request DownloadRequest) (DownloadResponse, error)
	Upload(ctx context.Context, request UploadRequest) (UploadResponse, error)
}

// Client routes validated SSH/SFTP requests to an injected provider.
type Client struct {
	provider Provider
}

// Option customizes a Client.
type Option func(*Client)

// WithProvider sets the provider used by command and transfer calls.
func WithProvider(provider Provider) Option {
	return func(c *Client) {
		if provider != nil {
			c.provider = provider
		}
	}
}

// New returns a provider-neutral SSH/SFTP client.
func New(opts ...Option) *Client {
	client := &Client{}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

// Run validates request and delegates to the configured provider.
func (c *Client) Run(ctx context.Context, request CommandRequest) (CommandResponse, error) {
	if err := request.Validate(); err != nil {
		return CommandResponse{}, err
	}
	if c == nil || c.provider == nil {
		return CommandResponse{}, ErrMissingProvider
	}
	response, err := c.provider.Run(ctx, request.Clone())
	if err != nil {
		return CommandResponse{}, fmt.Errorf("ssh run provider: %w", err)
	}
	if request.MaxOutputBytes > 0 {
		outputBytes := int64(len(response.Stdout)) + int64(len(response.Stderr))
		if outputBytes > request.MaxOutputBytes {
			return CommandResponse{}, fmt.Errorf("%w: command output has %d bytes, max %d", ErrOutputLimitExceeded, outputBytes, request.MaxOutputBytes)
		}
	}
	return response.Clone(), nil
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
		return ListResponse{}, fmt.Errorf("ssh list provider: %w", err)
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
		return DownloadResponse{}, fmt.Errorf("ssh download provider: %w", err)
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
		return UploadResponse{}, fmt.Errorf("ssh upload provider: %w", err)
	}
	if request.MaxBytes > 0 && response.Size > request.MaxBytes {
		return UploadResponse{}, fmt.Errorf("%w: uploaded %d bytes, max %d", ErrTransferLimitExceeded, response.Size, request.MaxBytes)
	}
	return response.Clone(), nil
}
