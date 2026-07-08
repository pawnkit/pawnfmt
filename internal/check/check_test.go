package check_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/check"
	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestParsesCleanlyAcceptsValidSource(t *testing.T) {
	ok, err := check.ParsesCleanly([]byte("stock F() { return 1; }\n"))
	if err != nil {
		t.Fatalf("ParsesCleanly returned an error for valid source: %v", err)
	}

	if !ok {
		t.Fatal("ParsesCleanly = false for valid source, want true")
	}
}

func TestParsesCleanlyRejectsGenuinelyBrokenSource(t *testing.T) {
	ok, err := check.ParsesCleanly([]byte("}"))
	if err != nil {
		t.Fatalf("ParsesCleanly returned an error: %v", err)
	}

	if ok {
		t.Fatal("ParsesCleanly = true for a stray closing brace, want false")
	}
}

func TestIdempotentReportsTrueWhenReformattingIsAFixedPoint(t *testing.T) {
	ok, err := check.Idempotent([]byte("stable"), func(b []byte) ([]byte, error) {
		return b, nil
	})
	if err != nil {
		t.Fatalf("Idempotent returned an error: %v", err)
	}

	if !ok {
		t.Fatal("Idempotent = false when the formatter returns its input unchanged, want true")
	}
}

func TestIdempotentReportsFalseWhenASecondPassChangesOutput(t *testing.T) {
	ok, err := check.Idempotent([]byte("first"), func(_ []byte) ([]byte, error) {
		return []byte("second"), nil
	})
	if err != nil {
		t.Fatalf("Idempotent returned an error: %v", err)
	}

	if ok {
		t.Fatal("Idempotent = true when a second pass changes the output, want false")
	}
}

func TestIdempotentPropagatesTheReformatError(t *testing.T) {
	wantErr := errors.New("boom")

	_, err := check.Idempotent([]byte("x"), func(_ []byte) ([]byte, error) {
		return nil, wantErr
	})
	if err == nil {
		t.Fatal("Idempotent should return an error when the reformat function fails")
	}

	if !errors.Is(err, wantErr) {
		t.Fatalf("Idempotent error = %v, want it to wrap %v", err, wantErr)
	}
}

func TestAnalyzeCorpusFileClassifiesACleanFileAsFull(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "clean.pwn")
	if err := os.WriteFile(path, []byte("stock F() {\n    return 1;\n}\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	r := check.AnalyzeCorpusFile(path, config.Default())
	if r.Status != check.CorpusStatusFull {
		t.Fatalf("Status = %q, want %q (detail: %s)", r.Status, check.CorpusStatusFull, r.Detail)
	}

	if !r.Idempotent {
		t.Fatalf("Idempotent = false for a clean file, want true")
	}

	if r.Path != path {
		t.Fatalf("Path = %q, want %q", r.Path, path)
	}
}

func TestAnalyzeCorpusFileFailsForAMissingFile(t *testing.T) {
	r := check.AnalyzeCorpusFile(filepath.Join(t.TempDir(), "does-not-exist.pwn"), config.Default())
	if r.Status != check.CorpusStatusFail {
		t.Fatalf("Status = %q, want %q for a missing file", r.Status, check.CorpusStatusFail)
	}

	if r.Detail == "" {
		t.Fatal("Detail should explain why a missing file failed")
	}
}

func TestAnalyzeCorpusFileFailsForBrokenSource(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "broken.pwn")
	if err := os.WriteFile(path, []byte("}"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	r := check.AnalyzeCorpusFile(path, config.Default())
	if r.Status != check.CorpusStatusFail {
		t.Fatalf("Status = %q, want %q for source the parser reports Broken", r.Status, check.CorpusStatusFail)
	}
}

func TestAnalyzeCorpusFileFailsWhenTheConfigItselfIsInvalid(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "valid.pwn")
	if err := os.WriteFile(path, []byte("new x;\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cfg := config.Default()
	cfg.DirectiveIndent = "not-a-real-value"

	r := check.AnalyzeCorpusFile(path, cfg)
	if r.Status != check.CorpusStatusFail {
		t.Fatalf("Status = %q, want %q when the config fails validation", r.Status, check.CorpusStatusFail)
	}

	if r.Detail == "" {
		t.Fatal("Detail should explain the format error")
	}
}

func TestCollectPawnFilesFindsOnlyPwnAndIncSortedByPath(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"b.pwn", "a.inc", "c.txt", "sub/d.pwn"} {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}

		if err := os.WriteFile(full, []byte("new x;\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	files, err := check.CollectPawnFiles(dir)
	if err != nil {
		t.Fatalf("CollectPawnFiles: %v", err)
	}

	want := []string{
		filepath.Join(dir, "a.inc"),
		filepath.Join(dir, "b.pwn"),
		filepath.Join(dir, "sub/d.pwn"),
	}
	if len(files) != len(want) {
		t.Fatalf("CollectPawnFiles = %v, want %v", files, want)
	}

	for i := range want {
		if files[i] != want[i] {
			t.Fatalf("CollectPawnFiles[%d] = %q, want %q", i, files[i], want[i])
		}
	}
}
