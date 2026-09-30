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
	"regexp"
	"syscall"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/cievidence"
	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: ciresult run|collect [flags]")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := execute(ctx, os.Args[1:])
	stop()
	os.Exit(code)
}

func execute(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	id := flags.String("id", "", "command result ID")
	out := flags.String("out", ".aiflow/ci/records", "record directory (run) or manifest file (collect)")
	results := flags.String("results", ".aiflow/ci/results", "downloaded result directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	switch args[0] {
	case "run":
		if !regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`).MatchString(*id) || len(flags.Args()) == 0 {
			fmt.Fprintln(os.Stderr, "run requires a safe ID and command after --")
			return 2
		}
		return run(ctx, *root, *id, filepath.Join(*out, *id+".json"), flags.Args())
	case "collect":
		manifest, err := collect(*root, *results, os.Getenv("CI_NEEDS_JSON"))
		if err != nil {
			manifest.Status = "failed"
			manifest.Failures = append(manifest.Failures, err.Error())
		}
		if writeErr := cievidence.WriteJSON(*out, manifest); writeErr != nil {
			fmt.Fprintln(os.Stderr, writeErr)
			return 1
		}
		if manifest.Status != "passed" {
			fmt.Fprintln(os.Stderr, "CI evidence rejected:", manifest.Failures)
			return 1
		}
		fmt.Printf("validated %d command records for %s run %s/%s\n", len(manifest.Records), manifest.Candidate.Commit, manifest.Candidate.RunID, manifest.Candidate.Attempt)
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown mode", args[0])
		return 2
	}
}

func run(ctx context.Context, root, id, out string, args []string) int {
	r := cievidence.Record{Schema: 1, ID: id, Command: args, Status: "failed", StartedAt: time.Now().UTC()}
	code := 2
	var err error
	r.Before, err = cievidence.Capture(root)
	if err == nil {
		cmd := exec.CommandContext(ctx, "go", "env", "-json", "GOVERSION", "GOTOOLCHAIN", "GOWORK")
		cmd.Dir = root
		var data []byte
		data, err = cmd.Output()
		if err == nil {
			var env struct{ GOVERSION, GOTOOLCHAIN, GOWORK string }
			err = json.Unmarshal(data, &env)
			r.GoVersion = env.GOVERSION
			r.GoToolchain = env.GOTOOLCHAIN
			r.GoWork = env.GOWORK
		}
	}
	if err == nil {
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		code = 0
		if err != nil {
			code = 1
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() > 0 {
				code = exitErr.ExitCode()
			}
		}
		if ctx.Err() != nil {
			code = 130
			r.Status = "cancelled"
		}
		var captureErr error
		r.After, captureErr = cievidence.Capture(root)
		if captureErr != nil {
			err = errors.Join(err, captureErr)
			if code == 0 {
				code = 1
			}
		}
		if r.Before != r.After {
			err = errors.Join(err, errors.New("candidate changed during command"))
			if code == 0 {
				code = 1
			}
		}
	}
	if err != nil {
		r.Error = err.Error()
	}
	if ctx.Err() != nil {
		code = 130
		r.Status = "cancelled"
	}
	if code == 0 {
		r.Status = "passed"
	}
	r.ExitCode = &code
	r.FinishedAt = time.Now().UTC()
	if err := cievidence.WriteJSON(out, r); err != nil {
		fmt.Fprintln(os.Stderr, "cannot write result:", err)
		return 1
	}
	return code
}

func collect(root, dir, needs string) (cievidence.Manifest, error) {
	c, err := cievidence.Capture(root)
	if err != nil {
		return cievidence.Manifest{}, err
	}
	p, err := toolchainpolicy.Load(root)
	if err != nil {
		return cievidence.Manifest{}, err
	}
	records, err := cievidence.ReadRecords(dir)
	if err != nil {
		return cievidence.Manifest{}, err
	}
	var jobs map[string]cievidence.JobResult
	if err := json.Unmarshal([]byte(needs), &jobs); err != nil {
		return cievidence.Manifest{}, fmt.Errorf("invalid CI_NEEDS_JSON: %w", err)
	}
	return cievidence.Validate(c, p, records, jobs), nil
}
