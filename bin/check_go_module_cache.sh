#!/usr/bin/env bash
set -euo pipefail

go_cmd="${GO:-go}"
go_cache="$("$go_cmd" env GOCACHE)"
if [ -z "${go_cache}" ]; then
	echo "GO MODULE CACHE CHECK ERROR: go env GOCACHE returned an empty path." >&2
	echo "Agent and governance targets use /tmp/knifer-go-gocache by default; set GOCACHE to override it." >&2
	exit 1
fi
if ! mkdir -p "${go_cache}" 2>/dev/null; then
	echo "GO MODULE CACHE CHECK ERROR: GOCACHE is not creatable: ${go_cache}" >&2
	echo "Set GOCACHE=/tmp/knifer-go-gocache or choose another writable cache path." >&2
	exit 1
fi
probe="${go_cache}/.knifer-go-cache-write-test"
if ! printf 'ok\n' >"${probe}" 2>/dev/null; then
	echo "GO MODULE CACHE CHECK ERROR: GOCACHE is not writable: ${go_cache}" >&2
	echo "Set GOCACHE=/tmp/knifer-go-gocache or choose another writable cache path." >&2
	exit 1
fi
if ! grep -q '^ok$' "${probe}" 2>/dev/null; then
	echo "GO MODULE CACHE CHECK ERROR: GOCACHE is not readable after write: ${go_cache}" >&2
	echo "Set GOCACHE=/tmp/knifer-go-gocache or choose another writable cache path." >&2
	rm -f "${probe}" 2>/dev/null || true
	exit 1
fi
rm -f "${probe}" 2>/dev/null || true

packages=(
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/makiuchi-d/gozxing"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/text/encoding"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

failed=()
for pkg in "${packages[@]}"; do
	if ! output="$(GOWORK=off "$go_cmd" list -mod=mod -f '{{.ImportPath}}' "$pkg" 2>&1)"; then
		failed+=("$pkg"$'\n'"$output")
	fi
done

if ((${#failed[@]} > 0)); then
	echo "GO MODULE CACHE CHECK ERROR: required module packages could not be resolved." >&2
	echo >&2
	"$go_cmd" env GOVERSION GOTOOLCHAIN GOWORK GOMOD GOMODCACHE GOCACHE GOPROXY >&2 || true
	echo "check GOWORK=off" >&2
	echo >&2
	echo "This usually means the module cache is stale or partially corrupted, even if go.mod and go.sum are valid." >&2
	echo "Try one of these before rerunning validation:" >&2
	echo "  GOMODCACHE=\$("$go_cmd" env GOMODCACHE) $go_cmd clean -modcache" >&2
	echo "  GOMODCACHE=/private/tmp/knifer-go-\$(date +%s)-gomodcache GOCACHE=/private/tmp/knifer-go-\$(date +%s)-gocache make quick-check" >&2
	echo >&2
	for failure in "${failed[@]}"; do
		echo "---" >&2
		echo "$failure" >&2
	done
	exit 1
fi

echo "go module cache package resolution is healthy"
