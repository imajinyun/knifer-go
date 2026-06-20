.PHONY: help doctor install-hooks uninstall-hooks worktree-check change-policy-check security-sensitive-diff agent-evidence agent-evidence-check test test-race coverage-profile coverage-report coverage-check api-check tools-check tools-gen tools-report ai-context-check ci-workflow-check docs-gen docs-check generate mod-verify tidy-check diff-whitespace diff-clean diff-check vet arch lint govulncheck quick-check security-check full-check agent-check agent-full-check agent-security-check ci-agent-governance bench bench-core bench-facade bench-codec bench-smoke check ci-test

GO ?= go
GOLANGCI_LINT ?= golangci-lint
PKGS ?= ./...
COVERAGE_FILE ?= /tmp/go-knifer-coverage.out
BENCH ?= .
BENCH_PKGS ?= ./internal/slice ./internal/maps ./internal/str ./internal/num ./internal/bean ./internal/db ./internal/poi ./internal/imgx ./internal/template ./internal/cli ./internal/ai ./internal/ftp ./internal/ssh ./internal/pinyin ./internal/tokenize
BENCH_FACADE_PKGS ?= ./vslice ./vmap ./vstr ./vnum ./vbean ./vdb ./vcrypto ./vpoi ./vimg ./vtpl ./vcli ./vai ./vftp ./vssh ./vhan ./vtok
BENCH_CODEC_PKGS ?= ./internal/json ./vjson ./internal/xml ./vxml
BENCHTIME ?= 1s
BENCHCOUNT ?= 1

help:
	@echo "Targets:"
	@echo "  test            Run unit tests"
	@echo "  test-race       Run race/shuffle tests and write coverage"
	@echo "  coverage-profile Generate race/shuffle coverage profile"
	@echo "  coverage-report  Print function coverage from COVERAGE_FILE"
	@echo "  coverage-check  Enforce repository and package coverage gates"
	@echo "  bench-core      Run core benchmark baselines"
	@echo "  bench-facade    Run facade benchmark baselines"
	@echo "  bench-codec     Run JSON/XML benchmark baselines"
	@echo "  bench-smoke     Run a short core benchmark smoke check"
	@echo "  doctor          Diagnose local Go/tooling/Git environment"
	@echo "  install-hooks   Enable optional local Git validation hooks"
	@echo "  uninstall-hooks Disable optional local Git validation hooks"
	@echo "  worktree-check  Block unrelated untracked Go files in agent workflows"
	@echo "  change-policy-check Detect change policies from local diff"
	@echo "  security-sensitive-diff Detect changes to security-sensitive packages"
	@echo "  agent-evidence  Emit machine-readable Agent validation evidence"
	@echo "  agent-evidence-check Validate machine-readable Agent validation evidence"
	@echo "  quick-check     Run fast local governance gates"
	@echo "  security-check  Run lint and govulncheck"
	@echo "  full-check      Run full local gates with race coverage"
	@echo "  agent-check     Run default AI/Agent-safe validation gates"
	@echo "  agent-full-check Run full AI/Agent validation gates"
	@echo "  agent-security-check Run AI/Agent security validation gates"
	@echo "  ci-agent-governance Run CI Agent governance policy/evidence gates"
	@echo "  generate        Run go:generate directives (API snapshot, code gen)"
	@echo "  api-check       Verify exported API snapshot is current"
	@echo "  tools-check     Verify machine-readable tools catalog is current"
	@echo "  tools-gen       Regenerate machine-readable tools catalog"
	@echo "  tools-report    Print tools catalog quality report"
	@echo "  docs-gen        Regenerate documentation artifacts"
	@echo "  docs-check      Verify generated documentation artifacts are current"
	@echo "  ai-context-check Verify machine-readable AI context metadata"
	@echo "  ci-workflow-check Verify CI workflow governance invariants"
	@echo "  check           Run local stability gates"
	@echo "  ci-test         Run CI test-job gates"

doctor:
	@echo "== go =="
	@$(GO) version
	@echo "== python3 =="
	@if command -v python3 >/dev/null 2>&1; then python3 --version; else echo "python3 not found"; fi
	@echo "== golangci-lint =="
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then $(GOLANGCI_LINT) version; else echo "$(GOLANGCI_LINT) not found"; fi
	@echo "== govulncheck =="
	@if $(GO) tool govulncheck -version >/dev/null 2>&1; then $(GO) tool govulncheck -version; else echo "govulncheck tool not available; run: go get -tool golang.org/x/vuln/cmd/govulncheck"; fi
	@echo "== git status =="
	@git status --short --branch
	@echo "== module =="
	@$(GO) list -m | grep 'go-knifer' || $(GO) list -m
	@echo "== package list =="
	@$(GO) list ./... >/dev/null
	@echo "go list ./... OK"

install-hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-commit .githooks/pre-push
	@echo "Installed go-knifer Git hooks from .githooks"

uninstall-hooks:
	@git config --unset core.hooksPath || true
	@echo "Disabled go-knifer Git hooks"

worktree-check:
	@if [ "$${SKIP_WORKTREE_CHECK:-}" = "1" ]; then \
		echo "worktree-check skipped because SKIP_WORKTREE_CHECK=1"; \
	else \
		untracked_go="$$(git ls-files --others --exclude-standard -- '*.go')"; \
		if [ -n "$${untracked_go}" ]; then \
			echo "WORKTREE CHECK ERROR: untracked Go files can pollute local tests or commits:" >&2; \
			printf '%s\n' "$${untracked_go}" | while IFS= read -r path; do echo "?? $${path}" >&2; done; \
			echo "Commit/stash/remove them, or set SKIP_WORKTREE_CHECK=1 only when they are intentionally excluded." >&2; \
			exit 1; \
		fi; \
		echo "worktree has no untracked Go files"; \
	fi

change-policy-check:
	bash bin/check_change_policy.sh

security-sensitive-diff:
	bash bin/check_security_sensitive_diff.sh

agent-evidence:
	bash bin/agent_validation_report.sh

agent-evidence-check:
	bash bin/check_agent_evidence.sh

test:
	$(GO) test $(PKGS)

test-race:
	$(GO) test -race -shuffle=on -coverprofile=$(COVERAGE_FILE) $(PKGS)

coverage-profile: test-race

coverage-report:
	$(GO) tool cover -func=$(COVERAGE_FILE)

coverage-check:
	bash bin/check_coverage.sh $(COVERAGE_FILE)

api-check:
	bash bin/check_api_compat.sh

tools-check:
	$(GO) test ./bin/toolsgen

tools-gen:
	$(GO) run ./bin/toolsgen -out docs/api/tools.json -markdown docs/api/tools.md

tools-report:
	$(GO) run ./bin/toolsgen -quality

docs-gen: tools-gen

docs-check: tools-check

ai-context-check:
	bash bin/check_ai_context.sh

ci-workflow-check: ai-context-check

generate:
	$(GO) generate ./...

mod-verify:
	$(GO) mod verify

tidy-check:
	$(GO) mod tidy
	git diff --exit-code -- go.mod go.sum

diff-whitespace:
	git diff --check

diff-clean:
	git diff --exit-code

diff-check: diff-whitespace diff-clean

vet:
	$(GO) vet $(PKGS)

arch:
	bash bin/check_arch.sh

lint:
	$(GOLANGCI_LINT) run $(PKGS)

govulncheck:
	$(GO) tool govulncheck $(PKGS)

quick-check: worktree-check mod-verify vet arch test api-check tools-check ai-context-check diff-whitespace

security-check: lint govulncheck

full-check: worktree-check mod-verify vet arch test-race coverage-check api-check tools-check ai-context-check lint govulncheck diff-whitespace

agent-check: quick-check change-policy-check security-sensitive-diff

agent-full-check: full-check

agent-security-check: security-check

ci-agent-governance: change-policy-check ci-workflow-check agent-evidence agent-evidence-check

bench:
	$(GO) test -bench=$(BENCH) -benchmem -benchtime=$(BENCHTIME) -count=$(BENCHCOUNT) -run=^$$ $(BENCH_PKGS)

bench-core: bench

bench-facade:
	$(GO) test -bench=$(BENCH) -benchmem -benchtime=$(BENCHTIME) -count=$(BENCHCOUNT) -run=^$$ $(BENCH_FACADE_PKGS)

bench-codec:
	$(GO) test -bench=$(BENCH) -benchmem -benchtime=$(BENCHTIME) -count=$(BENCHCOUNT) -run=^$$ $(BENCH_CODEC_PKGS)

bench-smoke:
	$(GO) test -bench=Benchmark -benchtime=100ms -count=1 -run=^$$ $(BENCH_PKGS) $(BENCH_FACADE_PKGS) $(BENCH_CODEC_PKGS)

check: full-check

ci-test: mod-verify vet tidy-check diff-check arch test-race coverage-check api-check tools-check ai-context-check
