// Package pawnfmt exposes Pawn source formatting.
package pawnfmt

import (
	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

// Options controls library formatting.
type Options struct {
	TabSize   int
	UseTabs   bool
	ParseMode ParseMode
}

// ParseMode controls how formatting handles parser errors.
type ParseMode string

const (
	ParseModeStrict   ParseMode = "strict"
	ParseModeTolerant ParseMode = "tolerant"
)

// Range is a byte range in the original source.
type Range struct {
	Start int
	End   int
}

// RangeResult contains formatted source and the replaced range.
type RangeResult struct {
	Source         []byte
	FormattedRange Range
}

// Format formats a complete Pawn source file.
func Format(source []byte, opts Options) ([]byte, error) {
	cfg := optionsConfig(opts)
	return formatter.Source(source, cfg)
}

// FormatRange formats the top-level syntax unit containing the range.
func FormatRange(source []byte, start, end int, opts Options) (RangeResult, error) {
	result, err := formatter.SourceRange(source, optionsConfig(opts), start, end)
	if err != nil {
		return RangeResult{}, err
	}

	return RangeResult{
		Source:         result.Source,
		FormattedRange: Range{Start: result.FormattedRange.Start, End: result.FormattedRange.End},
	}, nil
}

func optionsConfig(opts Options) config.Config {
	cfg := config.Default()
	if opts.TabSize > 0 {
		cfg.IndentWidth = opts.TabSize
	}

	if opts.UseTabs {
		cfg.IndentStyle = config.IndentStyleTab
	}
	if opts.ParseMode != "" {
		cfg.ParseMode = config.ParseMode(opts.ParseMode)
	}

	return cfg
}
