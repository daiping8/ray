#!/usr/bin/env bash
# Print the Go packages of this module that compile and test under a plain Go
# toolchain (no bazel), one per line. A package is excluded, mechanically:
#   - when any of its non-test files use cgo (they link the C++ core
#     libraries, which only the bazel build provides), or it transitively
#     depends on such a package;
#   - when any package in its transitive dependency closure fails to load,
#     when any of its own or test files import such a package, or when it has
#     no non-test files at all (a test-only package cannot be built by
#     "go build" standalone). The generated protobuf sources under go/proto
#     exist only as bazel build outputs and are absent from the source tree,
#     so every package that transitively imports them fails to load outside
#     bazel; undeclared or build-tag-only module imports fail the same way.
# The module graph is resolved against a throwaway copy of go.mod, so running
# this script never modifies the working tree.
set -euo pipefail
cd "$(dirname "$0")/.."

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT
cp go.mod "$tmpdir/go.mod"
cp go.sum "$tmpdir/go.sum" 2>/dev/null || true

# Resolve the module graph with -mod=mod against the copy; this fills in the
# missing requirements there instead of failing, and leaves the real go.mod
# untouched. -e keeps packages that fail to load visible in the output.
golist() {
  go list -mod=mod -modfile="$tmpdir/go.mod" -e "$@" ./...
}

# Every package in the dependency graph, marked "bad" when it fails to load
# or has non-test cgo files.
golist -deps -f '{{.ImportPath}}	{{if or .Error (ne (len .CgoFiles) 0)}}bad{{end}}' \
  | awk -F'\t' '$2 == "bad" { print $1 }' > "$tmpdir/excluded"

# Keep the module's own packages that have non-test files, are not excluded
# themselves, and import no excluded package directly — neither from their
# non-test files (p.Deps is their transitive closure) nor from their test
# files, so "go build", "go vet" and "go test" all succeed.
golist -f '{{.ImportPath}}|{{if eq (len .GoFiles) 0}}notest|{{end}}{{join .Deps " "}}|{{join .TestImports " "}}|{{join .XTestImports " "}}' \
  | awk -F'|' -v excluded="$tmpdir/excluded" '
    BEGIN { while ((getline line < excluded) > 0) bad[line] = 1 }
    function clean(list,   n, i) {
      n = split(list, imports, " ")
      for (i = 1; i <= n; i++) if (imports[i] in bad) return 0
      return 1
    }
    $2 == "notest" { next }
    {
      if (!($1 in bad) && clean($2) && clean($3) && clean($4)) print $1
    }'
