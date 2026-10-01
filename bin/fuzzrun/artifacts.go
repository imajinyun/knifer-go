package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func command(ctx context.Context, dir, exe string, args []string, log io.Writer) (string, int, error) {
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	var output bytes.Buffer
	cmd.Stdout = log
	cmd.Stderr = log
	if len(args) > 0 && !strings.HasPrefix(args[0], "-test.") {
		cmd.Stdout = io.MultiWriter(&output, log)
	}
	err := cmd.Run()
	if ctx.Err() != nil {
		return output.String(), 124, ctx.Err()
	}
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return output.String(), exit.ExitCode(), err
		}
		return output.String(), 1, err
	}
	return output.String(), 0, nil
}

func revision(root string) (string, string, bool) {
	git := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		data, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}
	commit := git("rev-parse", "HEAD")
	return commit, git("rev-parse", "HEAD^{tree}"), commit == "" || git("status", "--porcelain") != ""
}

func writeSummary(s Summary) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.Directory, "summary.json")
	if err := os.WriteFile(path+".tmp", append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(path+".tmp", path)
}

func fileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyTestdata(source, destination string) error {
	path := filepath.Join(source, "testdata")
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	return os.CopyFS(filepath.Join(destination, "testdata"), os.DirFS(path))
}

func collectCorpus(dir, runDir string, prior error) ([]Input, error) {
	inputs := []Input{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("corpus must contain regular files: %s", path)
		}
		hash, err := fileHash(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(runDir, path)
		if err != nil {
			return err
		}
		inputs = append(inputs, Input{Path: filepath.ToSlash(rel), SHA256: hash})
		return nil
	})
	return inputs, errors.Join(prior, err)
}
