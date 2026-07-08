package format

import (
	"strings"
	"testing"
)

func TestSharedContinuationIndentUsesGivenWidth(t *testing.T) {
	continuation, columns := sharedContinuationIndent(8, false)
	if continuation != strings.Repeat(" ", 8) || columns != 8 {
		t.Fatalf("sharedContinuationIndent(8, false) = (%q, %d), want (8 spaces, 8)", continuation, columns)
	}
}

func TestWrapSharedLineUsesContinuationWidthDistinctFromIndentWidth(t *testing.T) {
	line := "aaaaaaaaaa + bbbbbbbbbb + cccccccccc + dddddddddd + eeeeeeeeee;"

	out := wrapSharedLine(line, 30, 4, 8, false)
	if len(out) < 2 {
		t.Fatalf("wrapSharedLine did not wrap at all: %#v", out)
	}

	for i, seg := range out[1:] {
		if !strings.HasPrefix(seg, strings.Repeat(" ", 8)) {
			t.Fatalf("continuation line %d = %q, want it to start with 8 spaces (continuationWidth), not indentWidth's 4", i+1, seg)
		}
	}
}
