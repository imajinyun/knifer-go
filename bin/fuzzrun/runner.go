package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type commandFunc func(context.Context, string, string, []string, io.Writer) (string, int, error)

type Config struct {
	Root, Go, Artifacts                string
	Packages                           []string
	FuzzTime, MinimizeTime, RunTimeout time.Duration
	Parallel                           int
	execute                            commandFunc
}

type Input struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type Target struct {
	Package      string   `json:"package"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	ExitCode     int      `json:"exit_code"`
	Directory    string   `json:"directory"`
	BinarySHA256 string   `json:"binary_sha256,omitempty"`
	Args         []string `json:"args,omitempty"`
	Corpus       []Input  `json:"corpus"`
	Error        string   `json:"error,omitempty"`
}
type Summary struct {
	Directory    string    `json:"artifact_directory"`
	Status       string    `json:"status"`
	Commit       string    `json:"commit"`
	Tree         string    `json:"tree"`
	Dirty        bool      `json:"worktree_dirty"`
	GoVersion    string    `json:"go_version"`
	GoToolchain  string    `json:"go_toolchain"`
	FuzzTime     string    `json:"fuzz_time"`
	MinimizeTime string    `json:"minimize_time"`
	RunTimeout   string    `json:"run_timeout"`
	Parallel     int       `json:"parallel"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	Targets      []Target  `json:"targets"`
	Error        string    `json:"error,omitempty"`
}

func Run(parent context.Context, cfg Config) (s Summary, runErr error) {
	if cfg.FuzzTime <= 0 || cfg.MinimizeTime <= 0 || cfg.RunTimeout <= 0 || cfg.Parallel <= 0 {
		return s, errors.New("fuzz budgets and worker count must be positive")
	}
	var err error
	if cfg.Root, err = filepath.Abs(cfg.Root); err != nil {
		return s, err
	}
	if !filepath.IsAbs(cfg.Artifacts) {
		cfg.Artifacts = filepath.Join(cfg.Root, cfg.Artifacts)
	}
	if rel, err := filepath.Rel(cfg.Root, cfg.Artifacts); err != nil {
		return s, err
	} else if filepath.IsLocal(rel) && !strings.HasPrefix(filepath.ToSlash(rel), ".aiflow/") {
		return s, errors.New("workspace artifact directories must be under .aiflow")
	}
	if err := os.MkdirAll(cfg.Artifacts, 0o750); err != nil {
		return s, err
	}
	runDir, err := os.MkdirTemp(cfg.Artifacts, "run-")
	if err != nil {
		return s, err
	}
	s = Summary{Directory: runDir, Status: "running", StartedAt: time.Now().UTC(), FuzzTime: cfg.FuzzTime.String(), MinimizeTime: cfg.MinimizeTime.String(), RunTimeout: cfg.RunTimeout.String(), Parallel: cfg.Parallel, Targets: []Target{}}
	s.Commit, s.Tree, s.Dirty = revision(cfg.Root)
	defer func() {
		s.FinishedAt = time.Now().UTC()
		if runErr != nil {
			s.Status = "failed"
			s.Error = runErr.Error()
		}
		if err := writeSummary(s); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()
	if err := writeSummary(s); err != nil {
		return s, err
	}
	buildDir := filepath.Join(runDir, ".build")
	if err := os.MkdirAll(buildDir, 0o700); err != nil {
		return s, err
	}
	defer func() { _ = os.RemoveAll(buildDir) }()
	ctx, cancel := context.WithTimeout(parent, cfg.RunTimeout)
	defer cancel()
	if cfg.execute == nil {
		cfg.execute = command
	}
	log, err := os.Create(filepath.Join(runDir, "discovery.log"))
	if err != nil {
		return s, err
	}
	defer func() { _ = log.Close() }()
	text, _, err := cfg.execute(ctx, cfg.Root, cfg.Go, []string{"env", "-json", "GOVERSION", "GOTOOLCHAIN"}, log)
	if err != nil {
		return s, err
	}
	var env struct{ GOVERSION, GOTOOLCHAIN string }
	if err := json.Unmarshal([]byte(text), &env); err != nil {
		return s, err
	}
	s.GoVersion, s.GoToolchain = env.GOVERSION, env.GOTOOLCHAIN
	args := append([]string{"test", "-list", "^Fuzz"}, cfg.Packages...)
	text, _, err = cfg.execute(ctx, cfg.Root, cfg.Go, args, log)
	if err != nil {
		return s, fmt.Errorf("discover fuzz targets: %w", err)
	}
	s.Targets, err = discover(text)
	if err != nil {
		return s, err
	}
	type build struct {
		binary, source, hash string
		code                 int
		err                  error
	}
	builds := map[string]build{}
	failed := false
	for i := range s.Targets {
		target := &s.Targets[i]
		if ctx.Err() != nil {
			target.Status = "skipped_timeout"
			target.Error = ctx.Err().Error()
			failed = true
			continue
		}
		b, ok := builds[target.Package]
		if !ok {
			b.binary = filepath.Join(buildDir, fmt.Sprintf("package-%d.test", len(builds)))
			b.source, b.code, b.err = cfg.execute(ctx, cfg.Root, cfg.Go, []string{"list", "-f", "{{.Dir}}", target.Package}, log)
			b.source = strings.TrimSpace(b.source)
			if b.err == nil && b.source == "" {
				b.err = errors.New("empty package directory")
			}
			if b.err == nil {
				_, b.code, b.err = cfg.execute(ctx, cfg.Root, cfg.Go, []string{"test", "-c", "-fuzz=.", "-o", b.binary, target.Package}, log)
			}
			if b.err == nil {
				b.hash, b.err = fileHash(b.binary)
			}
			builds[target.Package] = b
		}
		if b.err != nil {
			target.Status = "build_failed"
			target.ExitCode = b.code
			target.Error = b.err.Error()
			failed = true
			continue
		}
		work := filepath.Join(runDir, "targets", fmt.Sprintf("%03d-%s", i, target.Name))
		target.Directory = work
		target.BinarySHA256 = b.hash
		if err := os.MkdirAll(work, 0o750); err != nil {
			return s, err
		}
		if err := copyTestdata(b.source, work); err != nil {
			target.Status = "fixture_failed"
			target.Error = err.Error()
			failed = true
			continue
		}
		target.Status = "running"
		if err := writeSummary(s); err != nil {
			return s, err
		}
		target.Args = []string{"-test.run=^$", "-test.fuzz=^" + regexp.QuoteMeta(target.Name) + "$", "-test.fuzztime=" + cfg.FuzzTime.String(), "-test.fuzzminimizetime=" + cfg.MinimizeTime.String(), fmt.Sprintf("-test.parallel=%d", cfg.Parallel), "-test.fuzzcachedir=" + filepath.Join(buildDir, fmt.Sprintf("cache-%d", i)), "-test.timeout=" + (cfg.FuzzTime + cfg.MinimizeTime + 30*time.Second).String()}
		if err := executeTarget(ctx, cfg, b.binary, target, runDir); err != nil {
			failed = true
		}
		if err := writeSummary(s); err != nil {
			return s, err
		}
	}
	if failed {
		return s, errors.New("fuzz run failed or incomplete; inspect summary.json and target logs")
	}
	s.Status = "passed"
	return s, nil
}

func discover(text string) ([]Target, error) {
	targets := []Target{}
	names := []string{}
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) == 1 && strings.HasPrefix(fields[0], "Fuzz") {
			names = append(names, fields[0])
			continue
		}
		if fields[0] == "ok" && len(fields) > 1 {
			for _, name := range names {
				targets = append(targets, Target{Package: fields[1], Name: name, Status: "pending", ExitCode: -1, Corpus: []Input{}})
			}
			names = nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(names) > 0 || len(targets) == 0 {
		return nil, errors.New("no complete Fuzz target listing found")
	}
	return targets, nil
}

func executeTarget(ctx context.Context, cfg Config, binary string, target *Target, runDir string) error {
	log, err := os.Create(filepath.Join(target.Directory, "fuzz.log"))
	if err != nil {
		target.Status = "artifact_failed"
		target.Error = err.Error()
		return err
	}
	defer func() { _ = log.Close() }()
	deadline, cancel := context.WithTimeout(ctx, cfg.FuzzTime+cfg.MinimizeTime+40*time.Second)
	defer cancel()
	_, code, err := cfg.execute(deadline, target.Directory, binary, target.Args, log)
	target.ExitCode = code
	target.Status = "passed"
	if err != nil || code != 0 {
		target.Status = "failed"
		if err == nil {
			err = fmt.Errorf("exit %d", code)
		}
		target.Error = err.Error()
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || deadline.Err() != nil {
		target.Status = "timeout"
	}
	target.Corpus, err = collectCorpus(filepath.Join(target.Directory, "testdata", "fuzz", target.Name), runDir, err)
	if err != nil && target.Status == "passed" {
		target.Status = "artifact_failed"
		target.Error = err.Error()
	}
	return err
}
