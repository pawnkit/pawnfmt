package format

import (
	"bytes"
	"testing"

	"github.com/pawnkit/pawn-parser"
)

func TestVerifySemanticTokensAllowsFormattingStructure(t *testing.T) {
	t.Parallel()

	before := []byte("if (ready) return Call(a, b);\n")

	after := []byte("if (ready) {\n    return Call(a, b,);\n}\n")
	if err := verifySemanticTokens(before, after); err != nil {
		t.Fatalf("structural formatting should be allowed: %v", err)
	}
}

func TestVerifySemanticStructureAllowsControlBodyBraces(t *testing.T) {
	t.Parallel()

	before := parser.Parse([]byte("stock F() { if (ready) return 1; else return 2; }\n"))

	after := parser.Parse([]byte("stock F() { if (ready) { return 1; } else { return 2; } }\n"))
	if err := verifySemanticStructure(before, after); err != nil {
		t.Fatalf("control-body braces should be structurally equivalent: %v", err)
	}
}

func TestVerifySemanticStructureRejectsExpandedControlScope(t *testing.T) {
	t.Parallel()

	beforeSource := []byte("stock F() { if (ready) Call(); return 1; }\n")
	afterSource := []byte("stock F() { if (ready) { Call(); return 1; } }\n")

	if err := verifySemanticTokens(beforeSource, afterSource); err != nil {
		t.Fatalf("fixture must demonstrate a change the token-only check permits: %v", err)
	}

	if err := verifySemanticStructure(parser.Parse(beforeSource), parser.Parse(afterSource)); err == nil {
		t.Fatal("expected expanded control scope to be rejected")
	}
}

func TestVerifySemanticStructureRejectsMovedElseBranch(t *testing.T) {
	t.Parallel()

	beforeSource := []byte("stock F() { if (a) if (b) Call(); else Other(); }\n")
	afterSource := []byte("stock F() { if (a) { if (b) Call(); } else Other(); }\n")

	if err := verifySemanticTokens(beforeSource, afterSource); err != nil {
		t.Fatalf("fixture must demonstrate a change the token-only check permits: %v", err)
	}

	if err := verifySemanticStructure(parser.Parse(beforeSource), parser.Parse(afterSource)); err == nil {
		t.Fatal("expected an else branch moving between if statements to be rejected")
	}
}

func TestVerifySemanticTokensRejectsMeaningfulChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		after string
	}{
		{name: "identifier", after: "value = replacement + 1;"},
		{name: "operator", after: "value = source - 1;"},
		{name: "literal", after: "value = source + 2;"},
	}
	before := []byte("value = source + 1;")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := verifySemanticTokens(before, []byte(test.after)); err == nil {
				t.Fatal("expected semantic change to be rejected")
			}
		})
	}
}

func TestVerifySemanticTokensWithSortedIncludesAllowsOnlyIncludeReordering(t *testing.T) {
	t.Parallel()

	beforeSource := []byte("#include <zeta>\n#tryinclude <alpha>\nnew value = 1;\n")

	afterSource := []byte("#tryinclude <alpha>\n#include <zeta>\nnew value = 1;\n")
	if err := verifySemanticTokensWithSortedIncludes(beforeSource, afterSource,
		parser.Parse(beforeSource), parser.Parse(afterSource)); err != nil {
		t.Fatalf("include-only reordering should be allowed: %v", err)
	}

	changedSource := []byte("#tryinclude <alpha>\n#include <zeta>\nnew replacement = 1;\n")
	if err := verifySemanticTokensWithSortedIncludes(beforeSource, changedSource,
		parser.Parse(beforeSource), parser.Parse(changedSource)); err == nil {
		t.Fatal("a non-include semantic change must be rejected")
	}
}

func TestVerifySemanticTokensFromFilesMatchesByteBasedResult(t *testing.T) {
	t.Parallel()

	before := []byte("if (ready) return Call(a, b);\n")
	after := []byte("if (ready) {\n    return Call(a, b,);\n}\n")

	byteErr := verifySemanticTokens(before, after)
	fileErr := verifySemanticTokensFromFiles(parser.Parse(before), parser.Parse(after), before, after)

	if (byteErr == nil) != (fileErr == nil) {
		t.Fatalf("byte-based and file-based checks disagree: byte=%v file=%v", byteErr, fileErr)
	}

	changed := []byte("if (ready) return Call(a, c);\n")

	byteErr = verifySemanticTokens(before, changed)
	fileErr = verifySemanticTokensFromFiles(parser.Parse(before), parser.Parse(changed), before, changed)

	if (byteErr == nil) != (fileErr == nil) {
		t.Fatalf("byte-based and file-based checks disagree on a real change: byte=%v file=%v", byteErr, fileErr)
	}
}

func TestSemanticTokensOutsideIncludesUsesFileTokens(t *testing.T) {
	t.Parallel()

	source := []byte("#include <zeta>\nnew value = 1;\n")
	file := parser.Parse(source)

	fromFile := semanticTokensOutsideIncludes(source, file)
	fromScratch := semanticTokensOutsideIncludes(append([]byte(nil), source...), parser.Parse(append([]byte(nil), source...)))

	if len(fromFile) != len(fromScratch) {
		t.Fatalf("token count diverged: %d vs %d", len(fromFile), len(fromScratch))
	}

	for i := range fromFile {
		if fromFile[i].kind != fromScratch[i].kind || !bytes.Equal(fromFile[i].text, fromScratch[i].text) {
			t.Fatalf("token %d diverged: %+v vs %+v", i, fromFile[i], fromScratch[i])
		}
	}
}
