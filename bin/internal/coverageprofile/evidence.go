package coverageprofile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/version"
	"os"
	"slices"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/sourceidentity"
)

type Evidence struct {
	Schema        int                     `json:"schema_version"`
	Before        sourceidentity.Snapshot `json:"before"`
	After         sourceidentity.Snapshot `json:"after"`
	ProfileSHA256 string                  `json:"profile_sha256"`
	GoVersion     string                  `json:"go_version"`
	RunID         string                  `json:"run_id,omitempty"`
	Attempt       string                  `json:"run_attempt,omitempty"`
	Command       []string                `json:"command"`
	Packages      []string                `json:"packages"`
	TestsExitCode int                     `json:"tests_exit_code"`
	CreatedAt     time.Time               `json:"created_at"`
	Complete      bool                    `json:"complete"`
}

func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

func ValidateEvidence(root, profile string) (*Evidence, error) {
	data, err := os.ReadFile(profile + ".meta.json")
	if err != nil {
		return nil, fmt.Errorf("coverage execution evidence missing: %w", err)
	}
	var evidence Evidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		return nil, err
	}
	current, err := sourceidentity.Capture(root)
	if err != nil {
		return nil, err
	}
	if evidence.Schema != 1 || !evidence.Complete || evidence.TestsExitCode != 0 || evidence.CreatedAt.IsZero() {
		return nil, fmt.Errorf("coverage evidence does not prove a complete passing run")
	}
	if !version.IsValid(evidence.GoVersion) {
		return nil, fmt.Errorf("coverage evidence must record an actual Go version")
	}
	if evidence.Before != current || evidence.After != current {
		return nil, fmt.Errorf("coverage source/commit/module identity is stale")
	}
	if !slices.Equal(evidence.Packages, []string{"./..."}) {
		return nil, fmt.Errorf("repository gate requires complete ./... test scope")
	}
	if !slices.Contains(evidence.Command, "-coverpkg=./...") {
		return nil, fmt.Errorf("coverage must include caller execution")
	}
	if run := os.Getenv("GITHUB_RUN_ID"); run != "" && (evidence.RunID != run || evidence.Attempt != os.Getenv("GITHUB_RUN_ATTEMPT")) {
		return nil, fmt.Errorf("coverage evidence belongs to another CI run/attempt")
	}
	hash, err := HashFile(profile)
	if err != nil {
		return nil, err
	}
	if hash != evidence.ProfileSHA256 {
		return nil, fmt.Errorf("coverage profile hash does not match execution evidence")
	}
	return &evidence, nil
}
