package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSharedSecurityCoverage(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "internal/shared")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "shared.go"), []byte("package shared\nfunc Check() bool { return true }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]any{"project": map[string]string{"module": "example.com/lib"}, "coverage_gates": map[string]any{"shared_security_packages": []string{"internal/shared"}, "security_sensitive_min_threshold": 80}, "public_facades": []any{}, "security_sensitive_packages": []string{}})
	if err := os.WriteFile(filepath.Join(root, "ai-context.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfig(root, "profile")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.securitySensitivePaths) != 1 || cfg.securitySensitivePaths[0] != "example.com/lib/internal/shared" {
		t.Fatalf("shared paths=%v", cfg.securitySensitivePaths)
	}
	for _, tt := range []struct {
		name    string
		lines   []profileLine
		wantErr bool
	}{
		{"missing", nil, true}, {"below", []profileLine{{file: "example.com/lib/internal/shared/shared.go", statements: 10, count: 0}}, true}, {"passed", []profileLine{{file: "example.com/lib/internal/shared/shared.go", statements: 10, count: 1}}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSecuritySensitive(tt.lines, cfg.securitySensitivePaths, 80, false, true)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v", err)
			}
		})
	}
}
