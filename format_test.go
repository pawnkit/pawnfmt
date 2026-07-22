package pawnfmt_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	formatted, err := pawnfmt.Format([]byte("main(){new value=1;}\n"), pawnfmt.Options{TabSize: 2})
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(formatted, []byte("main(){new value=1;}\n")) {
		t.Fatalf("source was not formatted: %q", formatted)
	}
}

func TestFormatToleratesBrokenRegions(t *testing.T) {
	t.Parallel()

	source := []byte("new   first=1;\n}\n#include <YSI_Server\\y_flooding>\nnew   second=2;\n")
	formatted, err := pawnfmt.Format(source, pawnfmt.Options{ParseMode: pawnfmt.ParseModeTolerant})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"new first = 1;", "\n}\n", "#include <YSI_Server\\y_flooding>", "new second = 2;"} {
		if !strings.Contains(string(formatted), want) {
			t.Fatalf("formatted source missing %q:\n%s", want, formatted)
		}
	}
}

func TestFormatZeroOptionsUseDefaults(t *testing.T) {
	t.Parallel()

	formatted, err := pawnfmt.Format([]byte("main(){new value=1;}\n"), pawnfmt.Options{})
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Contains(formatted, []byte("\t")) {
		t.Fatalf("zero options produced tab indentation: %q", formatted)
	}
}

func TestFormatRange(t *testing.T) {
	t.Parallel()

	source := []byte("stock First(){new value=1;}\nstock Second(){new value=2;}\n")
	start := bytes.Index(source, []byte("value=2"))

	result, err := pawnfmt.FormatRange(source, start, start+len("value=2"), pawnfmt.Options{})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(result.Source, []byte("new value = 2;")) {
		t.Fatalf("selected function was not formatted: %s", result.Source)
	}

	if !bytes.Contains(result.Source, []byte("First(){new value=1;}")) {
		t.Fatalf("unselected function changed: %s", result.Source)
	}

	if result.FormattedRange.Start > start || result.FormattedRange.End <= start {
		t.Fatalf("formatted range = %+v", result.FormattedRange)
	}
}
