.PHONY: help test test-race coverage-profile coverage-report coverage-check api-check mod-verify tidy-check diff-check vet arch lint govulncheck security-check bench bench-core bench-facade bench-smoke check ci-test

GO ?= go
GOLANGCI_LINT ?= golangci-lint
PKGS ?= ./...
COVERAGE_FILE ?= coverage.out
BENCH ?= .
BENCH_PKGS ?= ./internal/slice ./internal/maps ./internal/str ./internal/num
BENCH_FACADE_PKGS ?= ./vslice ./vmap ./vstr ./vnum
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
	@echo "  bench-smoke     Run a short core benchmark smoke check"
	@echo "  security-check  Run lint and govulncheck"
	@echo "  api-check       Verify exported API snapshot is current"
	@echo "  check           Run local stability gates"
	@echo "  ci-test         Run CI test-job gates"

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

mod-verify:
	$(GO) mod verify

tidy-check:
	$(GO) mod tidy
	git diff --exit-code -- go.mod go.sum

diff-check:
	git diff --check
	git diff --exit-code

vet:
	$(GO) vet $(PKGS)

arch:
	bash bin/check_arch.sh

lint:
	$(GOLANGCI_LINT) run $(PKGS)

govulncheck:
	$(GO) tool govulncheck $(PKGS)

security-check: lint govulncheck

bench:
	$(GO) test -bench=$(BENCH) -benchmem -benchtime=$(BENCHTIME) -count=$(BENCHCOUNT) -run=^$$ $(BENCH_PKGS)

bench-core: bench

bench-facade:
	$(GO) test -bench=$(BENCH) -benchmem -benchtime=$(BENCHTIME) -count=$(BENCHCOUNT) -run=^$$ $(BENCH_FACADE_PKGS)

bench-smoke:
	$(GO) test -bench=Benchmark -benchtime=100ms -count=1 -run=^$$ $(BENCH_PKGS) $(BENCH_FACADE_PKGS)

check: mod-verify vet arch test-race coverage-check api-check lint govulncheck

ci-test: mod-verify vet tidy-check diff-check arch test-race coverage-check api-check
