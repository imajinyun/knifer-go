.PHONY: help doctor install-hooks uninstall-hooks worktree-check change-policy-check security-sensitive-diff agent-evidence agent-evidence-check test test-race race-test shuffle-test fuzz-smoke coverage-profile coverage-report coverage-check release-notes-check api-check api-freeze-check governance-maturity-check tools-check tools-gen tools-report docs-quickstart-check ai-context-check ci-workflow-check docs-gen docs-check generate mod-verify tidy-check mod-check diff-whitespace diff-clean diff-check vet arch lint govulncheck quick-check security-check full-check release-check agent-check agent-full-check agent-security-check ci-agent-governance bench bench-core bench-facade bench-codec bench-smoke bench-baseline bench-compare bench-regression-check benchstat check ci-test

GO ?= go
GOLANGCI_LINT ?= golangci-lint
PKGS ?= ./...
COVERAGE_FILE ?= /tmp/knifer-go-coverage.out
BENCH ?= .
BENCH_PKGS ?= ./internal/slice ./internal/maps ./internal/str ./internal/num ./internal/bean ./internal/db ./internal/poi ./internal/imgx ./internal/template ./internal/cli ./internal/ai ./internal/ftp ./internal/ssh ./internal/pinyin ./internal/tokenize
BENCH_FACADE_PKGS ?= ./vslice ./vmap ./vstr ./vnum ./vbean ./vdb ./vcrypto ./vpoi ./vimg ./vtpl ./vcli ./vai ./vftp ./vssh ./vhan ./vtok
BENCH_CODEC_PKGS ?= ./internal/json ./vjson ./internal/xml ./vxml
BENCHTIME ?= 1s
BENCHCOUNT ?= 1
FUZZTIME ?= 1s
FUZZ_PKGS ?= ./internal/codec ./internal/json ./internal/sets
BENCH_BASELINE ?=
BENCH_CURRENT ?=
BENCH_BASELINE_OUT ?= /tmp/knifer-go-bench-baseline.txt
BENCH_CURRENT_OUT ?= /tmp/knifer-go-bench-current.txt

help:
	@echo "Targets:"
	@echo "  test            Run unit tests"
	@echo "  test-race       Run race/shuffle tests and write coverage"
	@echo "  race-test       Run race-enabled tests"
	@echo "  shuffle-test    Run order-shuffled tests"
	@echo "  fuzz-smoke      Run short fuzz/property smoke checks"
	@echo "  coverage-profile Generate race/shuffle coverage profile"
	@echo "  coverage-report  Print function coverage from COVERAGE_FILE"
	@echo "  coverage-check  Enforce repository and package coverage gates"
	@echo "  release-notes-check Verify changelog release-note readiness"
	@echo "  api-freeze-check Verify v1 API freeze/deprecation governance"
	@echo "  governance-maturity-check Verify API convergence, lifecycle, error, threat, and benchmark governance"
	@echo "  bench-core      Run core benchmark baselines"
	@echo "  bench-facade    Run facade benchmark baselines"
	@echo "  bench-codec     Run JSON/XML benchmark baselines"
	@echo "  bench-smoke     Run a short core benchmark smoke check"
	@echo "  bench-baseline  Write repeated benchmark baseline to BENCH_BASELINE_OUT"
	@echo "  bench-compare   Write repeated benchmark current output and compare with benchstat"
	@echo "  bench-regression-check Verify benchmark regression metadata and thresholds"
	@echo "  benchstat       Compare BENCH_BASELINE and BENCH_CURRENT files with benchstat"
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
	@echo "  release-check   Run release readiness gates"
	@echo "  agent-check     Run default AI/Agent-safe validation gates"
	@echo "  agent-full-check Run full AI/Agent validation gates"
	@echo "  agent-security-check Run AI/Agent security validation gates"
	@echo "  ci-agent-governance Run CI Agent governance policy/evidence gates"
	@echo "  generate        Run go:generate directives (API snapshot, code gen)"
	@echo "  mod-check       Verify go.mod and go.sum are tidy"
	@echo "  api-check       Verify exported API snapshot is current"
	@echo "  tools-check     Verify machine-readable tools catalog is current"
	@echo "  tools-gen       Regenerate machine-readable tools catalog"
	@echo "  tools-report    Print tools catalog quality report"
	@echo "  docs-quickstart-check Verify facade quickstart documentation structure"
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
	@$(GO) list -m | grep 'knifer-go' || $(GO) list -m
	@echo "== package list =="
	@$(GO) list ./... >/dev/null
	@echo "go list ./... OK"

install-hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-commit .githooks/pre-push
	@echo "Installed knifer-go Git hooks from .githooks"

uninstall-hooks:
	@git config --unset core.hooksPath || true
	@echo "Disabled knifer-go Git hooks"

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

race-test:
	$(GO) test -race $(PKGS)

shuffle-test:
	$(GO) test -shuffle=on $(PKGS)

fuzz-smoke:
	@for pkg in $(FUZZ_PKGS); do \
		fuzzes="$$( $(GO) test -list '^Fuzz' $$pkg | grep '^Fuzz' || true )"; \
		for fuzz in $$fuzzes; do \
			$(GO) test -run=^$$ -fuzz=^$$fuzz$$ -fuzztime=$(FUZZTIME) $$pkg || exit $$?; \
		done; \
	done

coverage-profile: test-race

coverage-report:
	$(GO) tool cover -func=$(COVERAGE_FILE)

coverage-check:
	bash bin/check_coverage.sh $(COVERAGE_FILE)

release-notes-check:
	bash bin/check_release_notes.sh

api-check:
	bash bin/check_api_compat.sh

api-freeze-check: api-check tools-check
	bash bin/check_api_freeze.sh

governance-maturity-check: ai-context-check tools-check
	bash bin/check_governance_maturity.sh

tools-check:
	$(GO) test ./bin/toolsgen

tools-gen:
	$(GO) run ./bin/toolsgen -out docs/api/tools.json -markdown docs/api/tools.md

tools-report:
	$(GO) run ./bin/toolsgen -quality

docs-quickstart-check:
	bash bin/check_docs_quickstart.sh

docs-gen: tools-gen

docs-check: tools-check docs-quickstart-check

ai-context-check:
	bash bin/check_ai_context.sh

ci-workflow-check: bench-regression-check

generate:
	$(GO) generate ./...

mod-verify:
	$(GO) mod verify

tidy-check:
	$(GO) mod tidy
	git diff --exit-code -- go.mod go.sum

mod-check: tidy-check

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

quick-check: worktree-check mod-verify vet arch test api-check docs-check bench-regression-check diff-whitespace

security-check: lint govulncheck

full-check: worktree-check mod-verify vet arch test-race coverage-check api-check docs-check bench-regression-check lint govulncheck diff-whitespace

release-check: release-notes-check full-check ci-workflow-check

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

bench-baseline:
	$(GO) test -bench=$(BENCH) -benchmem -benchtime=$(BENCHTIME) -count=$(BENCHCOUNT) -run=^$$ $(BENCH_PKGS) $(BENCH_FACADE_PKGS) $(BENCH_CODEC_PKGS) | tee "$(BENCH_BASELINE_OUT)"

bench-compare:
	$(GO) test -bench=$(BENCH) -benchmem -benchtime=$(BENCHTIME) -count=$(BENCHCOUNT) -run=^$$ $(BENCH_PKGS) $(BENCH_FACADE_PKGS) $(BENCH_CODEC_PKGS) | tee "$(BENCH_CURRENT_OUT)"
	$(MAKE) benchstat BENCH_BASELINE="$(BENCH_BASELINE_OUT)" BENCH_CURRENT="$(BENCH_CURRENT_OUT)"

bench-regression-check: governance-maturity-check
	bash bin/check_governance_maturity.sh --bench-only

benchstat:
	@test -n "$(BENCH_BASELINE)" || (echo "BENCH_BASELINE is required" >&2; exit 2)
	@test -n "$(BENCH_CURRENT)" || (echo "BENCH_CURRENT is required" >&2; exit 2)
	@command -v benchstat >/dev/null 2>&1 || (echo "benchstat not found; run: go install golang.org/x/perf/cmd/benchstat@latest" >&2; exit 2)
	benchstat "$(BENCH_BASELINE)" "$(BENCH_CURRENT)"

check: full-check

ci-test: mod-verify vet tidy-check diff-check arch test-race coverage-check api-check docs-check ai-context-check
