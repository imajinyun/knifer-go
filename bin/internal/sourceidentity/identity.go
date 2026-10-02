// Package sourceidentity fingerprints committed and local repository inputs.
package sourceidentity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Snapshot struct {
	Commit        string `json:"commit"`
	Tree          string `json:"tree"`
	SourceSHA256  string `json:"source_sha256"`
	ModulesSHA256 string `json:"modules_sha256"`
}

func Capture(root string) (Snapshot, error) {
	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		data, err := cmd.Output()
		return strings.TrimSpace(string(data)), err
	}
	s := Snapshot{}
	var err error
	if s.Commit, err = git("rev-parse", "HEAD"); err != nil {
		return s, err
	}
	if s.Tree, err = git("rev-parse", "HEAD^{tree}"); err != nil {
		return s, err
	}
	files, err := git("ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return s, err
	}
	set := map[string]bool{}
	for _, name := range strings.Split(files, "\x00") {
		if name != "" {
			set[name] = true
		}
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	all, modules := sha256.New(), sha256.New()
	moduleCount := 0
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, name))
		state := "present"
		if err != nil {
			if !os.IsNotExist(err) {
				return s, err
			}
			state = "deleted"
		}
		digest := sha256.Sum256(data)
		entry := []byte(name + "\x00" + state + "\x00" + hex.EncodeToString(digest[:]) + "\x00")
		_, _ = all.Write(entry)
		if base := filepath.Base(name); base == "go.mod" || base == "go.sum" {
			moduleCount++
			_, _ = modules.Write(entry)
		}
	}
	if moduleCount == 0 {
		return s, fmt.Errorf("no module inputs found")
	}
	s.SourceSHA256 = hex.EncodeToString(all.Sum(nil))
	s.ModulesSHA256 = hex.EncodeToString(modules.Sum(nil))
	return s, nil
}
