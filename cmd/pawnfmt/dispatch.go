package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func dispatch(opts *options, stdin io.Reader, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)
	if (opts.RangeStart >= 0) != (opts.RangeEnd >= 0) {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "--range-start and --range-end must be provided together")
		return exitConfigError
	}
	if rangeEnabled(opts) && (opts.DebugTokens || opts.DebugCST || opts.DebugFormatDoc) {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "range formatting cannot be combined with debug modes")
		return exitConfigError
	}
	if opts.CursorOffset >= 0 && opts.OutputFormat != "json" {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "--cursor-offset requires --output-format=json")
		return exitConfigError
	}
	if opts.OutputFormat == "json" && (opts.Write || opts.Check || opts.Diff) {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "--output-format=json cannot be combined with --write, --check, or --diff")
		return exitConfigError
	}
	if opts.OutputFormat == "json" && (opts.DebugTokens || opts.DebugCST || opts.DebugFormatDoc) {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "--output-format=json cannot be combined with debug modes")
		return exitConfigError
	}
	if opts.Stdin && len(opts.Paths) > 0 {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "--stdin cannot be combined with file/directory arguments")
		return exitConfigError
	}

	if opts.PrintConfig {
		filename := opts.StdinFilename
		if filename == "" && len(opts.Paths) > 0 {
			filename = opts.Paths[0]
		}
		if info, err := os.Stat(filename); err == nil && info.IsDir() {
			filename = ""
		}

		cfg, err := resolveConfigForFile(opts, filename)
		if err != nil {
			writeOptionErrorf(opts, stderr, errColors, "config", "", "%v", err)
			return exitConfigError
		}

		if err := printResolvedConfig(cfg, stdout); err != nil {
			writeOptionErrorf(opts, stderr, errColors, "internal", "", "%v", err)
			return exitInternalError
		}

		return exitOK
	}

	if opts.InitConfig {
		return runInitConfig(opts, stdout, stderr)
	}

	if opts.Stdin {
		return runStdin(opts, stdin, stdout, stderr)
	}

	if len(opts.Paths) == 0 {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "no input; pass file/directory paths or use --stdin")
		return exitConfigError
	}

	return runFiles(opts, stdout, stderr)
}

func runStdin(opts *options, stdin io.Reader, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)

	source, err := io.ReadAll(stdin)
	if err != nil {
		writeOptionErrorf(opts, stderr, errColors, "io", opts.StdinFilename, "read stdin: %v", err)
		return exitFormatError
	}

	cfg, err := resolveConfigForFile(opts, opts.StdinFilename)
	if err != nil {
		writeOptionErrorf(opts, stderr, errColors, "config", opts.StdinFilename, "%v", err)
		return exitConfigError
	}

	if code, handled := runDebugModes(opts, source, cfg, stdout, stderr); handled {
		return code
	}

	result, err := formatSourceForOptions(source, cfg, opts)
	if err != nil {
		writeOptionErrorf(opts, stderr, errColors, "format", opts.StdinFilename, "%v", err)
		return exitFormatError
	}

	if err := writeFormatResult(stdout, result, opts.OutputFormat); err != nil {
		writeOptionErrorf(opts, stderr, errColors, "io", "", "write stdout: %v", err)
		return exitInternalError
	}

	return exitOK
}

func rangeEnabled(opts *options) bool {
	return opts.RangeStart >= 0 && opts.RangeEnd >= 0
}

type formatRequestResult struct {
	formatted      []byte
	cursorOffset   *int
	formattedRange *formatter.Range
}

func formatSourceForOptions(source []byte, cfg config.Config, opts *options) (formatRequestResult, error) {
	if rangeEnabled(opts) {
		result, err := formatter.SourceRange(source, cfg, opts.RangeStart, opts.RangeEnd)
		if err != nil {
			return formatRequestResult{}, err
		}

		request := formatRequestResult{formatted: result.Source, formattedRange: &result.FormattedRange}
		if opts.CursorOffset >= 0 {
			adjusted, err := formatter.AdjustCursorOffset(source, result.Source, opts.CursorOffset)
			if err != nil {
				return formatRequestResult{}, err
			}
			request.cursorOffset = &adjusted
		}
		return request, nil
	}

	if opts.CursorOffset >= 0 {
		result, err := formatter.SourceWithCursor(source, cfg, opts.CursorOffset)
		if err != nil {
			return formatRequestResult{}, err
		}
		return formatRequestResult{formatted: result.Source, cursorOffset: &result.CursorOffset}, nil
	}

	formatted, err := formatter.Source(source, cfg)
	return formatRequestResult{formatted: formatted}, err
}

func writeFormatResult(w io.Writer, result formatRequestResult, outputFormat string) error {
	if outputFormat == "text" {
		_, err := w.Write(result.formatted)
		return err
	}

	type jsonRange struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}
	payload := struct {
		Formatted      string     `json:"formatted"`
		CursorOffset   *int       `json:"cursor_offset,omitempty"`
		FormattedRange *jsonRange `json:"formatted_range,omitempty"`
	}{Formatted: string(result.formatted), CursorOffset: result.cursorOffset}
	if result.formattedRange != nil {
		payload.FormattedRange = &jsonRange{Start: result.formattedRange.Start, End: result.formattedRange.End}
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(payload)
}

func runDebugModes(opts *options, source []byte, cfg config.Config, stdout, stderr io.Writer) (code int, handled bool) {
	switch {
	case opts.DebugTokens:
		debugTokens(source, stdout)
		return exitOK, true
	case opts.DebugCST:
		_, _ = fmt.Fprintln(stdout, formatter.DebugCST(source))
		return exitOK, true
	case opts.DebugFormatDoc:
		s, err := formatter.DebugDocTree(source, cfg)
		if err != nil {
			writeOptionErrorf(opts, stderr, colorsFor(opts.Color, stderr), "format", opts.StdinFilename, "%v", err)
			return exitFormatError, true
		}

		_, _ = fmt.Fprintln(stdout, s)

		return exitOK, true
	default:
		return exitOK, false
	}
}
