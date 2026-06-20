package ssh

import (
	"errors"
	"testing"
)

func TestCommandRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request CommandRequest
		wantErr error
	}{
		{name: "valid command", request: CommandRequest{Command: "ls", Args: []string{"-la"}}},
		{name: "missing command", request: CommandRequest{}, wantErr: ErrInvalidCommandRequest},
		{name: "command nul", request: CommandRequest{Command: "ls\x00"}, wantErr: ErrInvalidCommandRequest},
		{name: "arg nul", request: CommandRequest{Command: "ls", Args: []string{"bad\x00"}}, wantErr: ErrInvalidCommandRequest},
		{name: "negative output limit", request: CommandRequest{Command: "ls", MaxOutputBytes: -1}, wantErr: ErrInvalidCommandRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request ListRequest
		wantErr error
	}{
		{name: "root", request: ListRequest{RemoteDir: "/"}},
		{name: "missing path", request: ListRequest{}, wantErr: ErrInvalidListRequest},
		{name: "nul path", request: ListRequest{RemoteDir: "/tmp\x00"}, wantErr: ErrInvalidListRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
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
		{name: "valid file", request: DownloadRequest{RemotePath: "/tmp/file.txt", MaxBytes: 1}},
		{name: "missing path", request: DownloadRequest{}, wantErr: ErrInvalidDownloadRequest},
		{name: "nul path", request: DownloadRequest{RemotePath: "/tmp\x00"}, wantErr: ErrInvalidDownloadRequest},
		{name: "negative limit", request: DownloadRequest{RemotePath: "/tmp/file.txt", MaxBytes: -1}, wantErr: ErrInvalidDownloadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
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
		{name: "valid file", request: UploadRequest{RemotePath: "/tmp/file.txt", Content: []byte("ok"), MaxBytes: 2}},
		{name: "missing path", request: UploadRequest{}, wantErr: ErrInvalidUploadRequest},
		{name: "nul path", request: UploadRequest{RemotePath: "/tmp\x00"}, wantErr: ErrInvalidUploadRequest},
		{name: "negative limit", request: UploadRequest{RemotePath: "/tmp/file.txt", MaxBytes: -1}, wantErr: ErrInvalidUploadRequest},
		{name: "content exceeds limit", request: UploadRequest{RemotePath: "/tmp/file.txt", Content: []byte("hello"), MaxBytes: 4}, wantErr: ErrTransferLimitExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloneCopiesMutableFields(t *testing.T) {
	cmd := CommandRequest{Command: "printf", Args: []string{"hello"}, Metadata: map[string]string{"trace": "one"}}
	cmdClone := cmd.Clone()
	cmd.Args[0] = "changed"
	cmd.Metadata["trace"] = "changed"
	if cmdClone.Args[0] != "hello" || cmdClone.Metadata["trace"] != "one" {
		t.Fatalf("command request clone was mutated: %+v", cmdClone)
	}

	cmdResp := CommandResponse{Stdout: []byte("out"), Stderr: []byte("err"), Metadata: map[string]string{"trace": "one"}}
	cmdRespClone := cmdResp.Clone()
	cmdResp.Stdout[0] = 'O'
	cmdResp.Stderr[0] = 'E'
	cmdResp.Metadata["trace"] = "changed"
	if string(cmdRespClone.Stdout) != "out" || string(cmdRespClone.Stderr) != "err" || cmdRespClone.Metadata["trace"] != "one" {
		t.Fatalf("command response clone was mutated: %+v", cmdRespClone)
	}

	list := ListResponse{Entries: []Entry{{Name: "file", Metadata: map[string]string{"mode": "0644"}}}, Metadata: map[string]string{"trace": "one"}}
	listClone := list.Clone()
	list.Entries[0].Name = "changed"
	list.Entries[0].Metadata["mode"] = "0600"
	list.Metadata["trace"] = "changed"
	if listClone.Entries[0].Name != "file" || listClone.Entries[0].Metadata["mode"] != "0644" || listClone.Metadata["trace"] != "one" {
		t.Fatalf("list response clone was mutated: %+v", listClone)
	}

	download := DownloadResponse{Content: []byte("hello"), Metadata: map[string]string{"trace": "one"}}
	downloadClone := download.Clone()
	download.Content[0] = 'H'
	download.Metadata["trace"] = "changed"
	if string(downloadClone.Content) != "hello" || downloadClone.Metadata["trace"] != "one" {
		t.Fatalf("download response clone was mutated: %+v", downloadClone)
	}

	upload := UploadRequest{RemotePath: "/tmp/file.txt", Content: []byte("hello"), Metadata: map[string]string{"trace": "one"}}
	uploadClone := upload.Clone()
	upload.Content[0] = 'H'
	upload.Metadata["trace"] = "changed"
	if string(uploadClone.Content) != "hello" || uploadClone.Metadata["trace"] != "one" {
		t.Fatalf("upload request clone was mutated: %+v", uploadClone)
	}
}
