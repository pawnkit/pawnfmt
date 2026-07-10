package format_test

import (
	"bytes"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func TestFormatSourceWithCursorPreservesPositionInsideIdentifier(t *testing.T) {
	t.Parallel()

	source := []byte("new   playerScore=1;\n")
	cursor := bytes.Index(source, []byte("playerScore")) + 6

	result, err := formatter.SourceWithCursor(source, config.Default(), cursor)
	if err != nil {
		t.Fatalf("SourceWithCursor: %v", err)
	}

	want := bytes.Index(result.Source, []byte("playerScore")) + 6
	if result.CursorOffset != want {
		t.Fatalf("cursor offset = %d, want %d in %q", result.CursorOffset, want, result.Source)
	}
}

func TestFormatSourceWithCursorMapsEndOfFile(t *testing.T) {
	t.Parallel()

	source := []byte("new x=1;")

	result, err := formatter.SourceWithCursor(source, config.Default(), len(source))
	if err != nil {
		t.Fatalf("SourceWithCursor: %v", err)
	}

	if result.CursorOffset != len(result.Source) {
		t.Fatalf("end cursor = %d, want formatted length %d", result.CursorOffset, len(result.Source))
	}
}

func TestFormatSourceWithCursorRejectsInvalidOffset(t *testing.T) {
	t.Parallel()

	if _, err := formatter.SourceWithCursor([]byte("new x;\n"), config.Default(), 100); err == nil {
		t.Fatal("expected invalid cursor offset to fail")
	}
}

func TestFormatSourceWithCursorSuppressesIncludeSorting(t *testing.T) {
	t.Parallel()

	source := []byte("#include <zeta>\n#include <alpha>\nnew value=1;\n")
	cfg := config.Default()
	cfg.SortIncludes = true

	result, err := formatter.SourceWithCursor(source, cfg, bytes.Index(source, []byte("value")))
	if err != nil {
		t.Fatalf("SourceWithCursor: %v", err)
	}

	if bytes.Index(result.Source, []byte("zeta")) > bytes.Index(result.Source, []byte("alpha")) {
		t.Fatalf("cursor formatting unexpectedly reordered includes:\n%s", result.Source)
	}
}
