package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

type fileResult struct {
	path      string
	source    []byte
	formatted []byte
	changed   bool
	err       error
}

func runFiles(opts *options, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)

	cfg, err := resolveConfig(opts, startDirFor(opts))
	if err != nil {
		writeErrorf(stderr, errColors, "%v", err)
		return exitConfigError
	}

	files, err := collectFiles(opts.Paths, cfg.Include, cfg.Exclude, !opts.NoGitignore)
	if err != nil {
		writeErrorf(stderr, errColors, "%v", err)
		return exitFormatError
	}

	if len(files) == 0 {
		writeErrorf(stderr, errColors, "no .pwn/.inc files found in the given paths")
		return exitConfigError
	}

	if opts.DebugTokens || opts.DebugCST || opts.DebugFormatDoc {
		return runFileDebugMode(opts, files, cfg, stdout, stderr)
	}

	return reportFileResults(opts, files, formatFilesParallel(files, cfg), stdout, stderr)
}

func runFileDebugMode(opts *options, files []string, cfg config.Config, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)
	if len(files) != 1 {
		writeErrorf(stderr, errColors, "--debug-tokens/--debug-cst/--debug-format-doc require exactly one input file")
		return exitConfigError
	}

	source, err := os.ReadFile(files[0])
	if err != nil {
		writeErrorf(stderr, errColors, "%v", err)
		return exitFormatError
	}

	code, _ := runDebugModes(opts, source, cfg, stdout, stderr)

	return code
}

func reportFileResults(opts *options, files []string, results []fileResult, stdout, stderr io.Writer) int {
	stdoutColors := colorsFor(opts.Color, stdout)
	errColors := colorsFor(opts.Color, stderr)
	anyChanged := false
	anyError := false

	for _, r := range results {
		if r.err != nil {
			writeErrorf(stderr, errColors, "%s: %v", r.path, r.err)

			anyError = true

			continue
		}

		if !r.changed {
			continue
		}

		anyChanged = true

		switch {
		case opts.Check:
			_, _ = fmt.Fprintln(stdout, stdoutColors.yellow(r.path))
		case opts.Diff:
			_, _ = fmt.Fprint(stdout, unifiedDiffColored(r.path, r.source, r.formatted, stdoutColors))
		case opts.Write:
			if err := atomicWrite(r.path, r.formatted); err != nil {
				writeErrorf(stderr, errColors, "%s: %v", r.path, err)

				anyError = true
			}
		default:
			if len(files) == 1 {
				_, _ = stdout.Write(r.formatted)
			} else {
				writeErrorf(stderr, errColors, "%s: pass --write, --check, or --diff when formatting more than one file", r.path)

				anyError = true
			}
		}
	}

	if anyError {
		return exitFormatError
	}

	if opts.Check && anyChanged {
		return exitCheckChanges
	}

	return exitOK
}

func formatFilesParallel(files []string, cfg config.Config) []fileResult {
	results := make([]fileResult, len(files))
	sem := make(chan struct{}, runtime.NumCPU())

	var wg sync.WaitGroup
	for i, path := range files {
		wg.Add(1)

		sem <- struct{}{}

		go func(i int, path string) {
			defer wg.Done()
			defer func() { <-sem }()

			results[i] = formatOneFile(path, cfg)
		}(i, path)
	}

	wg.Wait()

	return results
}

func formatOneFile(path string, cfg config.Config) fileResult {
	source, err := os.ReadFile(path)
	if err != nil {
		return fileResult{path: path, err: err}
	}

	formatted, err := formatter.Source(source, cfg)
	if err != nil {
		return fileResult{path: path, source: source, err: fmt.Errorf("format: %w", err)}
	}

	return fileResult{
		path:      path,
		source:    source,
		formatted: formatted,
		changed:   !bytes.Equal(source, formatted),
	}
}
