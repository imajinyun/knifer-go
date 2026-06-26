package vssh_test

import (
	"context"
	"fmt"

	"github.com/imajinyun/knifer-go/vssh"
)

type exampleProvider struct{}

func (exampleProvider) Run(ctx context.Context, request vssh.CommandRequest) (vssh.CommandResponse, error) {
	return vssh.CommandResponse{ExitCode: 0, Stdout: []byte("hello gopher")}, nil
}

func (exampleProvider) List(ctx context.Context, request vssh.ListRequest) (vssh.ListResponse, error) {
	return vssh.ListResponse{Entries: []vssh.Entry{{Name: "report.csv", Path: request.RemoteDir + "/report.csv", Type: vssh.EntryTypeFile}}}, nil
}

func (exampleProvider) Download(ctx context.Context, request vssh.DownloadRequest) (vssh.DownloadResponse, error) {
	return vssh.DownloadResponse{RemotePath: request.RemotePath, Content: []byte("id,total\n1,42\n"), Size: 14}, nil
}

func (exampleProvider) Upload(ctx context.Context, request vssh.UploadRequest) (vssh.UploadResponse, error) {
	return vssh.UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
}

func ExampleNew() {
	client := vssh.New(vssh.WithProvider(exampleProvider{}))
	response, _ := client.Run(context.Background(), vssh.CommandRequest{Command: "printf"})
	fmt.Println(string(response.Stdout))
	// Output: hello gopher
}

func ExampleWithProvider() {
	client := vssh.New(vssh.WithProvider(exampleProvider{}))
	response, _ := client.Download(context.Background(), vssh.DownloadRequest{RemotePath: "/pub/report.csv"})
	fmt.Println(response.Size)
	// Output: 14
}

func ExampleRun() {
	response, _ := vssh.Run(context.Background(), exampleProvider{}, vssh.CommandRequest{Command: "printf", Args: []string{"hello"}, MaxOutputBytes: 1024})
	fmt.Println(response.ExitCode, string(response.Stdout))
	// Output: 0 hello gopher
}

func ExampleList() {
	response, _ := vssh.List(context.Background(), exampleProvider{}, vssh.ListRequest{RemoteDir: "/pub"})
	fmt.Println(len(response.Entries), response.Entries[0].Type)
	// Output: 1 file
}

func ExampleDownload() {
	response, _ := vssh.Download(context.Background(), exampleProvider{}, vssh.DownloadRequest{RemotePath: "/pub/report.csv", MaxBytes: 1024})
	fmt.Println(response.RemotePath, string(response.Content[:2]))
	// Output: /pub/report.csv id
}

func ExampleUpload() {
	response, _ := vssh.Upload(context.Background(), exampleProvider{}, vssh.UploadRequest{RemotePath: "/pub/out.csv", Content: []byte("ok\n"), MaxBytes: 1024})
	fmt.Println(response.RemotePath, response.Size)
	// Output: /pub/out.csv 3
}
