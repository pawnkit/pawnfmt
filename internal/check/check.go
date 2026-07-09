// Package check runs formatter correctness checks used by the corpus report.
package check

import (
	"bytes"
	"fmt"

	"github.com/pawnkit/pawn-parser"
)

// ParsesCleanly reports whether source parses without errors.
func ParsesCleanly(source []byte) (bool, error) {
	f := parser.Parse(source)
	return !f.HasParseErrors(), nil
}

// Idempotent reports whether formatting an already-formatted source is a no-op.
func Idempotent(formatted []byte, fn func([]byte) ([]byte, error)) (bool, error) {
	second, err := fn(formatted)
	if err != nil {
		return false, fmt.Errorf("re-format for idempotency check: %w", err)
	}

	return bytes.Equal(formatted, second), nil
}
