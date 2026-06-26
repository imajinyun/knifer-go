package conf

import (
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestLoadWithOptionsIncludeRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "outside.setting")
	confDir := filepath.Join(dir, "conf")
	if err := os.Mkdir(confDir, 0o755); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(confDir, "main.setting")
	if err := os.WriteFile(outside, []byte("secret=outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("include=../outside.setting\nname=main"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadWithOptions(main, LoadOptions{AllowInclude: true})
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestLoadWithOptionsIncludeRejectsAbsolutePathByDefault(t *testing.T) {
	dir := t.TempDir()
	common := filepath.Join(dir, "common.setting")
	main := filepath.Join(dir, "main.setting")
	if err := os.WriteFile(common, []byte("name=common"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("include="+common+"\nname=main"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadWithOptions(main, LoadOptions{AllowInclude: true})
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestLoadWithOptionsIncludeRootAllowsConfiguredDirectory(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	serviceDir := filepath.Join(root, "service")
	commonDir := filepath.Join(root, "common")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(commonDir, 0o755); err != nil {
		t.Fatal(err)
	}
	common := filepath.Join(commonDir, "base.setting")
	main := filepath.Join(serviceDir, "main.setting")
	if err := os.WriteFile(common, []byte("name=common"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("include=../common/base.setting\nmode=service"), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := LoadWithOptions(main, LoadOptions{AllowInclude: true, IncludeRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("name"); got != "common" {
		t.Fatalf("included name = %q", got)
	}
	if got := c.Get("mode"); got != "service" {
		t.Fatalf("main mode = %q", got)
	}
}
