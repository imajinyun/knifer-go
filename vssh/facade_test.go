package vssh_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vssh"
)

type providerFunc struct {
	run      func(context.Context, vssh.CommandRequest) (vssh.CommandResponse, error)
	list     func(context.Context, vssh.ListRequest) (vssh.ListResponse, error)
	download func(context.Context, vssh.DownloadRequest) (vssh.DownloadResponse, error)
	upload   func(context.Context, vssh.UploadRequest) (vssh.UploadResponse, error)
}

func (p providerFunc) Run(ctx context.Context, request vssh.CommandRequest) (vssh.CommandResponse, error) {
	return p.run(ctx, request)
}

func (p providerFunc) List(ctx context.Context, request vssh.ListRequest) (vssh.ListResponse, error) {
	return p.list(ctx, request)
}

func (p providerFunc) Download(ctx context.Context, request vssh.DownloadRequest) (vssh.DownloadResponse, error) {
	return p.download(ctx, request)
}

func (p providerFunc) Upload(ctx context.Context, request vssh.UploadRequest) (vssh.UploadResponse, error) {
	return p.upload(ctx, request)
}

func TestRunFacade(t *testing.T) {
	provider := providerFunc{
		run: func(ctx context.Context, request vssh.CommandRequest) (vssh.CommandResponse, error) {
			return vssh.CommandResponse{ExitCode: 0, Stdout: []byte("pong")}, nil
		},
	}

	response, err := vssh.Run(context.Background(), provider, vssh.CommandRequest{Command: "printf", Args: []string{"pong"}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if string(response.Stdout) != "pong" {
		t.Fatalf("Run response = %+v", response)
	}
}

func TestListFacade(t *testing.T) {
	provider := providerFunc{
		list: func(ctx context.Context, request vssh.ListRequest) (vssh.ListResponse, error) {
			return vssh.ListResponse{Entries: []vssh.Entry{{Name: "file.txt", Path: "/pub/file.txt", Type: vssh.EntryTypeFile}}}, nil
		},
	}

	response, err := vssh.List(context.Background(), provider, vssh.ListRequest{RemoteDir: "/pub"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(response.Entries) != 1 || response.Entries[0].Name != "file.txt" {
		t.Fatalf("List response = %+v", response)
	}
}

func TestDownloadFacade(t *testing.T) {
	provider := providerFunc{
		download: func(ctx context.Context, request vssh.DownloadRequest) (vssh.DownloadResponse, error) {
			return vssh.DownloadResponse{RemotePath: request.RemotePath, Content: []byte("hello"), Size: 5}, nil
		},
	}

	response, err := vssh.Download(context.Background(), provider, vssh.DownloadRequest{RemotePath: "/pub/file.txt"})
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if string(response.Content) != "hello" {
		t.Fatalf("Download response = %+v", response)
	}
}

func TestUploadFacade(t *testing.T) {
	provider := providerFunc{
		upload: func(ctx context.Context, request vssh.UploadRequest) (vssh.UploadResponse, error) {
			return vssh.UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
		},
	}

	response, err := vssh.Upload(context.Background(), provider, vssh.UploadRequest{RemotePath: "/pub/file.txt", Content: []byte("hello")})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if response.Size != 5 {
		t.Fatalf("Upload response = %+v", response)
	}
}

func TestFacadeExposesErrors(t *testing.T) {
	_, err := vssh.Run(context.Background(), nil, vssh.CommandRequest{Command: "printf"})
	if !errors.Is(err, vssh.ErrMissingProvider) {
		t.Fatalf("Run error = %v, want ErrMissingProvider", err)
	}

	_, err = vssh.Run(context.Background(), providerFunc{}, vssh.CommandRequest{Command: "printf", MaxOutputBytes: -1})
	if !errors.Is(err, vssh.ErrInvalidCommandRequest) {
		t.Fatalf("Run error = %v, want ErrInvalidCommandRequest", err)
	}

	_, err = vssh.Upload(context.Background(), providerFunc{}, vssh.UploadRequest{RemotePath: "/pub/file.txt", Content: []byte("hello"), MaxBytes: 4})
	if !errors.Is(err, vssh.ErrTransferLimitExceeded) {
		t.Fatalf("Upload error = %v, want ErrTransferLimitExceeded", err)
	}
}

func TestFacadeProviderErrorContract(t *testing.T) {
	cause := errors.New("backend unavailable")
	secretPath := "/secret/private.txt"
	_, err := vssh.Download(context.Background(), providerFunc{
		download: func(context.Context, vssh.DownloadRequest) (vssh.DownloadResponse, error) {
			return vssh.DownloadResponse{}, cause
		},
	}, vssh.DownloadRequest{RemotePath: secretPath})
	if !errors.Is(err, cause) {
		t.Fatalf("Download error = %v, want provider cause", err)
	}
	if !errors.Is(err, knifer.ErrCodeProviderFailure) {
		t.Fatalf("Download error = %v, want ErrCodeProviderFailure", err)
	}
	if strings.Contains(err.Error(), secretPath) || strings.Contains(err.Error(), "secret") {
		t.Fatalf("Download error leaked remote path or secret: %v", err)
	}
}
