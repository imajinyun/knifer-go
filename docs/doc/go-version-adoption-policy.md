# Go Version Adoption Policy

`knifer-go` requires Go 1.26 or later. The module and several direct dependencies
require this language level. The version-independent configuration in
`ai-context.json.ci_workflows.tool_versions` is the source of toolchain pins;
the checkers compare that policy with `go.mod`, dependencies and workflows.

| Decision | Current value | Validation |
| --- | --- | --- |
| Minimum language version | Go 1.26 (`go 1.26.0`) | Module directive and dependency Go requirements must agree. |
| CI compatibility matrix | Go 1.26.8 and Go 1.27.1 | Each fixed toolchain runs with `GOTOOLCHAIN=local`. |
| Release and tool-build version | Go 1.27.1 | Release, lint and governance jobs explicitly install and record this version. |
| Linter | golangci-lint v2.14.0, built with Go 1.27.1 | The lint gate checks the tool version and compiler support for the active Go line. |
| Downgrade | Not supported today | Older Go support would require compatible dependency versions and full revalidation. |

The release uses the newer tested Go line; minimum-version compatibility is
verified independently. Automatic toolchain downloads during validation are
disabled, so a green minimum-version job cannot silently mean a newer compiler
was used. Go installation itself is an explicit provisioning step.

Public `v*` API compatibility remains a separate contract from minimum Go support.
The minimum changed through dependency updates; `testing.B.Loop`, available since
Go 1.24, is not the reason for requiring Go 1.26. New benchmarks may use it.

## Local validation

Put the chosen, already installed Go binary on PATH and run:

```bash
GOTOOLCHAIN=local go version
make toolchain-check
make ai-context-check
make ci-workflow-check
make governance-maturity-check
make full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out
```

Clear an inherited `GOROOT` before switching Go installations: the selected
binary must use its own standard library and compiler directory. Separate
build caches can also make multi-toolchain diagnostics easier to interpret.

`make release-check` additionally enforces all package coverage thresholds,
all discovered Fuzz targets and workflow consistency. Governance Make targets
force `GOTOOLCHAIN=local`. The toolchain check records the actual Go version and
rejects dependency requirements above the module minimum. The lint check rejects
a linter compiled with an older Go language line than the active analysis toolchain.

## Updating the policy

Update the generic minimum/release/test pins and corresponding workflow values,
adoption metadata, README requirements and release notes together. Version checks
must validate relationships instead of requiring a historical minor version.
Test fixtures cover mismatched module minimums, unpinned versions, missing minimum
matrix coverage, implicit switching, missing setup and outdated linter builds.
