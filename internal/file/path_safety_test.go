package file

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestIsLocalPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "simple file", path: "data.txt", want: true},
		{name: "nested file", path: "reports/2026/data.txt", want: true},
		{name: "dot cleaned", path: "./reports/../data.txt", want: true},
		{name: "blank", path: " ", want: false},
		{name: "current directory", path: ".", want: false},
		{name: "parent directory", path: "..", want: false},
		{name: "parent escape", path: "../secret.txt", want: false},
		{name: "nested parent escape", path: "safe/../../secret.txt", want: false},
		{name: "unix absolute", path: "/tmp/secret.txt", want: false},
		{name: "windows drive", path: `C:\temp\secret.txt`, want: false},
		{name: "windows relative segments", path: `safe\secret.txt`, want: false},
		{name: "windows unc", path: `\\server\share\secret.txt`, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocalPath(tt.path); got != tt.want {
				t.Fatalf("IsLocalPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestSafeJoin(t *testing.T) {
	root := t.TempDir()
	got, err := SafeJoin(root, "nested/data.txt")
	if err != nil {
		t.Fatalf("SafeJoin valid path: %v", err)
	}
	want := filepath.Join(root, "nested", "data.txt")
	if got != want {
		t.Fatalf("SafeJoin valid = %q, want %q", got, want)
	}

	for _, path := range []string{"../escape.txt", "/tmp/escape.txt", `C:\temp\escape.txt`, `nested\escape.txt`} {
		t.Run(path, func(t *testing.T) {
			_, err := SafeJoin(root, path)
			assertFileCode(t, err, knifer.ErrCodeInvalidInput)
		})
	}
}

func TestSafeJoinRejectsExistingSymlinkParentEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skip("symlink creation is not permitted")
		}
		t.Fatalf("create symlink: %v", err)
	}
	_, err := SafeJoin(root, "link/escape.txt")
	assertFileCode(t, err, knifer.ErrCodeInvalidInput)
}
