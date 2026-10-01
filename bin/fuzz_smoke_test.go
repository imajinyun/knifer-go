package main

import (
	"os"
	"strings"
	"testing"
)

func TestFuzzSmoke(t *testing.T) {
	fixture := newGovernanceFixture(t)
	trace := fixture.WriteTempFile("trace", "")
	fakeGo := fixture.WriteTempFile("fake go", `#!/usr/bin/env bash
printf '%s\n' "$@" > "$FUZZ_TRACE"
exit 17
`)
	if err := os.Chmod(fakeGo, 0o700); err != nil {
		t.Fatal(err)
	}
	output, err := fixture.RunScriptArgs("bin/fuzz_smoke.sh", []string{"./internal/codec"}, "GO="+fakeGo, "FUZZ_TRACE="+trace, "FUZZTIME=2s", "FUZZ_MINIMIZE_TIME=1s", "FUZZ_RUN_TIMEOUT=1m", "FUZZ_PARALLEL=2", "FUZZ_ARTIFACT_DIR="+fixture.Root())
	if err == nil {
		t.Fatalf("wrapper swallowed runner failure: %s", output)
	}
	data, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"run", "./bin/fuzzrun", "-go", fakeGo, "-artifacts", fixture.Root(), "-time", "2s", "-minimize", "1s", "-timeout", "1m", "-parallel", "2", "./internal/codec"}
	if strings.TrimSpace(string(data)) != strings.Join(want, "\n") {
		t.Fatalf("args=%q want=%q", data, want)
	}
}
