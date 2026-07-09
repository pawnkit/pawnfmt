package format_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestGolden(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join(testdataDir(), "input", "*.pwn"))
	if err != nil {
		t.Fatalf("glob golden inputs: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no golden inputs found")
	}

	for _, inputPath := range files {
		t.Run(strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)), func(t *testing.T) {
			t.Parallel()
			source := readFile(t, inputPath)
			formatted := mustFormat(t, source, config.Default())
			expectedPath := filepath.Join(testdataDir(), "expected", filepath.Base(inputPath))

			expected := ensureTrailingNewline(readFile(t, expectedPath))
			if string(formatted) != string(expected) {
				t.Fatalf("formatted output mismatch\nexpected:\n%s\nactual:\n%s", expected, formatted)
			}
		})
	}
}

func TestIdempotence(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join(testdataDir(), "idempotence", "*.pwn"))
	if err != nil {
		t.Fatalf("glob idempotence inputs: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no idempotence inputs found")
	}

	for _, inputPath := range files {
		t.Run(strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)), func(t *testing.T) {
			t.Parallel()
			source := readFile(t, inputPath)
			first := mustFormat(t, source, config.Default())

			second := mustFormat(t, first, config.Default())
			if string(first) != string(second) {
				t.Fatalf("formatter is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}

func TestParseAfterFormat(t *testing.T) {
	t.Parallel()

	directories := []string{"input", "idempotence"}
	for _, directory := range directories {
		files, err := filepath.Glob(filepath.Join(testdataDir(), directory, "*.pwn"))
		if err != nil {
			t.Fatalf("glob %s: %v", directory, err)
		}

		for _, inputPath := range files {
			t.Run(directory+"/"+strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)), func(t *testing.T) {
				t.Parallel()
				formatted := mustFormat(t, readFile(t, inputPath), config.Default())

				parsed := parser.Parse(formatted)
				if parsed.HasParseErrors() {
					t.Fatal("formatted output has parse errors")
				}
			})
		}
	}
}
