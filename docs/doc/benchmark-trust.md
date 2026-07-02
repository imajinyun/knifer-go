# Benchmark Trust Guide

Benchmark output in `knifer-go` is runtime evidence, not a universal performance
claim. Use this page to decide which benchmark command belongs in quick gates,
which command is manual opt-in evidence, and what proof is required before
publishing a performance statement.

## Quick Gates

| Command | Scope | Rule |
| --- | --- | --- |
| `make bench-smoke` | Short benchmark health check for CI and release readiness. | Keep quick gates deterministic and bounded. |
| `make bench-regression-check` | Metadata validation for tracked packages, thresholds, and benchstat policy. | This checks benchmark governance without running long benchmarks. |

Quick gates must use deterministic inputs, bounded runtime, local resources, and
no network-dependent workload. They should prove benchmark suites still run and
that benchmark metadata is coherent.

## Manual Opt-In Evidence

| Command | Use when | Rule |
| --- | --- | --- |
| `make bench-core` | Measuring core internal packages before changing hot paths. | Treat output as a local baseline unless repeated runs and benchstat show a change. |
| `make bench-facade` | Measuring public facade overhead or helper trade-offs. | Include package list, Go version, command, count, time, and input shape near any claim. |
| `make bench-codec` | Measuring codec and serialization helpers. | Compare against the direct standard-library path when relevant. |
| `make bench-baseline` | Saving repeated historical benchmark output outside the repository. | Use only when a future comparison is expected. |
| `make bench-compare` | Comparing saved and current repeated benchmark output. | Performance claims require repeated runs and benchstat. |

Manual opt-in benchmarks may be longer, workload-specific, or useful only for a
focused change. They should not run in the fast local gate unless they are made
deterministic, bounded, and cheap enough for repeated developer use.

## Publishing Rules

- benchmark output is local baseline evidence.
- performance claims require repeated runs and benchstat.
- quick gates are deterministic and bounded.
- long or workload-specific benchmarks are manual opt-in.
- do not publish universal performance claims.
- Document workload shape, Go version, OS/architecture, command, run count, and
  relevant environment assumptions beside any published benchmark result.
