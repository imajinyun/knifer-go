package toolchainpolicy

import "testing"

func validPolicy() Policy {
	return Policy{Minimum: "1.26.0", Release: "1.27.1", Test: []string{"1.26.8", "1.27.1"}, Lint: "v2.14.0"}
}

func TestValidate(t *testing.T) {
	for _, tt := range []struct {
		name    string
		edit    func(*Policy)
		module  string
		wantErr bool
	}{
		{name: "valid", module: "1.26.0"},
		{name: "module_upgrade", module: "1.27.0", wantErr: true},
		{name: "unfixed_release", module: "1.26.0", edit: func(p *Policy) { p.Release = "1.27" }, wantErr: true},
		{name: "old_test_line", module: "1.26.0", edit: func(p *Policy) { p.Test[0] = "1.25.13" }, wantErr: true},
		{name: "missing_minimum", module: "1.26.0", edit: func(p *Policy) { p.Test[0] = "1.27.0" }, wantErr: true},
		{name: "duplicate_line", module: "1.26.0", edit: func(p *Policy) { p.Test[1] = "1.26.9" }, wantErr: true},
		{name: "untested_release", module: "1.26.0", edit: func(p *Policy) { p.Release = "1.27.2" }, wantErr: true},
		{name: "unpinned_linter", module: "1.26.0", edit: func(p *Policy) { p.Lint = "latest" }, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := validPolicy()
			if tt.edit != nil {
				tt.edit(&p)
			}
			if err := p.Validate(tt.module); (err != nil) != tt.wantErr {
				t.Fatalf("Validate = %v, want error %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDependency(t *testing.T) {
	for _, v := range []string{"", "1.24.0", "1.26.0", "1.26.1", "1.27.0"} {
		t.Run(v, func(t *testing.T) {
			wantErr := v == "1.26.1" || v == "1.27.0"
			if err := validPolicy().ValidateDependency("example.com/dep", v); (err != nil) != wantErr {
				t.Fatalf("ValidateDependency(%q) = %v", v, err)
			}
		})
	}
}

func TestValidateRuntime(t *testing.T) {
	for _, tt := range []struct {
		name, actual, mode string
		wantErr            bool
	}{
		{"minimum", "go1.26.8", "local", false},
		{"release", "go1.27.1", "local", false},
		{"old", "go1.25.13", "local", true},
		{"auto", "go1.27.1", "auto", true},
		{"unknown", "devel", "local", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := validPolicy().ValidateRuntime(tt.actual, tt.mode); (err != nil) != tt.wantErr {
				t.Fatalf("ValidateRuntime = %v", err)
			}
		})
	}
}

func TestValidateLinter(t *testing.T) {
	for _, tt := range []struct {
		name, output, actual string
		wantErr              bool
	}{
		{"supported", "golangci-lint has version 2.14.0 built with go1.27.1 from test", "go1.26.8", false},
		{"old_build", "golangci-lint has version 2.14.0 built with go1.25.13 from test", "go1.26.8", true},
		{"older_than_analysis", "golangci-lint has version 2.14.0 built with go1.26.8 from test", "go1.27.1", true},
		{"wrong_version", "golangci-lint has version 2.12.2 built with go1.27.1 from test", "go1.27.1", true},
		{"unknown_build", "golangci-lint unknown", "go1.27.1", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := validPolicy().ValidateLinter(tt.output, tt.actual); (err != nil) != tt.wantErr {
				t.Fatalf("ValidateLinter = %v", err)
			}
		})
	}
}
