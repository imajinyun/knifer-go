package vfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFacadePathSafety(t *testing.T) {
	if !IsLocalPath("nested/data.txt") {
		t.Fatal("IsLocalPath rejected nested local path")
	}
	for _, path := range []string{"../escape.txt", "/tmp/escape.txt", `C:\temp\escape.txt`, `nested\escape.txt`} {
		t.Run(path, func(t *testing.T) {
			if IsLocalPath(path) {
				t.Fatalf("IsLocalPath(%q) = true, want false", path)
			}
		})
	}

	root := t.TempDir()
	got, err := SafeJoin(root, "nested/data.txt")
	if err != nil {
		t.Fatalf("SafeJoin valid path: %v", err)
	}
	if want := filepath.Join(root, "nested", "data.txt"); got != want {
		t.Fatalf("SafeJoin = %q, want %q", got, want)
	}
	if _, err := SafeJoin(root, "../escape.txt"); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("SafeJoin escape err = %v, want ErrCodeInvalidInput", err)
	}
}

func TestFacadeSafeJoinRejectsSymlinkParentEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skip("symlink creation is not permitted")
		}
		t.Fatalf("create symlink: %v", err)
	}
	if _, err := SafeJoin(root, "link/escape.txt"); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("SafeJoin symlink escape err = %v, want ErrCodeInvalidInput", err)
	}
}
