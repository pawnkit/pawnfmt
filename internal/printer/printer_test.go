package printer_test

import (
	"testing"

	"github.com/pawnkit/pawnfmt/internal/doc"
	"github.com/pawnkit/pawnfmt/internal/printer"
)

func opts(lineWidth int) printer.Options {
	return printer.Options{LineWidth: lineWidth, IndentWidth: 4, Newline: "\n", InsertFinalNewline: false}
}

func TestPrintNilRootReturnsEmpty(t *testing.T) {
	if got := printer.Print(nil, opts(80)); got != "" {
		t.Fatalf("Print(nil) = %q, want empty string", got)
	}
}

func TestPrintText(t *testing.T) {
	if got := printer.Print(doc.Text("hello"), opts(80)); got != "hello" {
		t.Fatalf("Print(Text) = %q, want %q", got, "hello")
	}
}

func TestPrintConcatPreservesOrder(t *testing.T) {
	d := doc.Concat(doc.Text("a"), doc.Text("b"), doc.Text("c"))
	if got := printer.Print(d, opts(80)); got != "abc" {
		t.Fatalf("Print(Concat) = %q, want %q", got, "abc")
	}
}

func TestGroupRendersFlatWhenItFits(t *testing.T) {
	d := doc.Group(doc.Concat(doc.Text("a"), doc.Line(), doc.Text("b")))
	if got := printer.Print(d, opts(80)); got != "a b" {
		t.Fatalf("Print(short Group with Line) = %q, want %q (flat: Line -> space)", got, "a b")
	}
}

func TestGroupBreaksWhenItDoesNotFit(t *testing.T) {
	d := doc.Group(doc.Concat(doc.Text("aaaaaaaaaa"), doc.Line(), doc.Text("bbbbbbbbbb")))
	want := "aaaaaaaaaa\nbbbbbbbbbb"
	if got := printer.Print(d, opts(5)); got != want {
		t.Fatalf("Print(long Group with Line) = %q, want %q (broken: Line -> newline)", got, want)
	}
}

func TestGroupSoftLineVanishesWhenFlatButBreaksWhenBroken(t *testing.T) {
	flatDoc := doc.Group(doc.Concat(doc.Text("a"), doc.SoftLine(), doc.Text("b")))
	if got := printer.Print(flatDoc, opts(80)); got != "ab" {
		t.Fatalf("Print(short Group with SoftLine) = %q, want %q (flat: SoftLine -> nothing)", got, "ab")
	}
	brokenDoc := doc.Group(doc.Concat(doc.Text("aaaaaaaaaa"), doc.SoftLine(), doc.Text("bbbbbbbbbb")))
	want := "aaaaaaaaaa\nbbbbbbbbbb"
	if got := printer.Print(brokenDoc, opts(5)); got != want {
		t.Fatalf("Print(long Group with SoftLine) = %q, want %q (broken: SoftLine -> newline)", got, want)
	}
}

func TestHardLineAlwaysBreaksEvenInsideFlatGroup(t *testing.T) {
	d := doc.Group(doc.Concat(doc.Text("a"), doc.HardLine(), doc.Text("b")))
	want := "a\nb"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(Group containing HardLine) = %q, want %q (HardLine always breaks)", got, want)
	}
}

func TestIndentAddsOneLevelForHardLines(t *testing.T) {
	d := doc.Concat(doc.Text("a"), doc.Indent(doc.Concat(doc.HardLine(), doc.Text("b"))))
	want := "a\n    b"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(Indent) = %q, want %q", got, want)
	}
}

func TestIndentNestsAdditively(t *testing.T) {
	d := doc.Indent(doc.Indent(doc.Concat(doc.HardLine(), doc.Text("b"))))
	want := "\n        b"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(Indent(Indent(...))) = %q, want %q (two levels = 8 spaces)", got, want)
	}
}

func TestResetIndentForcesAbsoluteZeroRegardlessOfAmbient(t *testing.T) {
	d := doc.Indent(doc.Indent(doc.Concat(doc.HardLine(), doc.ResetIndent(doc.Concat(doc.HardLine(), doc.Text("b"))))))
	want := "\n        \nb"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(nested Indent then ResetIndent) = %q, want %q", got, want)
	}
}

func TestOutdentIsOneLevelLessThanAmbient(t *testing.T) {
	d := doc.Indent(doc.Indent(doc.Concat(doc.HardLine(), doc.Outdent(doc.Concat(doc.HardLine(), doc.Text("b"))))))
	want := "\n        \n    b"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(Indent(Indent(Outdent(...)))) = %q, want %q (one level less than the ambient two)", got, want)
	}
}

func TestOutdentFloorsAtZeroRatherThanGoingNegative(t *testing.T) {
	d := doc.Outdent(doc.Concat(doc.HardLine(), doc.Text("b")))
	want := "\nb"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(Outdent at ambient 0) = %q, want %q (floored at 0, not -1)", got, want)
	}
}

func TestOutdentDoesNotCompoundAcrossSiblings(t *testing.T) {
	d := doc.Indent(doc.Indent(doc.Concat(
		doc.Outdent(doc.Concat(doc.HardLine(), doc.Text("first"))),
		doc.Outdent(doc.Concat(doc.HardLine(), doc.Text("second"))),
	)))
	want := "\n    first\n    second"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(two sibling Outdents) = %q, want %q (both at the same one-level-less column)", got, want)
	}
}

func TestIfBreakSelectsFlatWhenGroupFits(t *testing.T) {
	d := doc.Group(doc.Concat(doc.Text("a"), doc.IfBreak(doc.Text(","), doc.Text(""))))
	if got := printer.Print(d, opts(80)); got != "a" {
		t.Fatalf("Print(fitting Group with IfBreak) = %q, want %q (flat branch)", got, "a")
	}
}

func TestIfBreakSelectsBrokenWhenGroupBreaks(t *testing.T) {
	d := doc.Group(doc.Concat(doc.Text("aaaaaaaaaa"), doc.Line(), doc.IfBreak(doc.Text(","), doc.Text(""))))
	want := "aaaaaaaaaa\n,"
	if got := printer.Print(d, opts(3)); got != want {
		t.Fatalf("Print(breaking Group with IfBreak) = %q, want %q (broken branch)", got, want)
	}
}

func TestIfBreakNilBranchRendersNothing(t *testing.T) {
	flat := doc.Group(doc.IfBreak(doc.Text("only-broken"), nil))
	if got := printer.Print(flat, opts(80)); got != "" {
		t.Fatalf("Print(flat Group, IfBreak with nil flat branch) = %q, want empty", got)
	}
}

func TestFillKeepsShortItemsOnOneLineButBreaksWhereNeeded(t *testing.T) {
	d := doc.Fill(doc.Text("a"), doc.Line(), doc.Text("b"), doc.Line(), doc.Text("c"))
	if got := printer.Print(d, opts(80)); got != "a b c" {
		t.Fatalf("Print(Fill, all fits) = %q, want %q", got, "a b c")
	}
}

func TestFillBreaksOnlyThePairThatOverflows(t *testing.T) {
	d := doc.Fill(doc.Text("aa"), doc.Line(), doc.Text("bb"), doc.Line(), doc.Text("cccccccccc"))
	want := "aa bb\ncccccccccc"
	if got := printer.Print(d, opts(6)); got != want {
		t.Fatalf("Print(Fill, last item overflows) = %q, want %q", got, want)
	}
}

func TestFillSingleItem(t *testing.T) {
	d := doc.Fill(doc.Text("solo"))
	if got := printer.Print(d, opts(80)); got != "solo" {
		t.Fatalf("Print(Fill with one item) = %q, want %q", got, "solo")
	}
}

func TestFillContentWithEmbeddedHardLineForcesBreakForThatChunk(t *testing.T) {
	multiline := doc.Concat(doc.Text("a"), doc.HardLine(), doc.Text("b"))
	d := doc.Fill(multiline, doc.Line(), doc.Text("c"))
	want := "a\nb\nc"
	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(Fill with an embedded-HardLine content chunk) = %q, want %q", got, want)
	}
}

func TestWithDefaultsAppliesFallbacksForZeroValues(t *testing.T) {
	d := doc.Concat(doc.Text("a"), doc.HardLine(), doc.Indent(doc.Concat(doc.HardLine(), doc.Text("b"))))
	got := printer.Print(d, printer.Options{})
	want := "a\n\n    b"
	if got != want {
		t.Fatalf("Print with zero-value Options = %q, want %q (LineWidth/IndentWidth/Newline defaulted)", got, want)
	}
}

func TestInsertFinalNewlineAddsExactlyOne(t *testing.T) {
	o := opts(80)
	o.InsertFinalNewline = true
	if got := printer.Print(doc.Text("a"), o); got != "a\n" {
		t.Fatalf("Print with InsertFinalNewline on no-trailing-newline input = %q, want %q", got, "a\n")
	}
	d := doc.Concat(doc.Text("a"), doc.HardLine())
	if got := printer.Print(d, o); got != "a\n" {
		t.Fatalf("Print with InsertFinalNewline on already-trailing-newline input = %q, want %q (not doubled)", got, "a\n")
	}
}

func TestNoInsertFinalNewlineStripsTrailingNewline(t *testing.T) {
	o := opts(80)
	o.InsertFinalNewline = false
	d := doc.Concat(doc.Text("a"), doc.HardLine())
	if got := printer.Print(d, o); got != "a" {
		t.Fatalf("Print with InsertFinalNewline off = %q, want %q (trailing newline stripped)", got, "a")
	}
}

func TestTrimTrailingWhitespaceStripsSpacesBeforeNewlines(t *testing.T) {
	o := opts(80)
	o.TrimTrailingWhitespace = true
	d := doc.Concat(doc.Text("a   "), doc.HardLine(), doc.Text("b"))
	want := "a\nb"
	if got := printer.Print(d, o); got != want {
		t.Fatalf("Print with TrimTrailingWhitespace = %q, want %q", got, want)
	}
}

func TestNewlineStyleCRLF(t *testing.T) {
	o := opts(80)
	o.Newline = "\r\n"
	d := doc.Concat(doc.Text("a"), doc.HardLine(), doc.Text("b"))
	want := "a\r\nb"
	if got := printer.Print(d, o); got != want {
		t.Fatalf("Print with Newline=\\r\\n = %q, want %q", got, want)
	}
}

func TestIndentStyleTabUsesTabsNotSpaces(t *testing.T) {
	o := opts(80)
	o.IndentStyle = "tab"
	d := doc.Indent(doc.Concat(doc.HardLine(), doc.Text("b")))
	want := "\n\tb"
	if got := printer.Print(d, o); got != want {
		t.Fatalf("Print with IndentStyle=tab = %q, want %q", got, want)
	}
}

func TestBreakParentForcesEnclosingGroupToBreak(t *testing.T) {
	t.Parallel()

	d := doc.Group(doc.Concat(doc.Text("a"), doc.BreakParent(), doc.Line(), doc.Text("b")))
	want := "a\nb"

	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(fitting Group with BreakParent) = %q, want %q (forced break)", got, want)
	}
}

func TestBreakParentPropagatesThroughNestedGroups(t *testing.T) {
	t.Parallel()

	inner := doc.Group(doc.Concat(doc.Text("x"), doc.BreakParent()))
	outer := doc.Group(doc.Concat(inner, doc.Line(), doc.Text("y")))
	want := "x\ny"

	if got := printer.Print(outer, opts(80)); got != want {
		t.Fatalf("Print(outer Group containing a forced-break nested Group) = %q, want %q", got, want)
	}
}

func TestLineSuffixDefersContentUntilTheNextLineBreak(t *testing.T) {
	t.Parallel()

	d := doc.Concat(doc.Text("a"), doc.LineSuffix(doc.Text("//c")), doc.Text("b"), doc.HardLine(), doc.Text("d"))
	want := "ab//c\nd"

	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(LineSuffix before a HardLine) = %q, want %q (suffix flushed right before the break)", got, want)
	}
}

func TestLineSuffixWithBreakParentForcesAGroupToBreakAndFlushesAtThatBreak(t *testing.T) {
	t.Parallel()

	d := doc.Group(doc.Concat(
		doc.Text("a"),
		doc.Concat(doc.LineSuffix(doc.Text("//c")), doc.BreakParent()),
		doc.Text(" op"),
		doc.Line(),
		doc.Text("b"),
	))
	want := "a op//c\nb"

	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(LineSuffix+BreakParent inside an otherwise-fitting Group) = %q, want %q", got, want)
	}
}

func TestLineSuffixStillFlushesAtEndOfDocumentWithNoTrailingLineBreak(t *testing.T) {
	t.Parallel()

	d := doc.Concat(doc.Text("a"), doc.LineSuffix(doc.Text("//c")))
	want := "a//c"

	if got := printer.Print(d, opts(80)); got != want {
		t.Fatalf("Print(LineSuffix with nothing after it) = %q, want %q (flushed at end, not dropped)", got, want)
	}
}
