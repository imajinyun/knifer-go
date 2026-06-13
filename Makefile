.PHONY: help test test-race coverage-check mod-verify tidy-check diff-check vet arch lint govulncheck check ci-test

GO ?= go
GOLANGCI_LINT ?= golangci-lint
PKGS ?= ./...
COVERAGE_FILE ?= coverage.out

help:
	@echo "Targets:"
	@echo "  test            Run unit tests"
	@echo "  test-race       Run race/shuffle tests and write coverage"
	@echo "  coverage-check  Enforce repository and package coverage gates"
	@echo "  check           Run local stability gates"
	@echo "  ci-test         Run CI test-job gates"

test:
	$(GO) test $(PKGS)

test-race:
	$(GO) test -race -shuffle=on -coverprofile=$(COVERAGE_FILE) $(PKGS)

coverage-check:
	bash bin/check_coverage.sh $(COVERAGE_FILE)

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

check: mod-verify vet arch test-race coverage-check lint govulncheck

ci-test: mod-verify vet tidy-check diff-check arch test-race coverage-check
