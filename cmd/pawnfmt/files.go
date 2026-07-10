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
	path           string
	source         []byte
	formatted      []byte
	changed        bool
	err            error
	cursorOffset   *int
	formattedRange *formatter.Range
}

func singleFileModeRequested(opts *options) bool {
	return rangeEnabled(opts) || opts.CursorOffset >= 0 || opts.OutputFormat == formatJSON
}

func tooManyFilesForSingleMode(opts *options, fileCount int) bool {
	return singleFileModeRequested(opts) && fileCount != 1
}

func runFiles(opts *options, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)

	cfg, err := resolveConfig(opts, startDirFor(opts))
	if err != nil {
		writeOptionErrorf(opts, stderr, errColors, "config", "", "%v", err)
		return exitConfigError
	}

	files, err := collectFiles(opts.Paths, cfg.Include, cfg.Exclude, !opts.NoGitignore)
	if err != nil {
		writeOptionErrorf(opts, stderr, errColors, "io", "", "%v", err)
		return exitFormatError
	}

	if len(files) == 0 {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "no .pwn/.inc files found in the given paths")
		return exitConfigError
	}

	if tooManyFilesForSingleMode(opts, len(files)) {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "range, cursor, and JSON output modes require exactly one input file")
		return exitConfigError
	}

	configs, err := resolveConfigsForFiles(opts, files)
	if err != nil {
		writeOptionErrorf(opts, stderr, errColors, "config", "", "%v", err)
		return exitConfigError
	}

	if wantsDebugMode(opts) {
		return runFileDebugMode(opts, files, configs[0], stdout, stderr)
	}

	if singleFileModeRequested(opts) {
		result := formatOneFileRequest(files[0], configs[0], opts)
		return reportFileResults(opts, files, []fileResult{result}, stdout, stderr)
	}

	return reportFileResults(opts, files, formatFilesParallel(files, configs), stdout, stderr)
}

func runFileDebugMode(opts *options, files []string, cfg config.Config, stdout, stderr io.Writer) int {
	errColors := colorsFor(opts.Color, stderr)
	if len(files) != 1 {
		writeOptionErrorf(opts, stderr, errColors, "cli", "", "--debug-tokens/--debug-cst/--debug-format-doc require exactly one input file")
		return exitConfigError
	}

	source, err := os.ReadFile(files[0])
	if err != nil {
		writeOptionErrorf(opts, stderr, errColors, "io", files[0], "%v", err)
		return exitFormatError
	}

	code, _ := runDebugModes(opts, source, cfg, stdout, stderr)

	return code
}

func shouldWriteResultToStdout(opts *options, fileCount int) bool {
	return !opts.Write && !opts.Check && !opts.Diff && fileCount == 1
}

func writeSingleFileResult(stdout io.Writer, r fileResult, outputFormat string) error {
	return writeFormatResult(stdout, formatRequestResult{
		formatted: r.formatted, cursorOffset: r.cursorOffset, formattedRange: r.formattedRange,
	}, outputFormat)
}

func applyFileChange(opts *options, r fileResult, stdoutColors, errColors cliColors, stdout, stderr io.Writer) bool {
	switch {
	case opts.Check:
		_, _ = fmt.Fprintln(stdout, stdoutColors.yellow(r.path))
		return true
	case opts.Diff:
		_, _ = fmt.Fprint(stdout, unifiedDiffColored(r.path, r.source, r.formatted, stdoutColors))
		return true
	case opts.Write:
		if err := atomicWrite(r.path, r.formatted); err != nil {
			writeOptionErrorf(opts, stderr, errColors, "io", r.path, "%v", err)
			return false
		}

		return true
	default:
		writeOptionErrorf(opts, stderr, errColors, "cli", r.path, "pass --write, --check, or --diff when formatting more than one file")

		return false
	}
}

func reportFileResults(opts *options, files []string, results []fileResult, stdout, stderr io.Writer) int {
	stdoutColors := colorsFor(opts.Color, stdout)
	errColors := colorsFor(opts.Color, stderr)
	anyChanged := false
	anyError := false

	for _, r := range results {
		if r.err != nil {
			writeOptionErrorf(opts, stderr, errColors, "format", r.path, "%v", r.err)

			anyError = true

			continue
		}

		if shouldWriteResultToStdout(opts, len(files)) {
			if err := writeSingleFileResult(stdout, r, opts.OutputFormat); err != nil {
				writeOptionErrorf(opts, stderr, errColors, "io", "", "write stdout: %v", err)

				anyError = true
			}

			continue
		}

		if !r.changed {
			continue
		}

		anyChanged = true

		if !applyFileChange(opts, r, stdoutColors, errColors, stdout, stderr) {
			anyError = true
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

func formatFilesParallel(files []string, configs []config.Config) []fileResult {
	results := make([]fileResult, len(files))
	sem := make(chan struct{}, runtime.NumCPU())

	var wg sync.WaitGroup
	for i, path := range files {
		wg.Add(1)

		sem <- struct{}{}

		go func(i int, path string, cfg config.Config) {
			defer wg.Done()
			defer func() { <-sem }()

			results[i] = formatOneFile(path, cfg)
		}(i, path, configs[i])
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

func formatOneFileRequest(path string, cfg config.Config, opts *options) fileResult {
	source, err := os.ReadFile(path)
	if err != nil {
		return fileResult{path: path, err: err}
	}

	result, err := formatSourceForOptions(source, cfg, opts)
	if err != nil {
		return fileResult{path: path, source: source, err: fmt.Errorf("format: %w", err)}
	}

	return fileResult{
		path: path, source: source, formatted: result.formatted,
		changed:      !bytes.Equal(source, result.formatted),
		cursorOffset: result.cursorOffset, formattedRange: result.formattedRange,
	}
}
