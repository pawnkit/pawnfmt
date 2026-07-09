package main

import (
	"fmt"
	"io"
	"os"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func dispatch(opts *options, stdin io.Reader, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)
	if (opts.RangeStart >= 0) != (opts.RangeEnd >= 0) {
		writeErrorf(stderr, errColors, "--range-start and --range-end must be provided together")
		return exitConfigError
	}
	if rangeEnabled(opts) && (opts.DebugTokens || opts.DebugCST || opts.DebugFormatDoc) {
		writeErrorf(stderr, errColors, "range formatting cannot be combined with debug modes")
		return exitConfigError
	}
	if opts.Stdin && len(opts.Paths) > 0 {
		writeErrorf(stderr, errColors, "--stdin cannot be combined with file/directory arguments")
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
			writeErrorf(stderr, errColors, "%v", err)
			return exitConfigError
		}

		if err := printResolvedConfig(cfg, stdout); err != nil {
			writeErrorf(stderr, errColors, "%v", err)
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
		writeErrorf(stderr, errColors, "no input; pass file/directory paths or use --stdin")
		return exitConfigError
	}

	return runFiles(opts, stdout, stderr)
}

func runStdin(opts *options, stdin io.Reader, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)

	source, err := io.ReadAll(stdin)
	if err != nil {
		writeErrorf(stderr, errColors, "read stdin: %v", err)
		return exitFormatError
	}

	cfg, err := resolveConfigForFile(opts, opts.StdinFilename)
	if err != nil {
		writeErrorf(stderr, errColors, "%v", err)
		return exitConfigError
	}

	if code, handled := runDebugModes(opts, source, cfg, stdout, stderr); handled {
		return code
	}

	formatted, err := formatSourceForOptions(source, cfg, opts)
	if err != nil {
		writeErrorf(stderr, errColors, "%v", err)
		return exitFormatError
	}

	if _, err := stdout.Write(formatted); err != nil {
		writeErrorf(stderr, errColors, "write stdout: %v", err)
		return exitInternalError
	}

	return exitOK
}

func rangeEnabled(opts *options) bool {
	return opts.RangeStart >= 0 && opts.RangeEnd >= 0
}

func formatSourceForOptions(source []byte, cfg config.Config, opts *options) ([]byte, error) {
	if !rangeEnabled(opts) {
		return formatter.Source(source, cfg)
	}

	result, err := formatter.SourceRange(source, cfg, opts.RangeStart, opts.RangeEnd)
	if err != nil {
		return nil, err
	}
	return result.Source, nil
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
			writeErrorf(stderr, colorsFor(opts.Color, stderr), "%v", err)
			return exitFormatError, true
		}

		_, _ = fmt.Fprintln(stdout, s)

		return exitOK, true
	default:
		return exitOK, false
	}
}
