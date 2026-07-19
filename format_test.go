package pawnfmt_test

import (
	"bytes"
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
