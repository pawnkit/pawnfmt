package format_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

const (
	elseDirective         = "#else"
	elseDirectiveIndented = "    #else"
	emptyBraceBody        = "{ }"
	closingBraceIndented  = "    }"
)

func mustFormat(t *testing.T, source []byte, cfg config.Config) []byte {
	t.Helper()
	formatted, err := formatter.FormatSource(source, cfg)
	if err != nil {
		t.Fatalf("format source: %v", err)
	}
	return formatted
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func ensureTrailingNewline(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}
	return append(data, '\n')
}

func testdataDir() string {
	return filepath.Join("..", "..", "testdata")
}

func requireSharedConditionalPath(t *testing.T, source []byte) {
	t.Helper()
	parsed := parser.Parse(source)
	if parsed.Root == nil {
		t.Fatalf("source did not parse at all; cannot verify it exercises the shared-conditional fallback path")
	}
	if !containsKind(parsed.Root, parser.KindSharedConditional, parser.KindSharedConditionalPrefix, parser.KindConditionalSplice) {
		t.Fatalf("fixture no longer parses to a shared-conditional/raw-fallback node -- it may have drifted onto " +
			"the structured rendering path and no longer tests what this test's name claims; run -debug-cst on the " +
			"source to see what it parses to now")
	}
}

func containsKind(n *parser.Node, kinds ...parser.Kind) bool {
	if n == nil {
		return false
	}
	if slices.Contains(kinds, n.Kind) {
		return true
	}
	for _, c := range n.Children {
		if containsKind(c, kinds...) {
			return true
		}
	}
	return false
}
