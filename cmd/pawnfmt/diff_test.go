package main

import (
	"strings"
	"testing"
)

func TestUnifiedDiffReturnsEmptyStringWhenNothingChanged(t *testing.T) {
	t.Parallel()

	if got := unifiedDiff("a.pwn", []byte("same\n"), []byte("same\n")); got != "" {
		t.Fatalf("unifiedDiff(unchanged) = %q, want empty", got)
	}
}

func TestUnifiedDiffIncludesFileHeaderAndHunkMarkers(t *testing.T) {
	t.Parallel()

	got := unifiedDiff("a.pwn", []byte("old\n"), []byte("new\n"))
	if !strings.Contains(got, "--- a.pwn\n") || !strings.Contains(got, "+++ a.pwn\n") {
		t.Fatalf("unifiedDiff should include ---/+++ headers naming the file:\n%s", got)
	}

	if !strings.Contains(got, "@@") {
		t.Fatalf("unifiedDiff should include an @@ hunk marker:\n%s", got)
	}

	if !strings.Contains(got, "-old") || !strings.Contains(got, "+new") {
		t.Fatalf("unifiedDiff should mark the removed and added lines:\n%s", got)
	}
}

func TestUnifiedDiffKeepsUnchangedContextLinesAroundAChange(t *testing.T) {
	t.Parallel()

	before := "line1\nline2\nold\nline4\nline5\n"
	after := "line1\nline2\nnew\nline4\nline5\n"

	got := unifiedDiff("a.pwn", []byte(before), []byte(after))
	for _, want := range []string{" line1", " line2", "-old", "+new", " line4", " line5"} {
		if !strings.Contains(got, want) {
			t.Fatalf("unifiedDiff missing %q in output:\n%s", want, got)
		}
	}
}

func TestUnifiedDiffNormalizesCRLFBeforeComparing(t *testing.T) {
	t.Parallel()
	// A pure CRLF vs LF difference in otherwise-identical content should
	// not be treated as a content change.
	if got := unifiedDiff("a.pwn", []byte("a\r\nb\r\n"), []byte("a\nb\n")); got != "" {
		t.Fatalf("unifiedDiff(CRLF vs LF, same content) = %q, want empty", got)
	}
}

func TestUnifiedDiffHandlesAppendedAndRemovedLinesAtTheEnd(t *testing.T) {
	t.Parallel()

	appended := unifiedDiff("a.pwn", []byte("a\nb\n"), []byte("a\nb\nc\n"))
	if !strings.Contains(appended, "+c") {
		t.Fatalf("unifiedDiff(append) should show +c:\n%s", appended)
	}

	removed := unifiedDiff("a.pwn", []byte("a\nb\nc\n"), []byte("a\nb\n"))
	if !strings.Contains(removed, "-c") {
		t.Fatalf("unifiedDiff(remove) should show -c:\n%s", removed)
	}
}

func TestSplitLinesDropsTrailingEmptyLineFromFinalNewline(t *testing.T) {
	t.Parallel()

	lines := splitLines("a\nb\n")

	want := []string{"a", "b"}
	if len(lines) != len(want) {
		t.Fatalf("splitLines(\"a\\nb\\n\") = %v, want %v", lines, want)
	}

	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("splitLines(\"a\\nb\\n\")[%d] = %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestSplitLinesEmptyStringReturnsNil(t *testing.T) {
	t.Parallel()

	if lines := splitLines(""); lines != nil {
		t.Fatalf("splitLines(\"\") = %v, want nil", lines)
	}
}
