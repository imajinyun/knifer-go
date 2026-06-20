package ftp

import (
	"context"
	"testing"
)

type benchmarkProvider struct{}

func (benchmarkProvider) List(ctx context.Context, request ListRequest) (ListResponse, error) {
	return ListResponse{Entries: []Entry{{Name: "file.txt", Path: "/file.txt", Type: EntryTypeFile}}}, nil
}

func (benchmarkProvider) Download(ctx context.Context, request DownloadRequest) (DownloadResponse, error) {
	return DownloadResponse{RemotePath: request.RemotePath, Content: []byte("hello"), Size: 5}, nil
}

func (benchmarkProvider) Upload(ctx context.Context, request UploadRequest) (UploadResponse, error) {
	return UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
}

func BenchmarkClientList(b *testing.B) {
	b.ReportAllocs()
	client := New(WithProvider(benchmarkProvider{}))
	request := ListRequest{RemoteDir: "/pub"}
	for b.Loop() {
		if _, err := client.List(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientDownload(b *testing.B) {
	b.ReportAllocs()
	client := New(WithProvider(benchmarkProvider{}))
	request := DownloadRequest{RemotePath: "/file.txt", MaxBytes: 1024}
	for b.Loop() {
		if _, err := client.Download(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientUpload(b *testing.B) {
	b.ReportAllocs()
	client := New(WithProvider(benchmarkProvider{}))
	request := UploadRequest{RemotePath: "/file.txt", Content: []byte("hello"), MaxBytes: 1024}
	for b.Loop() {
		if _, err := client.Upload(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}
