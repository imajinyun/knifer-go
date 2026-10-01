package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var cfg Config
	flag.StringVar(&cfg.Root, "root", ".", "module root")
	flag.StringVar(&cfg.Go, "go", "go", "Go executable")
	flag.StringVar(&cfg.Artifacts, "artifacts", ".aiflow/fuzz", "artifact root")
	flag.DurationVar(&cfg.FuzzTime, "time", time.Second, "fuzz time per target")
	flag.DurationVar(&cfg.MinimizeTime, "minimize", 2*time.Second, "minimization time per failure")
	flag.DurationVar(&cfg.RunTimeout, "timeout", 10*time.Minute, "whole-run deadline, including builds")
	flag.IntVar(&cfg.Parallel, "parallel", 4, "maximum fuzz workers per target")
	flag.Parse()
	cfg.Packages = flag.Args()
	if len(cfg.Packages) == 0 {
		cfg.Packages = []string{"./..."}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	summary, err := Run(ctx, cfg)
	stop()
	fmt.Printf("fuzz artifacts: %s\n", summary.Directory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("fuzz: passed %d target(s)\n", len(summary.Targets))
}
