// Package support provides the e2e harness shared by every Gherkin
// scenario: the compiled mt binary, temporary Vaults, the fake $EDITOR,
// and a way to run the binary with an isolated environment.
package support

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath is the path of the compiled mt binary, set once by the e2e
// suite's TestMain before any scenario runs.
var binaryPath string

// SetBinary records the path of the compiled mt binary.
func SetBinary(path string) { binaryPath = path }

// Binary returns the path of the compiled mt binary.
func Binary() string { return binaryPath }

// BuildBinary compiles the mt binary into dir and returns its path.
// It locates the module root via "go env GOMOD", so it works from any
// package directory inside the module.
func BuildBinary(dir string) (string, error) {
	gomod, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", fmt.Errorf("locating module root: %w", err)
	}
	root := filepath.Dir(strings.TrimSpace(string(gomod)))
	if root == "." || root == "" {
		return "", errors.New("module root not found: go env GOMOD is empty")
	}
	bin := filepath.Join(dir, "mt")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/mt")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("building mt: %w\n%s", err, out)
	}
	return bin, nil
}

// Result is the observable outcome of one mt run.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// RunCmd runs bin with args and env (extra variables appended to the
// inherited environment, overriding same-named ones), capturing stdout,
// stderr and the exit code.
func RunCmd(bin string, args, env []string) (Result, error) {
	cmd := exec.Command(bin, args...)
	cmd.Env = mergeEnv(env)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return res, fmt.Errorf("running %s: %w", bin, err)
		}
		res.ExitCode = exitErr.ExitCode()
	}
	return res, nil
}

// Run runs the compiled mt binary with args and env, failing the test
// if the process could not be started at all.
func Run(t *testing.T, args, env []string) Result {
	t.Helper()
	res, err := RunCmd(Binary(), args, env)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// mergeEnv returns the inherited environment with the extra variables
// applied last, dropping inherited duplicates so overrides always win.
func mergeEnv(extra []string) []string {
	drop := make(map[string]bool, len(extra))
	for _, e := range extra {
		if k, _, ok := strings.Cut(e, "="); ok {
			drop[k] = true
		}
	}
	env := make([]string, 0, len(os.Environ())+len(extra))
	for _, e := range os.Environ() {
		if k, _, ok := strings.Cut(e, "="); ok && drop[k] {
			continue
		}
		env = append(env, e)
	}
	return append(env, extra...)
}

// NewVaultIn creates a fresh Vault skeleton (a directory with an
// issues/ subdirectory) under base and returns the Vault path. The
// caller owns cleanup of base.
func NewVaultIn(base string) (string, error) {
	vault := filepath.Join(base, "vault")
	if err := os.MkdirAll(filepath.Join(vault, "issues"), 0o755); err != nil {
		return "", fmt.Errorf("creating temporary vault: %w", err)
	}
	return vault, nil
}

// TempVault creates a fresh Vault skeleton in a test-owned temporary
// directory and returns the Vault path.
func TempVault(t *testing.T) string {
	t.Helper()
	vault, err := NewVaultIn(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return vault
}

// FakeEditor is a script that plays the role of the user's $EDITOR in
// headless scenarios. mt invokes $EDITOR <path>; this script writes
// Content to that path, byte for byte.
type FakeEditor struct {
	// Path is the location of the fake editor script.
	Path string
	// Content is what the script writes into the file mt opens.
	Content string
}

// NewFakeEditor creates an executable fake editor script in dir.
func NewFakeEditor(dir, content string) (*FakeEditor, error) {
	path := filepath.Join(dir, "fake-editor.sh")
	// printf '%s' writes the content verbatim — no trailing-newline
	// normalization, no heredoc delimiter collisions.
	quoted := "'" + strings.ReplaceAll(content, "'", "'\\''") + "'"
	script := "#!/bin/sh\nprintf '%s' " + quoted + " > \"$1\"\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		return nil, fmt.Errorf("creating fake editor: %w", err)
	}
	return &FakeEditor{Path: path, Content: content}, nil
}

// EditorVar returns the EDITOR variable pointing at the script, for
// building an environment by hand.
func (e *FakeEditor) EditorVar() string {
	return "EDITOR=" + e.Path
}

// Env returns the EDITOR variable pointing at the script, ready to be
// passed to Run/RunCmd.
func (e *FakeEditor) Env() []string {
	return []string{e.EditorVar()}
}
