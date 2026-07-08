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

type Formatter struct {
	config config.Config
}

func New(cfg config.Config) (*Formatter, error) {
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Formatter{config: cfg}, nil
}

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
		return nil, errors.New("source does not parse cleanly")
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
		return nil, errors.New("formatted output does not parse cleanly")
	}

	return []byte(formatted), nil
}

func FormatSource(source []byte, cfg config.Config) ([]byte, error) {
	formatter, err := New(cfg)
	if err != nil {
		return nil, err
	}

	return formatter.FormatSource(source)
}
