#!/usr/bin/env bash
# Post-publish smoke test for github.com/messagebird/bird-sdk-go.
#
# From a throwaway module, fetch the just-tagged version through the Go module
# proxy and compile a program that references the public constructor. This
# proves the tag is resolvable via the proxy and the exported surface compiles
# with no monorepo context — the standalone equivalent of "did the release
# actually produce something installable". Import-only by design: it validates
# packaging, not API calls (a real call would need credentials and a live API).
#
# Usage: smoke.sh <version-without-leading-v>
# Called by: the mirror release workflow after the tag is pushed.
set -euo pipefail
ver="${1:?usage: smoke.sh <version-without-leading-v>}"
mod="github.com/messagebird/bird-sdk-go"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

go mod init birdsmoke >/dev/null

# Just-pushed tags can take a moment to become resolvable through the module
# proxy, so retry the fetch a few times before giving up.
for attempt in 1 2 3 4 5; do
	if GOFLAGS=-trimpath go get "${mod}@v${ver}"; then
		break
	fi
	[ "$attempt" -eq 5 ] && { echo "smoke: ${mod}@v${ver} not resolvable after 5 attempts" >&2; exit 1; }
	echo "smoke: ${mod}@v${ver} not available yet — retrying in 15s"
	sleep 15
done

cat > main.go <<EOF
package main

import (
	"fmt"

	bird "github.com/messagebird/bird-sdk-go"
)

func main() {
	// Reference the public constructor so the compiler proves the symbol and
	// the package are present in the published module; do not call it (no creds).
	_ = bird.NewClient
	fmt.Println("bird-sdk-go v${ver} smoke OK")
}
EOF

GOFLAGS=-trimpath go run .
