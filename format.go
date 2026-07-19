// Package pawnfmt exposes Pawn source formatting.
package pawnfmt

import (
	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

// Options controls library formatting.
type Options struct {
	TabSize int
	UseTabs bool
}

// Format formats a complete Pawn source file.
func Format(source []byte, opts Options) ([]byte, error) {
	cfg := config.Default()
	if opts.TabSize > 0 {
		cfg.IndentWidth = opts.TabSize
	}

	if opts.UseTabs {
		cfg.IndentStyle = config.IndentStyleTab
	}

	return formatter.Source(source, cfg)
}
