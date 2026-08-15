GO ?= go

# Packages of pure logic (Seam 2): black-box unit tested, with a per-package
# coverage gate and mutation testing. Command wiring is covered by e2e
# behavior instead and stays out of these gates. Add new pure-logic
# packages here as they land.
PURE_PACKAGES := ./internal/exitcode ./internal/vault ./internal/issue
COVERAGE_THRESHOLD := 90

.PHONY: check build unit e2e coverage-gate mutate

# check is the single automation target: unit + e2e + coverage gate + mutation.
check: unit e2e coverage-gate mutate

build:
	$(GO) build -o bin/mt ./cmd/mt

unit:
	$(GO) test ./internal/... ./cmd/...

e2e:
	$(GO) test ./e2e/...

coverage-gate:
	./scripts/coverage-gate.sh $(COVERAGE_THRESHOLD) $(PURE_PACKAGES)

# Mutation thresholds live in .gremlins.yaml: the float64 CLI flags are
# broken in gremlins v0.6.0 (viper reads them as strings, so they never
# gate). The config file is auto-discovered from the module root.
# `gremlins unleash` accepts a single package path, so expand one
# invocation per package (each with its own failure gate).
.SILENT: mutate

mutate:
	$(foreach pkg,$(PURE_PACKAGES),gremlins unleash $(pkg) || exit 1;)
