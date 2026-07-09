package config_test

import (
	"os"
	"path/filepath"
	"testing"

	editorconfig "github.com/editorconfig/editorconfig-core-go/v2"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestApplyEditorConfigMapsSupportedProperties(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	contents := "root = true\n\n[*.pwn]\nindent_style = space\nindent_size = 2\nend_of_line = crlf\ninsert_final_newline = false\ntrim_trailing_whitespace = false\nmax_line_length = 120\n"
	if err := os.WriteFile(filepath.Join(dir, ".editorconfig"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(dir, "script.pwn")

	cfg := config.Default()
	if err := config.ApplyEditorConfig(filename, &cfg, editorconfig.NewCachedParser()); err != nil {
		t.Fatalf("ApplyEditorConfig: %v", err)
	}

	if cfg.IndentStyle != config.IndentStyleSpace || cfg.IndentWidth != 2 {
		t.Fatalf("indent = %q/%d, want space/2", cfg.IndentStyle, cfg.IndentWidth)
	}
	if cfg.NewlineStyle != config.NewlineStyleCRLF || cfg.LineWidth != 120 {
		t.Fatalf("newline/width = %q/%d, want crlf/120", cfg.NewlineStyle, cfg.LineWidth)
	}
	if cfg.InsertFinalNewline || cfg.TrimTrailingWhitespace {
		t.Fatal("explicit false EditorConfig properties were not applied")
	}
}

func TestPawnConfigLoadedOverEditorConfigWins(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".editorconfig"), []byte("root = true\n[*]\nindent_size = 2\nmax_line_length = 120\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pawnPath := filepath.Join(dir, "pawnfmt.toml")
	if err := os.WriteFile(pawnPath, []byte("indent_width = 6\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	base := config.Default()
	if err := config.ApplyEditorConfig(filepath.Join(dir, "script.pwn"), &base, editorconfig.NewCachedParser()); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadFileWithBase(pawnPath, base)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IndentWidth != 6 || cfg.LineWidth != 120 {
		t.Fatalf("precedence produced indent_width=%d line_width=%d, want 6 and 120", cfg.IndentWidth, cfg.LineWidth)
	}
}

func TestApplyEditorConfigRejectsInvalidMappedValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".editorconfig"), []byte("root = true\n[*]\nmax_line_length = 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	if err := config.ApplyEditorConfig(filepath.Join(dir, "script.pwn"), &cfg, editorconfig.NewCachedParser()); err == nil {
		t.Fatal("expected invalid max_line_length to be rejected")
	}
}
