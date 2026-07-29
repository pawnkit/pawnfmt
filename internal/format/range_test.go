package format_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func TestFormatRangeFormatsOneTopLevelUnitOnly(t *testing.T) {
	t.Parallel()

	source := []byte("stock First(){new   value=1;return value;}\nstock Second(){new   untouched=2;return untouched;}\n")
	secondStart := bytes.Index(source, []byte("stock Second"))
	selection := bytes.Index(source, []byte("value=1"))

	result, err := formatter.SourceRange(source, config.Default(), selection, selection+len("value=1"))
	if err != nil {
		t.Fatalf("SourceRange: %v", err)
	}

	if !strings.Contains(string(result.Source), "new value = 1;") {
		t.Fatalf("selected function was not formatted:\n%s", result.Source)
	}

	formattedSecondStart := secondStart + (len(result.Source) - len(source))
	if !bytes.Equal(result.Source[formattedSecondStart:], source[secondStart:]) {
		t.Fatalf("source after the selected function changed\nexpected:\n%s\nactual:\n%s",
			source[secondStart:], result.Source[formattedSecondStart:])
	}

	if result.FormattedRange.Start > selection || selection >= result.FormattedRange.End {
		t.Fatalf("expanded range [%d,%d) does not contain selection %d",
			result.FormattedRange.Start, result.FormattedRange.End, selection)
	}

	if parsed := parser.Parse(result.Source); parsed.HasParseErrors() {
		t.Fatalf("range output does not parse cleanly: %v", parsed.Diagnostics)
	}
}

func TestFormatRangeFormatsSelectionAcrossTopLevelUnits(t *testing.T) {
	t.Parallel()

	source := []byte("new first=1;\nnew second=2;\nnew   untouched=3;\n")

	result, err := formatter.SourceRange(source, config.Default(), 4, bytes.Index(source, []byte("new   untouched")))
	if err != nil {
		t.Fatalf("SourceRange: %v", err)
	}

	if got := string(result.Source); got != "new first = 1;\nnew second = 2;\nnew   untouched=3;\n" {
		t.Fatalf("unexpected range output:\n%s", got)
	}

	if result.FormattedRange.Start != 0 || result.FormattedRange.End != bytes.Index(source, []byte("\nnew   untouched")) {
		t.Fatalf("formatted range = %#v", result.FormattedRange)
	}
}

func TestFormatRangeKeepsMacroGroupAlignment(t *testing.T) {
	t.Parallel()

	source := []byte("#define SHORT       1\n#define MUCH_LONGER 2\n")

	result, err := formatter.SourceRange(source, config.Default(), 0, len(source))
	if err != nil {
		t.Fatalf("SourceRange: %v", err)
	}

	if !bytes.Equal(result.Source, source) {
		t.Fatalf("range formatting removed alignment:\n%s", result.Source)
	}
}

func TestFormatRangeExpandsAcrossAlignedMacroGroup(t *testing.T) {
	t.Parallel()

	source := []byte("#define SHORT       1\n#define MUCH_LONGER 2\n")

	result, err := formatter.SourceRange(source, config.Default(), 8, 13)
	if err != nil {
		t.Fatalf("SourceRange: %v", err)
	}

	if !bytes.Equal(result.Source, source) {
		t.Fatalf("range formatting removed adjacent alignment:\n%s", result.Source)
	}
}

func TestFormatRangeMapsIntentionalControlBodyBraces(t *testing.T) {
	t.Parallel()

	source := []byte("stock F(){if(x) return   1;new   untouched=2;}\n")
	start := bytes.Index(source, []byte("return"))

	result, err := formatter.SourceRange(source, config.Default(), start, start+len("return"))
	if err != nil {
		t.Fatalf("SourceRange: %v", err)
	}

	text := string(result.Source)
	if !strings.Contains(text, "if(x) {\n        return 1;\n    }") {
		t.Fatalf("selected control body did not gain safely mapped braces:\n%s", text)
	}

	if !strings.Contains(text, "new   untouched=2;") {
		t.Fatalf("unselected sibling statement was formatted:\n%s", text)
	}

	if parsed := parser.Parse(result.Source); parsed.HasParseErrors() {
		t.Fatalf("range output does not parse cleanly: %v", parsed.Diagnostics)
	}
}

func TestFormatRangeRejectsInvalidBounds(t *testing.T) {
	t.Parallel()

	for _, bounds := range [][2]int{{-1, 0}, {4, 2}, {0, 100}} {
		if _, err := formatter.SourceRange([]byte("new x;\n"), config.Default(), bounds[0], bounds[1]); err == nil {
			t.Fatalf("expected invalid bounds %v to fail", bounds)
		}
	}
}
