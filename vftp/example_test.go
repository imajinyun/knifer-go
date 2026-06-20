package vftp_test

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vftp"
)

type exampleProvider struct{}

func (exampleProvider) List(ctx context.Context, request vftp.ListRequest) (vftp.ListResponse, error) {
	return vftp.ListResponse{Entries: []vftp.Entry{{Name: "report.csv", Path: request.RemoteDir + "/report.csv", Type: vftp.EntryTypeFile}}}, nil
}

func (exampleProvider) Download(ctx context.Context, request vftp.DownloadRequest) (vftp.DownloadResponse, error) {
	return vftp.DownloadResponse{RemotePath: request.RemotePath, Content: []byte("id,total\n1,42\n"), Size: 14}, nil
}

func (exampleProvider) Upload(ctx context.Context, request vftp.UploadRequest) (vftp.UploadResponse, error) {
	return vftp.UploadResponse{RemotePath: request.RemotePath, Size: int64(len(request.Content))}, nil
}

func ExampleNew() {
	client := vftp.New(vftp.WithProvider(exampleProvider{}))
	response, _ := client.List(context.Background(), vftp.ListRequest{RemoteDir: "/pub"})
	fmt.Println(response.Entries[0].Name)
	// Output: report.csv
}

func ExampleWithProvider() {
	client := vftp.New(vftp.WithProvider(exampleProvider{}))
	response, _ := client.Download(context.Background(), vftp.DownloadRequest{RemotePath: "/pub/report.csv"})
	fmt.Println(response.Size)
	// Output: 14
}

func ExampleList() {
	response, _ := vftp.List(context.Background(), exampleProvider{}, vftp.ListRequest{RemoteDir: "/pub"})
	fmt.Println(len(response.Entries), response.Entries[0].Type)
	// Output: 1 file
}

func ExampleDownload() {
	response, _ := vftp.Download(context.Background(), exampleProvider{}, vftp.DownloadRequest{RemotePath: "/pub/report.csv", MaxBytes: 1024})
	fmt.Println(response.RemotePath, string(response.Content[:2]))
	// Output: /pub/report.csv id
}

func ExampleUpload() {
	response, _ := vftp.Upload(context.Background(), exampleProvider{}, vftp.UploadRequest{RemotePath: "/pub/out.csv", Content: []byte("ok\n"), MaxBytes: 1024})
	fmt.Println(response.RemotePath, response.Size)
	// Output: /pub/out.csv 3
}
