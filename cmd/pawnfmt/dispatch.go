package main

import (
	"fmt"
	"io"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func dispatch(opts *options, stdin io.Reader, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)
	if opts.Stdin && len(opts.Paths) > 0 {
		writeErrorf(stderr, errColors, "--stdin cannot be combined with file/directory arguments")
		return exitConfigError
	}

	if opts.PrintConfig {
		cfg, err := resolveConfig(opts, startDirFor(opts))
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

	cfg, err := resolveConfig(opts, startDirFor(opts))
	if err != nil {
		writeErrorf(stderr, errColors, "%v", err)
		return exitConfigError
	}

	if code, handled := runDebugModes(opts, source, cfg, stdout, stderr); handled {
		return code
	}

	formatted, err := formatter.FormatSource(source, cfg)
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
