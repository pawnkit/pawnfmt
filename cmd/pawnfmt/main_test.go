package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCLIFixture(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runCLI(args []string, stdin string) (code int, stdout, stderr string) {
	var out, errBuf bytes.Buffer

	code = run(args, strings.NewReader(stdin), &out, &errBuf)

	return code, out.String(), errBuf.String()
}

func TestRunDefaultModePrintsFormattedOutputForOneFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new   x=1;\n")

	code, stdout, stderr := runCLI([]string{path}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if stdout != "new x = 1;\n" {
		t.Fatalf("stdout = %q, want %q", stdout, "new x = 1;\n")
	}
}

func TestRunWriteFlagWritesTheFileInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new   x=1;\n")

	code, stdout, stderr := runCLI([]string{"-w", path}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if stdout != "" {
		t.Fatalf("-w should not print to stdout, got %q", stdout)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file after -w: %v", err)
	}

	if string(got) != "new x = 1;\n" {
		t.Fatalf("file content after -w = %q, want %q", got, "new x = 1;\n")
	}
}

func TestRunCheckFlagReportsChangesWithoutWriting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new   x=1;\n")

	code, stdout, _ := runCLI([]string{"--check", path}, "")
	if code != exitCheckChanges {
		t.Fatalf("exit code = %d, want %d (exitCheckChanges)", code, exitCheckChanges)
	}

	if !strings.Contains(stdout, path) {
		t.Fatalf("-check stdout = %q, want it to name the changed file %s", stdout, path)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file after -check: %v", err)
	}

	if string(got) != "new   x=1;\n" {
		t.Fatalf("-check must not modify the file, got %q", got)
	}
}

func TestRunCheckFlagExitsZeroWhenAlreadyFormatted(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new x = 1;\n")

	code, _, stderr := runCLI([]string{"--check", path}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d for an already-formatted file; stderr:\n%s", code, exitOK, stderr)
	}
}

func TestRunDiffFlagPrintsAUnifiedDiff(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new   x=1;\n")

	code, stdout, _ := runCLI([]string{"--diff", path}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d", code, exitOK)
	}

	if !strings.Contains(stdout, "--- "+path) || !strings.Contains(stdout, "+++ "+path) {
		t.Fatalf("-diff stdout missing file headers:\n%s", stdout)
	}
}

func TestRunDiffFlagCanForceColour(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new   x=1;\n")

	code, stdout, _ := runCLI([]string{"--diff", "--color=always", path}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d", code, exitOK)
	}

	if !strings.Contains(stdout, "\x1b[31m") || !strings.Contains(stdout, "\x1b[32m") {
		t.Fatalf("forced colour diff missing ANSI colour codes:\n%s", stdout)
	}
}

func TestRunMultipleFilesWithoutWriteCheckOrDiffIsAnError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.pwn")
	b := filepath.Join(dir, "b.pwn")

	writeCLIFixture(t, a, "new   x=1;\n")
	writeCLIFixture(t, b, "new   y=2;\n")

	code, _, stderr := runCLI([]string{a, b}, "")
	if code != exitFormatError {
		t.Fatalf("exit code = %d, want %d (exitFormatError) when formatting 2 files with no -w/-check/-diff", code, exitFormatError)
	}

	if !strings.Contains(stderr, "--write") {
		t.Fatalf("stderr should suggest --write/--check/--diff:\n%s", stderr)
	}
}

func TestRunWriteAndCheckTogetherIsRejected(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI([]string{"-w", "--check", "whatever.pwn"}, "")
	if code != exitConfigError {
		t.Fatalf("exit code = %d, want %d (exitConfigError)", code, exitConfigError)
	}

	if !strings.Contains(stderr, "--write and --check") {
		t.Fatalf("stderr should explain the conflict:\n%s", stderr)
	}
}

func TestRunStdinModeFormatsAndWritesToStdout(t *testing.T) {
	t.Parallel()

	code, stdout, stderr := runCLI([]string{"--stdin"}, "new   x=1;\n")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if stdout != "new x = 1;\n" {
		t.Fatalf("stdin mode stdout = %q, want %q", stdout, "new x = 1;\n")
	}
}

func TestRunStdinCombinedWithPathsIsRejected(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI([]string{"--stdin", "a.pwn"}, "")
	if code != exitConfigError {
		t.Fatalf("exit code = %d, want %d (exitConfigError)", code, exitConfigError)
	}

	if !strings.Contains(stderr, "--stdin cannot be combined") {
		t.Fatalf("stderr should explain the conflict:\n%s", stderr)
	}
}

func TestRunNoInputAtAllIsAnError(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI(nil, "")
	if code != exitConfigError {
		t.Fatalf("exit code = %d, want %d (exitConfigError)", code, exitConfigError)
	}

	if !strings.Contains(stderr, "no input") {
		t.Fatalf("stderr should explain no input was given:\n%s", stderr)
	}
}

func TestRunUnknownFlagIsAConfigError(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI([]string{"--not-a-real-flag"}, "")
	if code != exitConfigError {
		t.Fatalf("exit code = %d, want %d (exitConfigError)", code, exitConfigError)
	}

	if stderr == "" {
		t.Fatal("stderr should explain the unknown flag")
	}
}

func TestRunHelpFlagExitsOK(t *testing.T) {
	t.Parallel()

	code, _, _ := runCLI([]string{"--help"}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d for -help", code, exitOK)
	}
}

func TestRunNoSuchFileIsAFormatError(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI([]string{filepath.Join(t.TempDir(), "missing.pwn")}, "")
	if code != exitFormatError {
		t.Fatalf("exit code = %d, want %d (exitFormatError)", code, exitFormatError)
	}

	if stderr == "" {
		t.Fatal("stderr should report the missing file")
	}
}

func TestRunPrintConfigPrintsResolvedTOML(t *testing.T) {
	t.Parallel()

	code, stdout, stderr := runCLI([]string{"--print-config", "--no-config"}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "line_width") {
		t.Fatalf("-print-config stdout missing expected key:\n%s", stdout)
	}
}

func TestRunInitConfigWritesAFileAndRefusesToOverwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "pawnfmt.toml")

	code, stdout, stderr := runCLI([]string{"--init-config", target}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, target) {
		t.Fatalf("stdout should confirm the written path:\n%s", stdout)
	}

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("init-config did not create %s: %v", target, err)
	}

	code2, _, stderr2 := runCLI([]string{"--init-config", target}, "")
	if code2 != exitConfigError {
		t.Fatalf("second -init-config exit code = %d, want %d (should refuse to overwrite)", code2, exitConfigError)
	}

	if !strings.Contains(stderr2, "already exists") {
		t.Fatalf("stderr should explain the file already exists:\n%s", stderr2)
	}
}

func TestRunDebugTokensPrintsTokenStream(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new x;\n")

	code, stdout, stderr := runCLI([]string{"--debug-tokens", path}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if stdout == "" {
		t.Fatal("-debug-tokens should print something")
	}
}

func TestRunDebugCSTOnMultipleFilesIsRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.pwn")
	b := filepath.Join(dir, "b.pwn")

	writeCLIFixture(t, a, "new x;\n")
	writeCLIFixture(t, b, "new y;\n")

	code, _, stderr := runCLI([]string{"--debug-cst", a, b}, "")
	if code != exitConfigError {
		t.Fatalf("exit code = %d, want %d (exitConfigError)", code, exitConfigError)
	}

	if !strings.Contains(stderr, "exactly one input file") {
		t.Fatalf("stderr should explain debug modes need exactly one file:\n%s", stderr)
	}
}
