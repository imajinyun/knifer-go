package cievidence

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/sourceidentity"
)

// Capture binds a clean checkout to its immutable tree, dependencies and CI run.
func Capture(root string) (Candidate, error) {
	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		data, err := cmd.Output()
		return strings.TrimSpace(string(data)), err
	}
	c := Candidate{RunID: os.Getenv("GITHUB_RUN_ID"), Attempt: os.Getenv("GITHUB_RUN_ATTEMPT"), Base: os.Getenv("AGENT_CHANGE_BASE_REF")}
	var err error
	if c.Commit, err = git("rev-parse", "HEAD"); err != nil {
		return c, err
	}
	if c.Tree, err = git("rev-parse", "HEAD^{tree}"); err != nil {
		return c, err
	}
	if sha := os.Getenv("GITHUB_SHA"); sha != "" && sha != c.Commit {
		return c, fmt.Errorf("checkout HEAD does not match GITHUB_SHA")
	}
	if c.RunID == "" || c.Attempt == "" {
		return c, fmt.Errorf("GITHUB_RUN_ID and GITHUB_RUN_ATTEMPT are required")
	}
	if status, err := git("status", "--porcelain", "--untracked-files=all"); err != nil || status != "" {
		return c, fmt.Errorf("candidate checkout must be clean: %s (%v)", status, err)
	}
	snapshot, err := sourceidentity.Capture(root)
	if err != nil {
		return c, err
	}
	c.ModulesSHA256 = snapshot.ModulesSHA256
	return c, nil
}

// WriteJSON atomically persists an artifact under its declared destination.
func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ci-evidence-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err = tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// ReadRecords loads all JSON result files; malformed files cannot be ignored.
func ReadRecords(dir string) ([]Record, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var records []Record
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var r Record
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		records = append(records, r)
	}
	return records, nil
}
