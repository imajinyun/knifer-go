package ssh

import (
	"context"
	"testing"
)

type benchmarkProvider struct{}

func (benchmarkProvider) Run(ctx context.Context, request CommandRequest) (CommandResponse, error) {
	return CommandResponse{ExitCode: 0, Stdout: []byte("ok")}, nil
}

func (benchmarkProvider) List(ctx context.Context, request ListRequest) (ListResponse, error) {
	return ListResponse{Entries: []Entry{{Name: "file.txt", Path: request.RemoteDir + "/file.txt", Type: EntryTypeFile}}}, nil
}

func (benchmarkProvider) Download(ctx context.Context, request DownloadRequest) (DownloadResponse, error) {
	return DownloadResponse{RemotePath: request.RemotePath, Content: []byte("hello"), Size: 5}, nil
}

func (benchmarkProvider) Upload(ctx context.Context, request UploadRequest) (UploadResponse, error) {
	return UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
}

func BenchmarkClientRun(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := CommandRequest{Command: "printf", Args: []string{"ok"}, MaxOutputBytes: 1024}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.Run(context.Background(), request)
	}
}

func BenchmarkClientList(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := ListRequest{RemoteDir: "/pub"}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.List(context.Background(), request)
	}
}

func BenchmarkClientDownload(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := DownloadRequest{RemotePath: "/pub/file.txt", MaxBytes: 1024}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.Download(context.Background(), request)
	}
}

func BenchmarkClientUpload(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := UploadRequest{RemotePath: "/pub/file.txt", Content: []byte("hello"), MaxBytes: 1024}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.Upload(context.Background(), request)
	}
}
