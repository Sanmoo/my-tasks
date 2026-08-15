package support_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/e2e/support"
)

func TestTempVaultCreatesIssuesDirectory(t *testing.T) {
	vault := support.TempVault(t)
	info, err := os.Stat(filepath.Join(vault, "issues"))
	if err != nil {
		t.Fatalf("issues/ missing from vault: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("issues/ is not a directory")
	}
}

func TestFakeEditorWritesContentToEditedFile(t *testing.T) {
	dir := t.TempDir()
	editor, err := support.NewFakeEditor(dir, "buffer prepared by the test\n")
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "issue.md")
	out, err := exec.Command(editor.Path, target).CombinedOutput()
	if err != nil {
		t.Fatalf("running fake editor: %v\n%s", err, out)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading edited file: %v", err)
	}
	if string(got) != editor.Content {
		t.Errorf("edited file = %q, want %q", got, editor.Content)
	}
}

func TestFakeEditorWritesContentVerbatim(t *testing.T) {
	// No trailing-newline normalization, and shell-hostile content
	// (single quotes, $, backslashes, heredoc delimiters) survives.
	content := "no final newline, 'quotes', $vars, back\\slash, MT_FAKE_EDITOR_EOF"
	editor, err := support.NewFakeEditor(t.TempDir(), content)
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(t.TempDir(), "issue.md")
	out, err := exec.Command(editor.Path, target).CombinedOutput()
	if err != nil {
		t.Fatalf("running fake editor: %v\n%s", err, out)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading edited file: %v", err)
	}
	if string(got) != content {
		t.Errorf("edited file = %q, want %q", got, content)
	}
}

func TestFakeEditorEnv(t *testing.T) {
	editor, err := support.NewFakeEditor(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	env := editor.Env()
	if len(env) != 1 || env[0] != "EDITOR="+editor.Path {
		t.Errorf("Env() = %v, want [EDITOR=%s]", env, editor.Path)
	}
}

func TestRunCmdCapturesResult(t *testing.T) {
	res, err := support.RunCmd("sh", []string{"-c", "echo out; echo err >&2; exit 3"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", res.ExitCode)
	}
	if res.Stdout != "out\n" {
		t.Errorf("Stdout = %q, want %q", res.Stdout, "out\n")
	}
	if res.Stderr != "err\n" {
		t.Errorf("Stderr = %q, want %q", res.Stderr, "err\n")
	}
}

func TestRunCmdExtraEnvOverridesInherited(t *testing.T) {
	res, err := support.RunCmd("sh", []string{"-c", "printf %s \"$MT_TEST_VAR\""}, []string{"MT_TEST_VAR=overridden"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Stdout != "overridden" {
		t.Errorf("env override lost: stdout = %q", res.Stdout)
	}
	if strings.Contains(res.Stdout, " ") {
		t.Errorf("duplicate variable leaked: stdout = %q", res.Stdout)
	}
}
