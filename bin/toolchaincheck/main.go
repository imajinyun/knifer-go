package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/imajinyun/knifer-go/bin/internal/toolchainpolicy"
)

func main() {
	root := flag.String("root", ".", "repository root")
	goCmd := flag.String("go", "go", "Go executable")
	lint := flag.String("lint", "", "optional golangci-lint executable")
	flag.Parse()
	if err := check(*root, *goCmd, *lint); err != nil {
		fmt.Fprintln(os.Stderr, "toolchain check:", err)
		os.Exit(1)
	}
}

func check(root, goCmd, lint string) error {
	p, err := toolchainpolicy.Load(root)
	if err != nil {
		return err
	}
	run := func(command string, args ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, command, args...)
		cmd.Dir = root
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil || stderr.Len() > 0 {
			return nil, fmt.Errorf("%s %v: %v %s", command, args, err, stderr.String())
		}
		return out, nil
	}
	data, err := run(goCmd, "env", "-json", "GOVERSION", "GOTOOLCHAIN", "GOROOT", "GOOS", "GOARCH")
	if err != nil {
		return err
	}
	var env struct{ GOVERSION, GOTOOLCHAIN, GOROOT, GOOS, GOARCH string }
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}
	if err := p.ValidateRuntime(env.GOVERSION, env.GOTOOLCHAIN); err != nil {
		return err
	}
	fmt.Printf("toolchain: actual=%s GOTOOLCHAIN=%s minimum=%s release=%s\n", env.GOVERSION, env.GOTOOLCHAIN, p.Minimum, p.Release)
	fmt.Printf("toolchain: GOROOT=%s platform=%s/%s\n", env.GOROOT, env.GOOS, env.GOARCH)
	data, err = run(goCmd, "list", "-mod=readonly", "-m", "-json", "all")
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var module struct{ Path, GoVersion string }
		err := decoder.Decode(&module)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if err := p.ValidateDependency(module.Path, module.GoVersion); err != nil {
			return err
		}
	}
	if lint != "" {
		data, err := run(lint, "version")
		if err != nil {
			return err
		}
		if err := p.ValidateLinter(string(data), env.GOVERSION); err != nil {
			return err
		}
		fmt.Print(string(data))
	}
	fmt.Println("toolchain policy and dependency requirements are valid")
	return nil
}
