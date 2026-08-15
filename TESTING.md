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
internal/vault/    pure logic: global config (bookmarks + default, XDG, add/
                   remove/list round-trip), vault config (mt.yaml: prefix,
                   status), vault resolution (@bookmark > --vault > default),
                   @-token extraction, ~ expansion, ID-prefix derivation
internal/issue/    pure logic: the Issue frontmatter round-trip (stable field
                   order, optional fields only-when-set, no id/updated_at)
                   and ID generation (prefix + short random suffix, collision
                   retry)
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
- The `I run \`mt …\`` step splits its argument list shell-style: double and
  single quotes group an argument, so `create "title with spaces"` passes one
  title argument. No globbing or variable expansion.
- Step arguments may use the per-scenario placeholders `<base>` (scratch dir),
  `<vault>` (temporary vault) and `<id>` (the issue ID remembered by the
  `I remember the issue ID` step — the trailing stdout token of the last run);
  they are expanded before use, including in the `contains`/`does not contain`
  assertion arguments.
- Assertion steps available beyond the basics: `stdout matches "<regex>"`,
  `stdout does not contain "…"`, `the file "…" matches "<regex>"`, `the file
  "…" does not contain "…"`, `the directory "…" contains <n> files`, and the
  docstring step `the fake editor writes` (replaces the fake $EDITOR's content
  for editor-flow scenarios).
- Scenarios prepare files with the docstring step
  `the file "<path>" is written with:` (the following indented block is the
  file content, parent directories are created).

## Vault addressing convention

- Vault-requiring commands address the vault by `@bookmark`, `--vault <path>`,
  or the default bookmark in the global config (`@bookmark` > `--vault` >
  `default`); with none, the command fails with instructions (exit 1).
- The `@bookmark` token (regex `@[A-Za-z0-9_-]+`) may appear anywhere among a
  command's positional arguments and is extracted before command parsing;
  at most one is allowed, and bookmark names must not start with `@`.
- The global config lives at `$XDG_CONFIG_HOME/mt/config.yaml`, or
  `$HOME/.config/mt/config.yaml`; the vault config is `mt.yaml` at the vault
  root.
- `mt bookmark add <name> <path>` / `list` / `rm <name>` manage the global
  config's bookmarks and default. Names are bare (letters, digits, '-' and
  '_', no leading '@'). `list` marks the default with `(default)`, and `rm`
  clears the default when it removes the default bookmark.
- `mt init [dir]` derives the ID prefix from the directory name when
  `--prefix` is omitted (`PrefixFor`: lowercased, non-alphanumeric stripped,
  ≤8 chars) and refuses to overwrite an existing vault config.
