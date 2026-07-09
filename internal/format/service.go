package format

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/printer"
	"github.com/pawnkit/pawnfmt/internal/trivia"
)

// Formatter formats Pawn source using a fixed configuration.
type Formatter struct {
	config config.Config
}

// New builds a Formatter after defaulting and validating cfg.
func New(cfg config.Config) (*Formatter, error) {
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Formatter{config: cfg}, nil
}

// FormatSource formats source, reformatting up to 4 times until it converges.
func (formatter *Formatter) FormatSource(source []byte) ([]byte, error) {
	current := source
	for range 4 {
		formatted, err := formatter.formatOnce(current)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(formatted, current) {
			return formatted, nil
		}

		current = formatted
	}

	return nil, errors.New("formatting did not converge after 4 passes")
}

func (formatter *Formatter) formatOnce(source []byte) ([]byte, error) {
	parsed := parser.Parse(source)
	if parsed.HasParseErrors() {
		return nil, parseDiagnostic(source, parsed, "source")
	}

	index := trivia.Scan(source)
	st := newState(parsed, formatter.config, index)

	formatted := printer.Print(st.formatNode(parsed.Root), st.printerOptions())
	if !formatter.config.SortIncludes {
		if err := verifySemanticTokens(source, []byte(formatted)); err != nil {
			return nil, fmt.Errorf("formatted output changed source semantics: %w", err)
		}
	}

	verified := parser.Parse([]byte(formatted))
	if verified.HasParseErrors() {
		return nil, parseDiagnostic([]byte(formatted), verified, "formatted output")
	}

	return []byte(formatted), nil
}

// Source formats source using cfg. It is a convenience wrapper around New.
func Source(source []byte, cfg config.Config) ([]byte, error) {
	formatter, err := New(cfg)
	if err != nil {
		return nil, err
	}

	return formatter.FormatSource(source)
}
