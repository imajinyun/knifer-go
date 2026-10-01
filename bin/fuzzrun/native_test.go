package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNativeFuzzArtifacts(t *testing.T) {
	root := t.TempDir()
	artifacts := filepath.Join(t.TempDir(), "artifacts")
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/fuzzfixture\ngo 1.26.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	const source = `package fuzzfixture
import "testing"
func FuzzFailure(f *testing.F) {
 f.Add("seed")
 f.Fuzz(func(t *testing.T, value string) { if value != "seed" { t.Fatal("intentional fixture failure") } })
}
func FuzzFollowing(f *testing.F) { f.Add("seed"); f.Fuzz(func(t *testing.T, value string) {}) }
`
	if err := os.WriteFile(filepath.Join(root, "fuzz_test.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Root: root, Go: "go", Artifacts: artifacts, Packages: []string{"./..."}, FuzzTime: time.Second, MinimizeTime: time.Second, RunTimeout: time.Minute, Parallel: 2}
	s, err := Run(context.Background(), cfg)
	if err == nil || len(s.Targets) != 2 {
		t.Fatalf("expected deliberate failure: %v %+v", err, s)
	}
	if s.Targets[0].Status != "failed" || len(s.Targets[0].Corpus) == 0 || s.Targets[1].Status != "passed" {
		t.Fatalf("targets=%+v", s.Targets)
	}
	if _, err := os.Stat(filepath.Join(root, "testdata")); !os.IsNotExist(err) {
		t.Fatal("failure wrote into source corpus")
	}
	input := s.Targets[0].Corpus[0]
	bytes, err := os.ReadFile(filepath.Join(s.Directory, filepath.FromSlash(input.Path)))
	if err != nil {
		t.Fatal(err)
	}
	corpus := filepath.Join(root, "testdata/fuzz/FuzzFailure")
	if err := os.MkdirAll(corpus, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(corpus, filepath.Base(input.Path)), bytes, 0o600); err != nil {
		t.Fatal(err)
	}
	// Deliberately promote the captured fixture input, then prove it reproduces.
	_, code, err := command(context.Background(), root, "go", []string{"test", "-run", "^FuzzFailure/" + filepath.Base(input.Path) + "$", "."}, os.Stdout)
	if err == nil || code == 0 {
		t.Fatal("captured failure input did not reproduce")
	}
}
