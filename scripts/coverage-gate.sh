#!/bin/sh
# coverage-gate.sh — enforces the Seam-2 coverage gate: every listed
# package must reach the threshold on its own (no global percentage).
#
# Usage: coverage-gate.sh <threshold> <package>...
set -eu

threshold="${1:-90}"
shift
if [ "$#" -eq 0 ]; then
	echo "usage: coverage-gate.sh <threshold> <package>..." >&2
	exit 2
fi

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

failed=0
i=0
for pkg in "$@"; do
	i=$((i + 1))
	out="$tmpdir/cov.$i.out"
	if ! go test -coverprofile="$out" -covermode=atomic "$pkg"; then
		echo "FAIL: tests failed for $pkg" >&2
		failed=1
		continue
	fi
	cov=$(go tool cover -func="$out" | awk '$1 == "total:" { gsub(/%/, "", $3); print $3 }')
	if [ -z "$cov" ]; then
		echo "FAIL: no coverage data for $pkg" >&2
		failed=1
		continue
	fi
	if awk -v c="$cov" -v t="$threshold" 'BEGIN { exit !(c >= t) }'; then
		echo "PASS: $pkg ${cov}% >= ${threshold}%"
	else
		echo "FAIL: $pkg ${cov}% < ${threshold}%" >&2
		failed=1
	fi
done

exit "$failed"
