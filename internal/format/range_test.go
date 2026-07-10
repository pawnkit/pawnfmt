package format_test

import (
	"bytes"
	"strings"
	"testing"

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
}

func TestFormatRangeRejectsSelectionAcrossTopLevelUnits(t *testing.T) {
	t.Parallel()

	source := []byte("new first=1;\nnew second=2;\n")

	_, err := formatter.SourceRange(source, config.Default(), 4, len(source)-2)
	if err == nil || !strings.Contains(err.Error(), "crosses multiple") {
		t.Fatalf("expected a cross-unit range error, got %v", err)
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
}

func TestFormatRangeRejectsInvalidBounds(t *testing.T) {
	t.Parallel()

	for _, bounds := range [][2]int{{-1, 0}, {4, 2}, {0, 100}} {
		if _, err := formatter.SourceRange([]byte("new x;\n"), config.Default(), bounds[0], bounds[1]); err == nil {
			t.Fatalf("expected invalid bounds %v to fail", bounds)
		}
	}
}
