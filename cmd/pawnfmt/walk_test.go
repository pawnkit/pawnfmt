package main

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
)

const keepFixtureName = "keep.pwn"

func writeFixture(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}

	if err := os.WriteFile(path, []byte("new x;\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCollectFilesExplicitSingleFileIsUsedRegardlessOfExtension(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "explicit.txt")
	writeFixture(t, path)

	files, err := collectFiles([]string{path}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if len(files) != 1 || files[0] != path {
		t.Fatalf("collectFiles(explicit file) = %v, want [%s] (an explicitly named file bypasses the extension filter)", files, path)
	}
}

func TestCollectFilesWalksADirectoryForPwnAndIncOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "a.pwn"))
	writeFixture(t, filepath.Join(dir, "b.inc"))
	writeFixture(t, filepath.Join(dir, "c.txt"))
	writeFixture(t, filepath.Join(dir, "sub", "d.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	want := []string{
		filepath.Join(dir, "a.pwn"),
		filepath.Join(dir, "b.inc"),
		filepath.Join(dir, "sub", "d.pwn"),
	}
	for _, w := range want {
		if !slices.Contains(files, w) {
			t.Fatalf("collectFiles(dir) = %v, missing %s", files, w)
		}
	}

	if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "c.txt" }) {
		t.Fatalf("collectFiles(dir) = %v, should not include c.txt", files)
	}
}

func TestCollectFilesSkipsDefaultIgnoredDirectories(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, keepFixtureName))
	writeFixture(t, filepath.Join(dir, "node_modules", "skip.pwn"))
	writeFixture(t, filepath.Join(dir, ".git", "skip.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	for _, f := range files {
		if filepath.Base(f) == "skip.pwn" {
			t.Fatalf("collectFiles(dir) = %v, should skip files under ignored directories", files)
		}
	}
}

func TestCollectFilesExcludePatternFiltersMatchingFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, keepFixtureName))
	writeFixture(t, filepath.Join(dir, "generated.pwn"))

	files, err := collectFiles([]string{dir}, nil, []string{"generated.pwn"}, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "generated.pwn" }) {
		t.Fatalf("collectFiles(dir, exclude=generated.pwn) = %v, should exclude it", files)
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == keepFixtureName }) {
		t.Fatalf("collectFiles(dir, exclude=generated.pwn) = %v, should still include keep.pwn", files)
	}
}

func TestCollectFilesIncludePatternRestrictsToMatchingFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, keepFixtureName))
	writeFixture(t, filepath.Join(dir, "other.pwn"))

	files, err := collectFiles([]string{dir}, []string{keepFixtureName}, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if len(files) != 1 || filepath.Base(files[0]) != keepFixtureName {
		t.Fatalf("collectFiles(dir, include=keep.pwn) = %v, want only keep.pwn", files)
	}
}

func TestCollectFilesIncludePatternCanRescueAnIgnoredDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, filepath.Join(dir, "vendor", "wanted.pwn"))

	files, err := collectFiles([]string{dir}, []string{"vendor", "wanted.pwn"}, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "wanted.pwn" }) {
		t.Fatalf("collectFiles(dir, include=[vendor,wanted.pwn]) = %v, want wanted.pwn rescued", files)
	}
}

func TestCollectFilesDeduplicatesTheSameFileNamedTwice(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dup.pwn")
	writeFixture(t, path)

	files, err := collectFiles([]string{path, path}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("collectFiles(same path twice) = %v, want exactly 1 entry", files)
	}
}

func TestCollectFilesReturnsErrorForANonexistentPath(t *testing.T) {
	t.Parallel()

	_, err := collectFiles([]string{filepath.Join(t.TempDir(), "missing")}, nil, nil, true)
	if err == nil {
		t.Fatal("collectFiles should return an error for a path that doesn't exist")
	}
}

func writeText(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCollectFilesRespectsGitignore(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeText(t, filepath.Join(dir, ".gitignore"), "*.gen.pwn\nbuild/\n")
	writeFixture(t, filepath.Join(dir, keepFixtureName))
	writeFixture(t, filepath.Join(dir, "skip.gen.pwn"))
	writeFixture(t, filepath.Join(dir, "build", "artifact.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == keepFixtureName }) {
		t.Fatalf("collectFiles(dir) = %v, want keep.pwn included", files)
	}

	if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "skip.gen.pwn" }) {
		t.Fatalf("collectFiles(dir) = %v, want *.gen.pwn excluded by .gitignore", files)
	}

	if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "artifact.pwn" }) {
		t.Fatalf("collectFiles(dir) = %v, want build/ excluded by .gitignore", files)
	}
}

func TestCollectFilesNoGitignoreOptOutFormatsEverything(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeText(t, filepath.Join(dir, ".gitignore"), "*.gen.pwn\n")
	writeFixture(t, filepath.Join(dir, "skip.gen.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, false)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "skip.gen.pwn" }) {
		t.Fatalf("collectFiles(dir, respectIgnoreFiles=false) = %v, want *.gen.pwn included since gitignore is disabled", files)
	}
}

func TestCollectFilesPawnfmtignoreLayersOnTopOfGitignore(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeText(t, filepath.Join(dir, ".gitignore"), "*.gen.pwn\n")
	writeText(t, filepath.Join(dir, ".pawnfmtignore"), "extra.pwn\n")
	writeFixture(t, filepath.Join(dir, keepFixtureName))
	writeFixture(t, filepath.Join(dir, "skip.gen.pwn"))
	writeFixture(t, filepath.Join(dir, "extra.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	for _, skip := range []string{"skip.gen.pwn", "extra.pwn"} {
		if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == skip }) {
			t.Fatalf("collectFiles(dir) = %v, want %s excluded", files, skip)
		}
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == keepFixtureName }) {
		t.Fatalf("collectFiles(dir) = %v, want keep.pwn included", files)
	}
}

func TestCollectFilesGitignoreNegationRescuesAFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeText(t, filepath.Join(dir, ".gitignore"), "*.pwn\n!important.pwn\n")
	writeFixture(t, filepath.Join(dir, "other.pwn"))
	writeFixture(t, filepath.Join(dir, "important.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "other.pwn" }) {
		t.Fatalf("collectFiles(dir) = %v, want other.pwn excluded", files)
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "important.pwn" }) {
		t.Fatalf("collectFiles(dir) = %v, want important.pwn rescued by negation", files)
	}
}

func TestCollectFilesGitignoreAnchoredPatternOnlyMatchesAtItsOwnLevel(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeText(t, filepath.Join(dir, ".gitignore"), "/root_only.pwn\n")
	writeFixture(t, filepath.Join(dir, "root_only.pwn"))
	writeFixture(t, filepath.Join(dir, "sub", "root_only.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if slices.ContainsFunc(files, func(f string) bool { return f == filepath.Join(dir, "root_only.pwn") }) {
		t.Fatalf("collectFiles(dir) = %v, want top-level root_only.pwn excluded", files)
	}

	if !slices.ContainsFunc(files, func(f string) bool { return f == filepath.Join(dir, "sub", "root_only.pwn") }) {
		t.Fatalf("collectFiles(dir) = %v, want sub/root_only.pwn NOT excluded (anchored pattern shouldn't match nested)", files)
	}
}

func TestCollectFilesGitignoreDoubleStarMatchesAnyDepth(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeText(t, filepath.Join(dir, ".gitignore"), "**/generated/*.pwn\n")
	writeFixture(t, filepath.Join(dir, "a", "generated", "x.pwn"))
	writeFixture(t, filepath.Join(dir, "a", "b", "generated", "y.pwn"))
	writeFixture(t, filepath.Join(dir, "a", keepFixtureName))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	for _, skip := range []string{"x.pwn", "y.pwn"} {
		if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == skip }) {
			t.Fatalf("collectFiles(dir) = %v, want %s excluded by **/generated/*.pwn", files, skip)
		}
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == keepFixtureName }) {
		t.Fatalf("collectFiles(dir) = %v, want keep.pwn included", files)
	}
}

func TestCollectFilesNestedGitignoreOverridesParent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeText(t, filepath.Join(dir, ".gitignore"), "*.pwn\n")
	writeText(t, filepath.Join(dir, "sub", ".gitignore"), "!rescued.pwn\n")
	writeFixture(t, filepath.Join(dir, "sub", "rescued.pwn"))
	writeFixture(t, filepath.Join(dir, "sub", "other.pwn"))

	files, err := collectFiles([]string{dir}, nil, nil, true)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "rescued.pwn" }) {
		t.Fatalf("collectFiles(dir) = %v, want rescued.pwn rescued by nested .gitignore", files)
	}

	if slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "other.pwn" }) {
		t.Fatalf("collectFiles(dir) = %v, want other.pwn still excluded by parent .gitignore", files)
	}
}

func TestGitignoreGlobToRegexp(t *testing.T) {
	t.Parallel()

	cases := []struct {
		glob  string
		match string
		want  bool
	}{
		{"*.log", "foo.log", true},
		{"*.log", "foo.log.txt", false},
		{"foo?bar", "fooxbar", true},
		{"foo?bar", "foobar", false},
		{"[abc].pwn", testFileA, true},
		{"[abc].pwn", "d.pwn", false},
		{"[!abc].pwn", "d.pwn", true},
	}
	for _, c := range cases {
		re := regexp.MustCompile(gitignoreGlobToRegexp(c.glob))
		if got := re.MatchString(c.match); got != c.want {
			t.Errorf("gitignoreGlobToRegexp(%q) matching %q = %v, want %v (regexp %q)", c.glob, c.match, got, c.want, re.String())
		}
	}
}

func TestIsFormattableExt(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"a.pwn": true, "a.inc": true, "a.PWN": true, "a.txt": false, "a": false,
	}
	for name, want := range cases {
		if got := isFormattableExt(name); got != want {
			t.Errorf("isFormattableExt(%q) = %v, want %v", name, got, want)
		}
	}
}
