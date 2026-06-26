package vftp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/vftp"
)

type providerFunc struct {
	list     func(context.Context, vftp.ListRequest) (vftp.ListResponse, error)
	download func(context.Context, vftp.DownloadRequest) (vftp.DownloadResponse, error)
	upload   func(context.Context, vftp.UploadRequest) (vftp.UploadResponse, error)
}

func (p providerFunc) List(ctx context.Context, request vftp.ListRequest) (vftp.ListResponse, error) {
	return p.list(ctx, request)
}

func (p providerFunc) Download(ctx context.Context, request vftp.DownloadRequest) (vftp.DownloadResponse, error) {
	return p.download(ctx, request)
}

func (p providerFunc) Upload(ctx context.Context, request vftp.UploadRequest) (vftp.UploadResponse, error) {
	return p.upload(ctx, request)
}

func TestListFacade(t *testing.T) {
	provider := providerFunc{
		list: func(ctx context.Context, request vftp.ListRequest) (vftp.ListResponse, error) {
			return vftp.ListResponse{Entries: []vftp.Entry{{Name: "file.txt", Path: "/pub/file.txt", Type: vftp.EntryTypeFile}}}, nil
		},
	}

	response, err := vftp.List(context.Background(), provider, vftp.ListRequest{RemoteDir: "/pub"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(response.Entries) != 1 || response.Entries[0].Name != "file.txt" {
		t.Fatalf("List response = %+v", response)
	}
}

func TestDownloadFacade(t *testing.T) {
	provider := providerFunc{
		download: func(ctx context.Context, request vftp.DownloadRequest) (vftp.DownloadResponse, error) {
			return vftp.DownloadResponse{RemotePath: request.RemotePath, Content: []byte("hello"), Size: 5}, nil
		},
	}

	response, err := vftp.Download(context.Background(), provider, vftp.DownloadRequest{RemotePath: "/pub/file.txt"})
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if string(response.Content) != "hello" {
		t.Fatalf("Download response = %+v", response)
	}
}

func TestUploadFacade(t *testing.T) {
	provider := providerFunc{
		upload: func(ctx context.Context, request vftp.UploadRequest) (vftp.UploadResponse, error) {
			return vftp.UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
		},
	}

	response, err := vftp.Upload(context.Background(), provider, vftp.UploadRequest{RemotePath: "/pub/file.txt", Content: []byte("hello")})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if response.Size != 5 {
		t.Fatalf("Upload response = %+v", response)
	}
}

func TestFacadeExposesErrors(t *testing.T) {
	_, err := vftp.Download(context.Background(), nil, vftp.DownloadRequest{RemotePath: "/pub/file.txt"})
	if !errors.Is(err, vftp.ErrMissingProvider) {
		t.Fatalf("Download error = %v, want ErrMissingProvider", err)
	}

	_, err = vftp.Upload(context.Background(), providerFunc{}, vftp.UploadRequest{RemotePath: "/pub/file.txt", Content: []byte("hello"), MaxBytes: 4})
	if !errors.Is(err, vftp.ErrTransferLimitExceeded) {
		t.Fatalf("Upload error = %v, want ErrTransferLimitExceeded", err)
	}
}
