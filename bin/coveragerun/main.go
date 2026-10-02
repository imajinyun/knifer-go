package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/cievidence"
	"github.com/imajinyun/knifer-go/bin/internal/coverageprofile"
	"github.com/imajinyun/knifer-go/bin/internal/sourceidentity"
)

func main() {
	root := flag.String("root", ".", "repository root")
	goCmd := flag.String("go", "go", "Go executable")
	out := flag.String("out", "/tmp/knifer-go-coverage.out", "normalized coverage output")
	flag.Parse()
	pkgs := flag.Args()
	if len(pkgs) == 0 {
		pkgs = []string{"./..."}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, *root, *goCmd, *out, pkgs)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage run:", err)
		os.Exit(1)
	}
}

func run(parent context.Context, root, goCmd, out string, pkgs []string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return err
	}
	for _, path := range []string{out, out + ".meta.json"} {
		if rel, err := filepath.Rel(root, path); err != nil {
			return err
		} else if filepath.IsLocal(rel) {
			cmd := exec.Command("git", "-C", root, "check-ignore", "--no-index", "-q", "--", rel)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("coverage artifacts inside the workspace must be ignored: %s", rel)
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
		return err
	}
	before, err := sourceidentity.Capture(root)
	if err != nil {
		return err
	}
	for _, path := range []string{out, out + ".meta.json"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	raw, err := os.CreateTemp("", "knifer-coverage-*.out")
	if err != nil {
		return err
	}
	rawPath := raw.Name()
	if err := raw.Close(); err != nil {
		return err
	}
	defer func() { _ = os.Remove(rawPath) }()
	ctx, cancel := context.WithTimeout(parent, 15*time.Minute)
	defer cancel()
	env := append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	version := exec.CommandContext(ctx, goCmd, "env", "GOVERSION")
	version.Dir = root
	version.Env = env
	goVersion, err := version.Output()
	if err != nil {
		return err
	}
	args := append([]string{"test", "-mod=readonly", "-race", "-shuffle=on", "-coverpkg=./...", "-coverprofile=" + rawPath}, pkgs...)
	cmd := exec.CommandContext(ctx, goCmd, args...)
	cmd.Dir = root
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	testErr := cmd.Run()
	exit := 0
	if testErr != nil {
		exit = 1
		var e *exec.ExitError
		if errors.As(testErr, &e) {
			exit = e.ExitCode()
		}
	}
	evidence := coverageprofile.Evidence{Schema: 1, Before: before, GoVersion: strings.TrimSpace(string(goVersion)), Command: append([]string{goCmd}, args...), Packages: pkgs, TestsExitCode: exit, RunID: os.Getenv("GITHUB_RUN_ID"), Attempt: os.Getenv("GITHUB_RUN_ATTEMPT"), CreatedAt: time.Now().UTC()}
	evidence.After, err = sourceidentity.Capture(root)
	if err == nil && evidence.Before != evidence.After {
		err = errors.New("repository inputs changed during coverage execution")
	}
	file, openErr := os.Open(rawPath)
	var blocks []coverageprofile.Block
	if openErr == nil {
		blocks, openErr = coverageprofile.Parse(file)
		_ = file.Close()
	}
	if openErr == nil {
		var output *os.File
		output, openErr = os.Create(out)
		if openErr == nil {
			openErr = coverageprofile.Write(output, blocks)
			openErr = errors.Join(openErr, output.Close())
		}
		if openErr == nil {
			evidence.ProfileSHA256, openErr = coverageprofile.HashFile(out)
		}
	}
	evidence.Complete = testErr == nil && err == nil && openErr == nil && evidence.Before == evidence.After
	if saveErr := cievidence.WriteJSON(out+".meta.json", evidence); saveErr != nil {
		return saveErr
	}
	if !evidence.Complete {
		return fmt.Errorf("incomplete test execution or changed source: %w", errors.Join(testErr, err, openErr))
	}
	data, _ := json.Marshal(map[string]any{"coverage": out, "commit": evidence.Before.Commit, "source_sha256": evidence.Before.SourceSHA256, "go_version": evidence.GoVersion})
	fmt.Println(string(data))
	return nil
}
