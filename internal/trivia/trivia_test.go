package trivia_test

import (
	"bytes"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/trivia"
)

func TestScanEmptySourceProducesOneEmptyLine(t *testing.T) {
	index := trivia.Scan([]byte(""))
	if len(index.Lines) != 1 {
		t.Fatalf("Scan(\"\") produced %d lines, want 1", len(index.Lines))
	}
	line := index.Lines[0]
	if line.Number != 1 || line.Text != "" || !line.IsBlank {
		t.Fatalf("Scan(\"\") line = %#v, want an empty, blank line 1", line)
	}
}

func TestScanSplitsLinesAndTracksByteOffsetsAndNumbers(t *testing.T) {
	source := []byte("first\nsecond\nthird")
	index := trivia.Scan(source)
	if len(index.Lines) != 3 {
		t.Fatalf("Scan(...) produced %d lines, want 3", len(index.Lines))
	}
	want := []struct {
		number             int
		text               string
		startByte, endByte int
	}{
		{1, "first", 0, 5},
		{2, "second", 6, 12},
		{3, "third", 13, 18},
	}
	for i, w := range want {
		l := index.Lines[i]
		if l.Number != w.number || l.Text != w.text || l.StartByte != w.startByte || l.EndByte != w.endByte {
			t.Fatalf("line %d = %#v, want {Number:%d Text:%q StartByte:%d EndByte:%d}", i, l, w.number, w.text, w.startByte, w.endByte)
		}
	}
}

func TestScanHandlesCRLFLineEndings(t *testing.T) {
	source := []byte("first\r\nsecond")
	index := trivia.Scan(source)
	if index.DetectedNewline != "\r\n" {
		t.Fatalf("DetectedNewline = %q, want \\r\\n", index.DetectedNewline)
	}
	if len(index.Lines) != 2 || index.Lines[0].Text != "first" || index.Lines[0].EndByte != 5 {
		t.Fatalf("CRLF line split = %#v, want first line %q ending at byte 5", index.Lines, "first")
	}
}

func TestScanDetectsPlainLFWhenNoCRLFPresent(t *testing.T) {
	index := trivia.Scan([]byte("a\nb\n"))
	if index.DetectedNewline != "\n" {
		t.Fatalf("DetectedNewline = %q, want \\n", index.DetectedNewline)
	}
}

func TestScanTrailingNewlineDoesNotProduceADanglingEmptyLine(t *testing.T) {
	index := trivia.Scan([]byte("a\n"))
	if len(index.Lines) != 2 {
		t.Fatalf("Scan(\"a\\n\") produced %d lines, want 2 (\"a\" plus the trailing empty line)", len(index.Lines))
	}
	if index.Lines[0].Text != "a" || index.Lines[1].Text != "" {
		t.Fatalf("Scan(\"a\\n\") lines = %#v, want [\"a\", \"\"]", index.Lines)
	}
}

func TestLineIsBlankForWhitespaceOnlyLines(t *testing.T) {
	index := trivia.Scan([]byte("   \t  \nreal"))
	if !index.Lines[0].IsBlank {
		t.Fatalf("whitespace-only line IsBlank = false, want true: %#v", index.Lines[0])
	}
	if index.Lines[1].IsBlank {
		t.Fatalf("non-blank line IsBlank = true, want false: %#v", index.Lines[1])
	}
}

func TestLineIsDirectiveDetectsLeadingHashIgnoringIndentation(t *testing.T) {
	index := trivia.Scan([]byte("#if A\n    #endif\nnew x;"))
	if !index.Lines[0].IsDirective {
		t.Fatalf("\"#if A\" IsDirective = false, want true")
	}
	if !index.Lines[1].IsDirective {
		t.Fatalf("indented \"#endif\" IsDirective = false, want true")
	}
	if index.Lines[2].IsDirective {
		t.Fatalf("\"new x;\" IsDirective = true, want false")
	}
}

func TestLineIsCommentOnlyForLineBlockAndContinuationComments(t *testing.T) {
	source := []byte("// line comment\n/* block start\n * continuation\n */\nnew x;")
	index := trivia.Scan(source)
	for i, want := range []bool{true, true, true, true, false} {
		if index.Lines[i].IsCommentOnly != want {
			t.Fatalf("line %d (%q) IsCommentOnly = %v, want %v", i, index.Lines[i].Text, index.Lines[i].IsCommentOnly, want)
		}
	}
}

func TestTurnsOffAndOnRequireCommentOnlyLines(t *testing.T) {
	cases := []struct {
		name        string
		text        string
		wantsOff    bool
		wantsOn     bool
		commentOnly bool
	}{
		{"off comment", "// pawnfmt off", true, false, true},
		{"on comment", "// pawnfmt on", false, true, true},
		{"case insensitive", "// PawnFmt Off", true, false, true},
		{"not a comment", "pawnfmt off", false, false, false},
		{"unrelated comment", "// just a comment", false, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			index := trivia.Scan([]byte(tc.text))
			line := index.Lines[0]
			if line.IsCommentOnly != tc.commentOnly {
				t.Fatalf("IsCommentOnly = %v, want %v", line.IsCommentOnly, tc.commentOnly)
			}
			if line.TurnsOff != tc.wantsOff {
				t.Fatalf("TurnsOff = %v, want %v", line.TurnsOff, tc.wantsOff)
			}
			if line.TurnsOn != tc.wantsOn {
				t.Fatalf("TurnsOn = %v, want %v", line.TurnsOn, tc.wantsOn)
			}
		})
	}
}

func TestDisabledRegionCoversOffToOnMarkerInclusive(t *testing.T) {
	source := []byte("new a;\n// pawnfmt off\nnew   b;\n// pawnfmt on\nnew c;\n")
	index := trivia.Scan(source)
	if len(index.Disabled) != 1 {
		t.Fatalf("Disabled regions = %#v, want exactly 1", index.Disabled)
	}
	region := index.Disabled[0]
	if region.StartLine != 2 || region.EndLine != 4 {
		t.Fatalf("region lines = [%d,%d], want [2,4] (the off/on marker lines themselves)", region.StartLine, region.EndLine)
	}
	middle := []byte("new   b;")
	if !index.OverlapsDisabled(uint32(bytes.Index(source, middle)), uint32(bytes.Index(source, middle)+len(middle))) {
		t.Fatalf("region does not cover the disabled statement between the markers:\n%s", source)
	}
	tail := []byte("new c;")
	if index.OverlapsDisabled(uint32(bytes.Index(source, tail)), uint32(bytes.Index(source, tail)+len(tail))) {
		t.Fatalf("region incorrectly covers a statement after the \"on\" marker:\n%s", source)
	}
}

func TestDisabledRegionWithNoMatchingOnExtendsToEndOfFile(t *testing.T) {
	source := []byte("new a;\n// pawnfmt off\nnew b;\n")
	index := trivia.Scan(source)
	if len(index.Disabled) != 1 {
		t.Fatalf("Disabled regions = %#v, want exactly 1", index.Disabled)
	}
	region := index.Disabled[0]
	if region.EndByte != len(source) {
		t.Fatalf("unterminated disabled region EndByte = %d, want %d (end of source)", region.EndByte, len(source))
	}
}

func TestRepeatedOffMarkersDoNotStackOrResetTheRegion(t *testing.T) {
	source := []byte("// pawnfmt off\nnew a;\n// pawnfmt off\nnew b;\n// pawnfmt on\nnew c;\n")
	index := trivia.Scan(source)
	if len(index.Disabled) != 1 {
		t.Fatalf("Disabled regions = %#v, want exactly 1 (repeated off markers should not create extra regions)", index.Disabled)
	}
	if index.Disabled[0].StartLine != 1 {
		t.Fatalf("region StartLine = %d, want 1 (the first off marker, not the second)", index.Disabled[0].StartLine)
	}
}

func TestMultipleDisabledRegionsAreIndependent(t *testing.T) {
	source := []byte(
		"new a;\n// pawnfmt off\nnew b;\n// pawnfmt on\nnew c;\n// pawnfmt off\nnew d;\n// pawnfmt on\nnew e;\n",
	)
	index := trivia.Scan(source)
	if len(index.Disabled) != 2 {
		t.Fatalf("Disabled regions = %#v, want exactly 2", index.Disabled)
	}
}

func TestOverlapsDisabledIsFalseOutsideAnyRegion(t *testing.T) {
	index := trivia.Scan([]byte("new a;\n"))
	if index.OverlapsDisabled(0, 6) {
		t.Fatal("OverlapsDisabled should be false when there are no disabled regions at all")
	}
}

func TestOverlapsDisabledIsFalseStrictlyBeforeARegion(t *testing.T) {
	source := []byte("new a;\n// pawnfmt off\nnew b;\n// pawnfmt on\nnew c;\n")
	index := trivia.Scan(source)
	before := []byte("new a;")
	start := bytes.Index(source, before)
	if index.OverlapsDisabled(uint32(start), uint32(start+len(before))) {
		t.Fatalf("OverlapsDisabled should be false for a range entirely before the disabled region:\n%s", source)
	}
}
