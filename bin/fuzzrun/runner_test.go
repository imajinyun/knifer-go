package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	for _, name := range []string{"passed", "discovery_error", "no_targets", "failure_continues", "timeout", "build_error", "run_timeout"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			out := filepath.Join(t.TempDir(), "artifacts")
			calls := []string{}
			cfg := Config{Root: root, Go: "fixture-go", Artifacts: out, Packages: []string{"./..."}, FuzzTime: time.Second, MinimizeTime: time.Second, RunTimeout: time.Minute, Parallel: 2}
			if name == "run_timeout" {
				cfg.RunTimeout = time.Nanosecond
			}
			cfg.execute = func(ctx context.Context, dir, exe string, args []string, log io.Writer) (string, int, error) {
				if exe == "fixture-go" {
					if args[0] == "env" {
						return `{"GOVERSION":"go1.27.1","GOTOOLCHAIN":"local"}`, 0, nil
					}
					if args[0] == "list" {
						return root, 0, nil
					}
					if args[1] == "-list" {
						if name == "discovery_error" {
							return "", 17, errors.New("discovery failed")
						}
						if name == "no_targets" {
							return "ok example/p 0.01s\n", 0, nil
						}
						return "FuzzFirst\nFuzzSecond\nok example/p 0.01s\n", 0, nil
					}
					if name == "build_error" {
						return "", 1, errors.New("compile failed")
					}
					for i, arg := range args {
						if arg == "-o" {
							if err := os.WriteFile(args[i+1], []byte("fixture binary"), 0o700); err != nil {
								t.Fatal(err)
							}
						}
					}
					return "", 0, nil
				}
				target := "FuzzSecond"
				for _, arg := range args {
					if strings.Contains(arg, "FuzzFirst") {
						target = "FuzzFirst"
					}
				}
				calls = append(calls, target)
				if target == "FuzzFirst" && name == "failure_continues" {
					corpus := filepath.Join(dir, "testdata/fuzz", target)
					if err := os.MkdirAll(corpus, 0o700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(corpus, "input"), []byte("go test fuzz v1\nstring(\"bad\")\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					_, _ = io.WriteString(log, "deliberate fixture failure\n")
					return "", 1, errors.New("failed")
				}
				if name == "timeout" {
					return "", 124, context.DeadlineExceeded
				}
				return "", 0, nil
			}
			s, err := Run(context.Background(), cfg)
			if (err == nil) != (name == "passed") {
				t.Fatalf("Run error=%v status=%s", err, s.Status)
			}
			if _, err := os.Stat(filepath.Join(s.Directory, "summary.json")); err != nil {
				t.Fatal(err)
			}
			if name == "failure_continues" {
				if len(calls) != 2 || len(s.Targets[0].Corpus) != 1 || s.Targets[1].Status != "passed" {
					t.Fatalf("results=%+v calls=%v", s.Targets, calls)
				}
				if _, err := os.Stat(filepath.Join(root, "testdata")); !os.IsNotExist(err) {
					t.Fatal("source corpus was modified")
				}
			}
			if name == "timeout" && s.Targets[0].Status != "timeout" {
				t.Fatalf("status=%s", s.Targets[0].Status)
			}
		})
	}
}
