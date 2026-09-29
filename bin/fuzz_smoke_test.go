package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestFuzzSmoke(t *testing.T) {
	const listing = "FuzzFirst\nFuzzSecond\nok  example/one 0.01s\n? example/empty [no test files]\nFuzzThird\nok  example/two 0.01s\n"
	for _, tt := range []struct {
		name         string
		listing      string
		discoverCode string
		fuzzCode     string
		wantCode     int
		wantTargets  []string
	}{
		{name: "all_packages", listing: listing, wantTargets: []string{"example/one ^FuzzFirst$", "example/one ^FuzzSecond$", "example/two ^FuzzThird$"}},
		{name: "no_targets", listing: "ok example/one 0.01s\n", wantCode: 1},
		{name: "discovery_error", listing: listing, discoverCode: "17", wantCode: 17},
		{name: "execution_error", listing: listing, fuzzCode: "29", wantCode: 29, wantTargets: []string{"example/one ^FuzzFirst$"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newGovernanceFixture(t)
			trace := fixture.WriteTempFile("trace", "")
			fakeGo := fixture.WriteTempFile("fake go", `#!/usr/bin/env bash
set -eu
if [ "$2" = '-list' ]; then
  printf '%s' "$FUZZ_LISTING"
  exit "${DISCOVER_CODE:-0}"
fi
printf '%s %s\n' "$7" "$5" >> "$FUZZ_TRACE"
if [ "$1 $2 $3 $4 $6" != 'test -run ^$ -fuzz -fuzztime=2s' ]; then
  exit 99
fi
exit "${FUZZ_CODE:-0}"
`)
			if err := os.Chmod(fakeGo, 0o700); err != nil {
				t.Fatal(err)
			}
			output, err := fixture.RunScriptArgs("bin/fuzz_smoke.sh", []string{"./..."},
				"GO="+fakeGo, "FUZZTIME=2s", "FUZZ_LISTING="+tt.listing,
				"DISCOVER_CODE="+tt.discoverCode, "FUZZ_CODE="+tt.fuzzCode, "FUZZ_TRACE="+trace)
			code := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				code = exitErr.ExitCode()
			}
			if code != tt.wantCode {
				t.Fatalf("exit = %d, want %d; output:\n%s", code, tt.wantCode, output)
			}
			calls, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := strings.TrimSpace(string(calls)), strings.Join(tt.wantTargets, "\n"); got != want {
				t.Fatalf("fuzz calls = %q, want %q", got, want)
			}
		})
	}
}
