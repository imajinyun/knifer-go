package ftp

import (
	"errors"
	"reflect"
	"testing"
)

func TestListRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request ListRequest
		wantErr error
	}{
		{name: "blank path", request: ListRequest{RemoteDir: " "}, wantErr: ErrInvalidListRequest},
		{name: "nul path", request: ListRequest{RemoteDir: "/tmp\x00bad"}, wantErr: ErrInvalidListRequest},
		{name: "root path", request: ListRequest{RemoteDir: "/"}},
		{name: "nested path", request: ListRequest{RemoteDir: "/pub/data"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDownloadRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request DownloadRequest
		wantErr error
	}{
		{name: "blank path", request: DownloadRequest{RemotePath: ""}, wantErr: ErrInvalidDownloadRequest},
		{name: "negative max bytes", request: DownloadRequest{RemotePath: "/file.txt", MaxBytes: -1}, wantErr: ErrInvalidDownloadRequest},
		{name: "unlimited", request: DownloadRequest{RemotePath: "/file.txt"}},
		{name: "bounded", request: DownloadRequest{RemotePath: "/file.txt", MaxBytes: 1024}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request UploadRequest
		wantErr error
	}{
		{name: "blank path", request: UploadRequest{RemotePath: ""}, wantErr: ErrInvalidUploadRequest},
		{name: "negative max bytes", request: UploadRequest{RemotePath: "/file.txt", MaxBytes: -1}, wantErr: ErrInvalidUploadRequest},
		{name: "content exceeds max bytes", request: UploadRequest{RemotePath: "/file.txt", Content: []byte("hello"), MaxBytes: 4}, wantErr: ErrTransferLimitExceeded},
		{name: "valid bounded", request: UploadRequest{RemotePath: "/file.txt", Content: []byte("hello"), MaxBytes: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListResponseCloneDefensivelyCopiesEntriesAndMetadata(t *testing.T) {
	response := ListResponse{
		Entries:  []Entry{{Name: "file.txt", Path: "/file.txt", Type: EntryTypeFile, Metadata: map[string]string{"owner": "one"}}},
		Metadata: map[string]string{"trace": "one"},
	}
	clone := response.Clone()
	response.Entries[0].Name = "changed"
	response.Entries[0].Metadata["owner"] = "changed"
	response.Metadata["trace"] = "changed"

	if clone.Entries[0].Name != "file.txt" || clone.Entries[0].Metadata["owner"] != "one" {
		t.Fatalf("cloned entries changed: %+v", clone.Entries)
	}
	if clone.Metadata["trace"] != "one" {
		t.Fatalf("cloned metadata changed: %#v", clone.Metadata)
	}
}

func TestTransferCloneDefensivelyCopiesContentAndMetadata(t *testing.T) {
	download := DownloadResponse{RemotePath: "/file.txt", Content: []byte("hello"), Size: 5, Metadata: map[string]string{"trace": "one"}}
	downloadClone := download.Clone()
	download.Content[0] = 'H'
	download.Metadata["trace"] = "changed"
	if !reflect.DeepEqual(downloadClone.Content, []byte("hello")) || downloadClone.Metadata["trace"] != "one" {
		t.Fatalf("download clone changed: %+v", downloadClone)
	}

	upload := UploadRequest{RemotePath: "/file.txt", Content: []byte("hello"), MaxBytes: 10, Metadata: map[string]string{"trace": "one"}}
	uploadClone := upload.Clone()
	upload.Content[0] = 'H'
	upload.Metadata["trace"] = "changed"
	if !reflect.DeepEqual(uploadClone.Content, []byte("hello")) || uploadClone.Metadata["trace"] != "one" {
		t.Fatalf("upload clone changed: %+v", uploadClone)
	}
}
