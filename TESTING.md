# Testing

Two seams, agreed in the spec (.scratch/mt-tracker/spec.md, Testing Decisions):

- **Seam 1 — the CLI process (primary, highest):** e2e with Gherkin via
  [godog], running the **compiled binary** against a temporary Vault per
  scenario. `$EDITOR` is a fake script for headless editor flows.
  Assertions: stdout, stderr, exit code, files on disk. Every user story
  becomes a scenario.
- **Seam 2 — exported APIs of pure logic:** black-box unit tests of the
  decision-dense packages (datetime parsing, frontmatter round-trip, rank
  renormalization, vault resolution, ID generation, exit codes). Only
  exported symbols; never internals.

No tests reach unexported symbols or assert internal structure. The Cobra
command wiring has no unit coverage gate — e2e behavior covers it.

## Targets

```sh
make check          # the one automation target: unit + e2e + coverage gate + mutation
make build          # bin/mt
make unit           # go test ./internal/... ./cmd/...
make e2e            # godog scenarios against the compiled binary
make coverage-gate  # ≥90% per pure-logic package (scripts/coverage-gate.sh)
make mutate         # gremlins on the pure-logic packages
```

## Layout

```text
cmd/mt/            thin main: os.Exit(cli.Execute())
internal/cli/      cobra wiring — process concerns (args, stdio, exit codes)
internal/exitcode/ pure logic: the exit code convention (0/1/2) and error mapping
e2e/
  main_test.go     TestMain: builds the binary once, runs the godog suite
  features/        *.feature — one scenario per user story
  steps/           step definitions; per-scenario state in the context
  support/         harness helpers: binary, temporary Vault, fake $EDITOR, env isolation
scripts/           coverage-gate.sh
```

## Adding a pure-logic package

1. Put it under `internal/<name>` with exported APIs only.
2. Black-box tests in `internal/<name>/<name>_test.go` (`package <name>_test`).
3. Add `./internal/<name>` to `PURE_PACKAGES` in the `Makefile` — this enrolls
   it in both gates (coverage ≥90% and mutation via gremlins).
4. `make check` must stay green.

## Harness contracts

- Every scenario gets a fresh temporary Vault (`<base>/vault/issues/`), a fake
  `$EDITOR`, and an isolated `XDG_CONFIG_HOME` — nothing leaks between
  scenarios or from the real user config.
- The CLI invokes `$EDITOR <path>` with a single file argument; the fake editor
  writes prepared content to that path, byte for byte.
- Exit codes: `0` success, `1` user error, `2` usage error; errors go to stderr.
- The `I run \`mt …\`` step splits its argument list on whitespace — no shell
  quoting yet. A step needing quoted arguments (e.g. `create "title with
  spaces"`) must grow real tokenization (or a docstring step) before its
  scenario lands.
