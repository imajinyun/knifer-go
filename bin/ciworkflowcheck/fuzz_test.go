package main

import (
	"os"
	"strings"
	"testing"
)

func TestFuzzWorkflowArtifacts(t *testing.T) {
	for _, name := range []string{"go", "release", "fuzz"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("../../.github/workflows/" + name + ".yml")
			if err != nil {
				t.Fatal(err)
			}
			for _, bad := range []bool{false, true} {
				text := string(data)
				if bad {
					text = strings.ReplaceAll(text, "path: .aiflow/fuzz/", "path: missing/")
				}
				c := &checker{}
				c.checkFuzzArtifacts(text, name+".yml", name)
				if (len(c.findings) > 0) != bad {
					t.Fatalf("bad=%v findings=%+v", bad, c.findings)
				}
			}
		})
	}
}
