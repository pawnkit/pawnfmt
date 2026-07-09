package format

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestFormatOnceIsStable(t *testing.T) {
	t.Parallel()

	paths, err := filepath.Glob(filepath.Join("..", "..", "testdata", "*", "*.pwn"))
	if err != nil {
		t.Fatalf("glob formatter fixtures: %v", err)
	}

	formatter, err := New(config.Default())
	if err != nil {
		t.Fatalf("create formatter: %v", err)
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			first, err := formatter.formatOnce(source)
			if err != nil {
				t.Fatalf("first pass: %v", err)
			}

			second, err := formatter.formatOnce(first)
			if err != nil {
				t.Fatalf("verification pass: %v", err)
			}

			if !bytes.Equal(first, second) {
				t.Fatalf("formatting requires more than one pass\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}

func TestExternalFixtureFormatOnceIsStable(t *testing.T) {
	t.Parallel()

	path := os.Getenv("PAWNFMT_ONE_PASS_FIXTURE")
	if path == "" {
		t.Skip("PAWNFMT_ONE_PASS_FIXTURE is not set")
	}

	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read external fixture: %v", err)
	}

	formatter, err := New(config.Default())
	if err != nil {
		t.Fatalf("create formatter: %v", err)
	}

	first, err := formatter.formatOnce(source)
	if err != nil {
		t.Fatalf("first pass: %v", err)
	}

	second, err := formatter.formatOnce(first)
	if err != nil {
		t.Fatalf("verification pass: %v", err)
	}

	if !bytes.Equal(first, second) {
		firstLines := strings.Split(string(first), "\n")

		secondLines := strings.Split(string(second), "\n")
		for i := 0; i < len(firstLines) && i < len(secondLines); i++ {
			if firstLines[i] != secondLines[i] {
				start := max(i-2, 0)
				end := min(i+3, len(firstLines))
				secondEnd := min(end, len(secondLines))
				t.Fatalf("external fixture requires more than one pass at line %d\nfirst:\n%s\nsecond:\n%s", i+1, strings.Join(firstLines[start:end], "\n"), strings.Join(secondLines[start:secondEnd], "\n"))
			}
		}

		t.Fatalf("external fixture requires more than one pass (line counts %d and %d)", len(firstLines), len(secondLines))
	}
}
