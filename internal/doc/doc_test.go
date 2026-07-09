package doc_test

import (
	"testing"

	"github.com/pawnkit/pawnfmt/internal/doc"
)

func TestConcatFiltersNilAndCollapsesSingle(t *testing.T) {
	t.Parallel()

	if got := doc.Concat(); got != (doc.TextDoc{Value: ""}) {
		t.Fatalf("Concat() = %#v, want empty TextDoc", got)
	}

	only := doc.Text("a")
	if got := doc.Concat(nil, only, nil); got != only {
		t.Fatalf("Concat(nil, only, nil) = %#v, want the single non-nil part unwrapped", got)
	}

	got := doc.Concat(doc.Text("a"), doc.Text("b"))

	multi, ok := got.(doc.ConcatDoc)
	if !ok || len(multi.Parts) != 2 {
		t.Fatalf("Concat(a, b) = %#v, want a 2-part ConcatDoc", got)
	}
}

func TestIndentResetIndentOutdentNilCollapseToEmptyText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		fn   func(doc.Doc) doc.Doc
	}{
		{"Indent", doc.Indent},
		{"ResetIndent", doc.ResetIndent},
		{"Outdent", doc.Outdent},
		{"Group", doc.Group},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.fn(nil); got != (doc.TextDoc{Value: ""}) {
				t.Fatalf("%s(nil) = %#v, want empty TextDoc", tc.name, got)
			}
		})
	}
}

func TestJoinInsertsSeparatorAndFiltersNil(t *testing.T) {
	t.Parallel()

	got := doc.Join(doc.Text(","), doc.Text("a"), nil, doc.Text("b"))

	concat, ok := got.(doc.ConcatDoc)
	if !ok {
		t.Fatalf("Join(...) = %#v, want a ConcatDoc", got)
	}

	want := []doc.Doc{doc.Text("a"), doc.Text(","), doc.Text("b")}
	if len(concat.Parts) != len(want) {
		t.Fatalf("Join(...) parts = %#v, want %#v", concat.Parts, want)
	}

	for i := range want {
		if concat.Parts[i] != want[i] {
			t.Fatalf("Join(...) part %d = %#v, want %#v", i, concat.Parts[i], want[i])
		}
	}
}

func TestJoinAllNilReturnsEmptyText(t *testing.T) {
	t.Parallel()

	if got := doc.Join(doc.Text(",")); got != (doc.TextDoc{Value: ""}) {
		t.Fatalf("Join() with no parts = %#v, want empty TextDoc", got)
	}
}

func TestFillFiltersNilAndCollapsesSingle(t *testing.T) {
	t.Parallel()

	only := doc.Text("a")
	if got := doc.Fill(nil, only, nil); got != only {
		t.Fatalf("Fill(nil, only, nil) = %#v, want the single non-nil part unwrapped", got)
	}

	got := doc.Fill(doc.Text("a"), doc.Line(), doc.Text("b"))

	fill, ok := got.(doc.FillDoc)
	if !ok || len(fill.Parts) != 3 {
		t.Fatalf("Fill(a, line, b) = %#v, want a 3-part FillDoc", got)
	}
}

func TestRawTextBlockEmptyIsEmptyText(t *testing.T) {
	t.Parallel()

	if got := doc.RawTextBlock(""); got != (doc.TextDoc{Value: ""}) {
		t.Fatalf("RawTextBlock(\"\") = %#v, want empty TextDoc", got)
	}
}

func TestRawTextBlockSingleLineStaysUnwrapped(t *testing.T) {
	t.Parallel()

	got := doc.RawTextBlock("no newline here")
	if got != (doc.TextDoc{Value: "no newline here"}) {
		t.Fatalf("RawTextBlock with no \\n = %#v, want a bare TextDoc (no ResetIndent wrapper)", got)
	}
}

func TestRawTextBlockMultiLineWrapsInResetIndent(t *testing.T) {
	t.Parallel()

	got := doc.RawTextBlock("a\nb")

	reset, ok := got.(doc.ResetIndentDoc)
	if !ok {
		t.Fatalf("RawTextBlock(\"a\\nb\") = %#v, want a ResetIndentDoc wrapper", got)
	}

	concat, ok := reset.Contents.(doc.ConcatDoc)
	if !ok || len(concat.Parts) != 3 {
		t.Fatalf("RawTextBlock(\"a\\nb\") contents = %#v, want [Text(a), HardLine, Text(b)]", reset.Contents)
	}

	if concat.Parts[0] != (doc.TextDoc{Value: "a"}) || concat.Parts[1] != (doc.HardLineDoc{}) || concat.Parts[2] != (doc.TextDoc{Value: "b"}) {
		t.Fatalf("RawTextBlock(\"a\\nb\") parts = %#v, want [Text(a), HardLine, Text(b)]", concat.Parts)
	}
}

func TestRawTextBlockStripsTrailingCROnEachLine(t *testing.T) {
	t.Parallel()

	got := doc.RawTextBlock("a\r\nb")

	reset, ok := got.(doc.ResetIndentDoc)
	if !ok {
		t.Fatalf("RawTextBlock(\"a\\r\\nb\") = %#v, want a ResetIndentDoc", got)
	}

	concat, ok := reset.Contents.(doc.ConcatDoc)
	if !ok {
		t.Fatalf("RawTextBlock(\"a\\r\\nb\") contents = %#v, want a ConcatDoc", reset.Contents)
	}

	if concat.Parts[0] != (doc.TextDoc{Value: "a"}) {
		t.Fatalf("RawTextBlock(\"a\\r\\nb\") first line = %#v, want Text(a) with \\r stripped", concat.Parts[0])
	}
}

func TestRawTextBlockTrailingNewlineProducesNoDanglingEmptyText(t *testing.T) {
	t.Parallel()

	got := doc.RawTextBlock("a\n")

	reset, ok := got.(doc.ResetIndentDoc)
	if !ok {
		t.Fatalf("RawTextBlock(\"a\\n\") = %#v, want a ResetIndentDoc", got)
	}

	concat, ok := reset.Contents.(doc.ConcatDoc)
	if !ok || len(concat.Parts) != 2 {
		t.Fatalf("RawTextBlock(\"a\\n\") contents = %#v, want exactly [Text(a), HardLine] with no trailing empty Text", reset.Contents)
	}
}
