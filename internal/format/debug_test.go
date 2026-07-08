package format_test

import (
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func TestDebugDocTreeCoversEveryDocKindUsedByTheFormatter(t *testing.T) {
	source := []byte("stock F(a, b) {\n    if (a) {\n        return b;\n    }\n}\n")

	got, err := formatter.DebugDocTree(source, config.Default())
	if err != nil {
		t.Fatalf("DebugDocTree returned an error for valid source: %v", err)
	}

	for _, want := range []string{"Concat", "Text(", "HardLine", "Indent", "Group"} {
		if !strings.Contains(got, want) {
			t.Fatalf("DebugDocTree output missing %q marker:\n%s", want, got)
		}
	}

	if strings.Contains(got, "<unknown doc node>") {
		t.Fatalf("DebugDocTree hit the unknown-doc-node fallback for a construct the formatter itself produced:\n%s", got)
	}
}

func TestDebugDocTreeCoversResetIndentAndOutdent(t *testing.T) {
	source := []byte("stock F() {\n    if (x) {\n        #if A\n        new y;\n        #endif\n    }\n}\n")

	got, err := formatter.DebugDocTree(source, config.Default())
	if err != nil {
		t.Fatalf("DebugDocTree returned an error: %v", err)
	}

	if !strings.Contains(got, "Outdent") {
		t.Fatalf("DebugDocTree output missing Outdent marker:\n%s", got)
	}

	cfg := config.Default()
	cfg.DirectiveIndent = config.DirectiveIndentNone

	gotNone, err := formatter.DebugDocTree(source, cfg)
	if err != nil {
		t.Fatalf("DebugDocTree (DirectiveIndentNone) returned an error: %v", err)
	}

	if !strings.Contains(gotNone, "ResetIndent") {
		t.Fatalf("DebugDocTree (DirectiveIndentNone) output missing ResetIndent marker:\n%s", gotNone)
	}
}

func TestDebugDocTreeCoversFill(t *testing.T) {
	cfg := config.Default()
	cfg.MultilineCallArgs = config.MultilineListBinPack
	cfg.LineWidth = 20
	source := []byte("stock F() {\n    Call(aaaaaaaaaa, bbbbbbbbbb, cccccccccc);\n}\n")

	got, err := formatter.DebugDocTree(source, cfg)
	if err != nil {
		t.Fatalf("DebugDocTree returned an error: %v", err)
	}

	if !strings.Contains(got, "Fill") {
		t.Fatalf("DebugDocTree output missing Fill marker:\n%s", got)
	}
}

func TestDebugDocTreeRejectsUnparseableSource(t *testing.T) {
	_, err := formatter.DebugDocTree([]byte("}"), config.Default())
	if err == nil {
		t.Fatal("DebugDocTree should return an error for source that does not parse cleanly")
	}
}

func TestDebugDocTreeRejectsInvalidConfig(t *testing.T) {
	cfg := config.Default()
	cfg.DirectiveIndent = "not-a-real-value"

	_, err := formatter.DebugDocTree([]byte("new x;\n"), cfg)
	if err == nil {
		t.Fatal("DebugDocTree should return an error for an invalid config")
	}
}

func TestDebugCSTReportsNodeKindsPositionsAndErrorMarkers(t *testing.T) {
	got := formatter.DebugCST([]byte("new x = 1;\n"))
	for _, want := range []string{"source_file", "variable_declaration", "[0:"} {
		if !strings.Contains(got, want) {
			t.Fatalf("DebugCST output missing %q:\n%s", want, got)
		}
	}

	if strings.Contains(got, "[raw/error]") {
		t.Fatalf("DebugCST should not mark clean source as raw/error:\n%s", got)
	}
}

func TestDebugCSTMarksBrokenNodes(t *testing.T) {
	got := formatter.DebugCST([]byte("stock F( {{{ ) {"))
	if !strings.Contains(got, "[raw/error]") && !strings.Contains(got, "broken") {
		t.Fatalf("DebugCST should surface broken/error nodes for unparseable source:\n%s", got)
	}
}
