package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunExplicitConfigPathIsApplied(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "custom.toml")
	writeCLIFixture(t, cfgPath, "indent_width = 2\n")

	srcPath := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, srcPath, "stock F() {\n\tnew x;\n}\n")

	code, stdout, stderr := runCLI([]string{"--config", cfgPath, srcPath}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "  new x;") || strings.Contains(stdout, "    new x;") {
		t.Fatalf("explicit -config indent_width=2 was not applied:\n%s", stdout)
	}
}

func TestRunExplicitConfigPathThatDoesNotExistIsAConfigError(t *testing.T) {
	t.Parallel()

	code, _, stderr := runCLI([]string{"--config", filepath.Join(t.TempDir(), "missing.toml"), "a.pwn"}, "")
	if code != exitConfigError {
		t.Fatalf("exit code = %d, want %d (exitConfigError)", code, exitConfigError)
	}

	if stderr == "" {
		t.Fatal("stderr should explain the missing config file")
	}
}

func TestRunNoConfigFlagIgnoresADiscoverableConfigFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeCLIFixture(t, filepath.Join(dir, "pawnfmt.toml"), "indent_width = 2\n")
	srcPath := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, srcPath, "stock F() {\n\tnew x;\n}\n")

	code, stdout, stderr := runCLI([]string{"--no-config", srcPath}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "    new x;") {
		t.Fatalf("-no-config should use the default indent_width=4, not the discoverable config's 2:\n%s", stdout)
	}
}

func TestRunDiscoversNearestConfigFileAutomatically(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeCLIFixture(t, filepath.Join(dir, "pawnfmt.toml"), "indent_width = 2\n")
	srcPath := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, srcPath, "stock F() {\n\tnew x;\n}\n")

	code, stdout, stderr := runCLI([]string{srcPath}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "  new x;") || strings.Contains(stdout, "    new x;") {
		t.Fatalf("automatically discovered pawnfmt.toml (indent_width=2) was not applied:\n%s", stdout)
	}
}

func TestRunDiscoversJSONConfigFileAutomatically(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeCLIFixture(t, filepath.Join(dir, "pawnfmt.json"), "{\"indent_width\": 2}\n")
	srcPath := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, srcPath, "stock F() {\n\tnew x;\n}\n")

	code, stdout, stderr := runCLI([]string{srcPath}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}
	if !strings.Contains(stdout, "  new x;") || strings.Contains(stdout, "    new x;") {
		t.Fatalf("automatically discovered pawnfmt.json was not applied:\n%s", stdout)
	}
}

func TestRunDiscoversConfigIndependentlyForEachFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	left := filepath.Join(root, "left")
	right := filepath.Join(root, "right")
	writeCLIFixture(t, filepath.Join(left, "pawnfmt.toml"), "indent_width = 2\n")
	writeCLIFixture(t, filepath.Join(right, "pawnfmt.toml"), "indent_width = 6\n")

	leftPath := filepath.Join(left, "a.pwn")
	rightPath := filepath.Join(right, "b.pwn")
	source := "stock F() {\n\tnew x;\n}\n"
	writeCLIFixture(t, leftPath, source)
	writeCLIFixture(t, rightPath, source)

	code, _, stderr := runCLI([]string{"--write", leftPath, rightPath}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	leftFormatted, err := os.ReadFile(leftPath)
	if err != nil {
		t.Fatal(err)
	}
	rightFormatted, err := os.ReadFile(rightPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(leftFormatted), "\n  new x;") {
		t.Fatalf("left file did not use its indent_width=2 config:\n%s", leftFormatted)
	}
	if !strings.Contains(string(rightFormatted), "\n      new x;") {
		t.Fatalf("right file did not use its indent_width=6 config:\n%s", rightFormatted)
	}
}

func TestRunReportsInvalidNestedConfigAsConfigError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	good := filepath.Join(root, "good")
	bad := filepath.Join(root, "bad")
	writeCLIFixture(t, filepath.Join(good, "pawnfmt.toml"), "indent_width = 2\n")
	writeCLIFixture(t, filepath.Join(bad, "pawnfmt.toml"), "unknown_option = true\n")
	goodPath := filepath.Join(good, "a.pwn")
	badPath := filepath.Join(bad, "b.pwn")
	writeCLIFixture(t, goodPath, "new x;\n")
	writeCLIFixture(t, badPath, "new y;\n")

	code, _, stderr := runCLI([]string{"--check", goodPath, badPath}, "")
	if code != exitConfigError {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitConfigError, stderr)
	}
	if !strings.Contains(stderr, "unknown_option") {
		t.Fatalf("stderr should identify the invalid nested config:\n%s", stderr)
	}
}

func TestRunStdinFilenameDrivesConfigDiscoveryForStdinMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeCLIFixture(t, filepath.Join(dir, "pawnfmt.toml"), "indent_width = 2\n")
	stdinFilename := filepath.Join(dir, "virtual.pwn")

	code, stdout, stderr := runCLI([]string{"--stdin", "--stdin-filename", stdinFilename}, "stock F() {\n\tnew x;\n}\n")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "  new x;") || strings.Contains(stdout, "    new x;") {
		t.Fatalf("-stdin-filename should drive config discovery to the sibling pawnfmt.toml (indent_width=2):\n%s", stdout)
	}
}

func TestRunDebugFormatDocPrintsTheDocTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.pwn")
	writeCLIFixture(t, path, "new x;\n")

	code, stdout, stderr := runCLI([]string{"--debug-format-doc", path}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "Concat") {
		t.Fatalf("-debug-format-doc should print the doc tree:\n%s", stdout)
	}
}

func TestRunDebugFormatDocReportsUnparseableSourceAsAFormatError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.pwn")
	writeCLIFixture(t, path, "}")

	code, _, stderr := runCLI([]string{"--debug-format-doc", path}, "")
	if code != exitFormatError {
		t.Fatalf("exit code = %d, want %d (exitFormatError)", code, exitFormatError)
	}

	if stderr == "" {
		t.Fatal("stderr should explain the parse failure")
	}
}

func TestStartDirForPrefersTheDirectoryOfTheFirstPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "a.pwn")
	writeCLIFixture(t, path, "new x;\n")

	got := startDirFor(&options{Paths: []string{path}})

	want, _ := filepath.Abs(filepath.Dir(path))
	if got != want {
		t.Fatalf("startDirFor(file path) = %q, want %q", got, want)
	}
}

func TestStartDirForUsesADirectoryPathDirectly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	got := startDirFor(&options{Paths: []string{dir}})

	want, _ := filepath.Abs(dir)
	if got != want {
		t.Fatalf("startDirFor(directory path) = %q, want %q", got, want)
	}
}

func TestStartDirForFallsBackToWorkingDirectoryWithNoPathsOrStdinFilename(t *testing.T) {
	t.Parallel()

	got := startDirFor(&options{})

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	if got != wd {
		t.Fatalf("startDirFor(no paths) = %q, want the working directory %q", got, wd)
	}
}
