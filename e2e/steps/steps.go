// Package steps maps Gherkin steps onto the e2e harness helpers.
// Every scenario gets a fresh temporary Vault, a fake $EDITOR, and an
// isolated XDG_CONFIG_HOME — the seams later scenarios build on.
package steps

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/Sanmoo/my-tasks2/e2e/support"
)

type stateKey struct{}

type state struct {
	// base is the per-scenario scratch directory; everything else lives
	// under it and dies with it.
	base string
	// vault is a fresh temporary Vault, ready for any scenario.
	vault string
	// editor is the fake $EDITOR every run uses.
	editor *support.FakeEditor
	// dir is the working directory of the following mt runs ("" =
	// inherit the test process cwd).
	dir string
	// result is the outcome of the last mt run.
	result support.Result
	// ran is true once a run has been recorded in this scenario.
	ran bool
}

// env is the baseline environment for every mt run in the scenario:
// fake $EDITOR plus an isolated config home, so the real user config
// can never leak into a scenario.
func (st *state) env() []string {
	return []string{
		st.editor.EditorVar(),
		"XDG_CONFIG_HOME=" + filepath.Join(st.base, "config"),
	}
}

// expand substitutes the per-scenario path placeholders in a step
// argument: <base> is the scenario scratch directory, <vault> the
// scenario's temporary vault.
func (st *state) expand(arg string) string {
	arg = strings.ReplaceAll(arg, "<base>", st.base)
	arg = strings.ReplaceAll(arg, "<vault>", st.vault)
	return arg
}

func (st *state) requireResult() error {
	if !st.ran {
		return errors.New("no mt run recorded in this scenario yet")
	}
	return nil
}

// InitializeScenario registers the hooks and step definitions of the e2e suite.
func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		base, err := os.MkdirTemp("", "mt-e2e-")
		if err != nil {
			return ctx, fmt.Errorf("creating scenario scratch dir: %w", err)
		}
		vault, err := support.NewVaultIn(base)
		if err != nil {
			os.RemoveAll(base)
			return ctx, err
		}
		editor, err := support.NewFakeEditor(base, "")
		if err != nil {
			os.RemoveAll(base)
			return ctx, err
		}
		st := &state{base: base, vault: vault, editor: editor}
		return context.WithValue(ctx, stateKey{}, st), nil
	})

	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		if st, ok := ctx.Value(stateKey{}).(*state); ok {
			os.RemoveAll(st.base)
		}
		return ctx, nil
	})

	sc.Step(`^I run \x60mt(?: (.*))?\x60$`, iRunMt)
	sc.Step(`^the working directory is "([^"]*)"$`, workingDirectoryIs)
	sc.Step(`^the exit code is (\d+)$`, exitCodeIs)
	sc.Step(`^stdout contains "([^"]*)"$`, stdoutContains)
	sc.Step(`^stderr contains "([^"]*)"$`, stderrContains)
	sc.Step(`^a temporary vault exists$`, temporaryVaultExists)
	sc.Step(`^the vault contains an issues directory$`, vaultHasIssuesDirectory)
	sc.Step(`^a fake editor is available$`, fakeEditorAvailable)
	sc.Step(`^the file "([^"]*)" exists$`, fileExists)
	sc.Step(`^the directory "([^"]*)" exists$`, dirExists)
	sc.Step(`^the file "([^"]*)" contains "([^"]*)"$`, fileContains)
}

func stateFrom(ctx context.Context) (*state, error) {
	st, ok := ctx.Value(stateKey{}).(*state)
	if !ok {
		return nil, errors.New("scenario state missing: was InitializeScenario registered?")
	}
	return st, nil
}

func iRunMt(ctx context.Context, args string) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	args = st.expand(args)
	res, err := support.RunCmdIn(support.Binary(), st.dir, strings.Fields(args), st.env())
	if err != nil {
		return ctx, err
	}
	st.result = res
	st.ran = true
	return ctx, nil
}

// workingDirectoryIs sets the working directory of the following mt
// runs. The directory is created when missing.
func workingDirectoryIs(ctx context.Context, dir string) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	dir = st.expand(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ctx, fmt.Errorf("creating working directory %q: %w", dir, err)
	}
	st.dir = dir
	return ctx, nil
}

func exitCodeIs(ctx context.Context, want string) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	if err := st.requireResult(); err != nil {
		return ctx, err
	}
	wantCode, err := strconv.Atoi(want)
	if err != nil {
		return ctx, fmt.Errorf("parsing expected exit code %q: %w", want, err)
	}
	if st.result.ExitCode != wantCode {
		return ctx, fmt.Errorf(
			"exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			st.result.ExitCode, wantCode, st.result.Stdout, st.result.Stderr)
	}
	return ctx, nil
}

func stdoutContains(ctx context.Context, want string) (context.Context, error) {
	return streamContains(ctx, "stdout", want)
}

func stderrContains(ctx context.Context, want string) (context.Context, error) {
	return streamContains(ctx, "stderr", want)
}

func streamContains(ctx context.Context, stream, want string) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	if err := st.requireResult(); err != nil {
		return ctx, err
	}
	var got string
	if stream == "stdout" {
		got = st.result.Stdout
	} else {
		got = st.result.Stderr
	}
	if !strings.Contains(got, want) {
		return ctx, fmt.Errorf("%s does not contain %q\n%s:\n%s", stream, want, stream, got)
	}
	return ctx, nil
}

func temporaryVaultExists(ctx context.Context) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	if st.vault == "" {
		return ctx, errors.New("no temporary vault was created for this scenario")
	}
	if info, err := os.Stat(st.vault); err != nil || !info.IsDir() {
		return ctx, fmt.Errorf("temporary vault %q is not a directory", st.vault)
	}
	return ctx, nil
}

func vaultHasIssuesDirectory(ctx context.Context) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	if info, err := os.Stat(filepath.Join(st.vault, "issues")); err != nil || !info.IsDir() {
		return ctx, fmt.Errorf("vault %q has no issues/ directory", st.vault)
	}
	return ctx, nil
}

func fakeEditorAvailable(ctx context.Context) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	if st.editor == nil || st.editor.Path == "" {
		return ctx, errors.New("no fake editor was created for this scenario")
	}
	info, err := os.Stat(st.editor.Path)
	if err != nil {
		return ctx, fmt.Errorf("fake editor script: %w", err)
	}
	if info.Mode()&0o111 == 0 {
		return ctx, fmt.Errorf("fake editor %q is not executable", st.editor.Path)
	}
	return ctx, nil
}

func fileExists(ctx context.Context, path string) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	path = st.expand(path)
	if _, err := os.Stat(path); err != nil {
		return ctx, fmt.Errorf("file %q does not exist", path)
	}
	return ctx, nil
}

func dirExists(ctx context.Context, path string) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	path = st.expand(path)
	info, err := os.Stat(path)
	if err != nil {
		return ctx, fmt.Errorf("directory %q does not exist", path)
	}
	if !info.IsDir() {
		return ctx, fmt.Errorf("%q is not a directory", path)
	}
	return ctx, nil
}

func fileContains(ctx context.Context, path, want string) (context.Context, error) {
	st, err := stateFrom(ctx)
	if err != nil {
		return ctx, err
	}
	path = st.expand(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return ctx, fmt.Errorf("reading %q: %w", path, err)
	}
	if !strings.Contains(string(data), want) {
		return ctx, fmt.Errorf("%q does not contain %q:\n%s", path, want, data)
	}
	return ctx, nil
}
