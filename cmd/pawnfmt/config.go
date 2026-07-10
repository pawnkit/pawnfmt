package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	editorconfig "github.com/editorconfig/editorconfig-core-go/v2"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func resolveConfig(opts *options, startDir string) (config.Config, error) {
	if opts.Config != "" {
		return config.LoadFile(opts.Config)
	}

	if opts.NoConfig {
		return config.Default(), nil
	}

	found, err := config.Discover(startDir)
	if err != nil {
		return config.Config{}, err
	}

	if found == "" {
		return config.Default(), nil
	}

	return config.LoadFile(found)
}

func resolveConfigForFile(opts *options, filename string) (config.Config, error) {
	if opts.NoConfig {
		return config.Default(), nil
	}

	base := config.Default()
	if err := config.ApplyEditorConfig(filename, &base, editorconfig.NewCachedParser()); err != nil {
		return config.Config{}, err
	}

	if opts.Config != "" {
		return config.LoadFileWithBase(opts.Config, base)
	}

	startDir := startDirFor(opts)
	if filename != "" {
		startDir = filepath.Dir(filename)
	}

	found, err := config.Discover(startDir)
	if err != nil {
		return config.Config{}, err
	}

	if found == "" {
		return base, nil
	}

	return config.LoadFileWithBase(found, base)
}

func resolveConfigsForFiles(opts *options, files []string) ([]config.Config, error) {
	configs := make([]config.Config, len(files))

	if opts.NoConfig {
		cfg := config.Default()
		for i := range configs {
			configs[i] = cfg
		}

		return configs, nil
	}

	parser := editorconfig.NewCachedParser()

	type cacheKey struct {
		path           string
		lineWidth      int
		indentStyle    config.IndentStyle
		indentWidth    int
		newlineStyle   config.NewlineStyle
		insertFinal    bool
		trimWhitespace bool
	}

	loaded := make(map[cacheKey]config.Config)

	for i, path := range files {
		base := config.Default()
		if err := config.ApplyEditorConfig(path, &base, parser); err != nil {
			return nil, err
		}

		found := opts.Config
		if found == "" {
			var err error

			found, err = config.Discover(filepath.Dir(path))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
		}

		if found == "" {
			configs[i] = base
			continue
		}

		key := cacheKey{
			path: found, lineWidth: base.LineWidth, indentStyle: base.IndentStyle,
			indentWidth: base.IndentWidth, newlineStyle: base.NewlineStyle,
			insertFinal: base.InsertFinalNewline, trimWhitespace: base.TrimTrailingWhitespace,
		}
		if cfg, ok := loaded[key]; ok {
			configs[i] = cfg
			continue
		}

		cfg, err := config.LoadFileWithBase(found, base)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", found, err)
		}

		loaded[key] = cfg
		configs[i] = cfg
	}

	return configs, nil
}

func startDirFor(opts *options) string {
	if len(opts.Paths) > 0 {
		abs, err := filepath.Abs(opts.Paths[0])
		if err == nil {
			if info, statErr := os.Stat(abs); statErr == nil && !info.IsDir() {
				return filepath.Dir(abs)
			}

			return abs
		}

		return filepath.Dir(opts.Paths[0])
	}

	if opts.StdinFilename != "" {
		abs, err := filepath.Abs(opts.StdinFilename)
		if err == nil {
			return filepath.Dir(abs)
		}
	}

	if wd, err := os.Getwd(); err == nil {
		return wd
	}

	return "."
}

func printResolvedConfig(cfg config.Config, w io.Writer) error {
	enc := toml.NewEncoder(w)
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return nil
}

func runInitConfig(opts *options, stdout, stderr io.Writer) int {
	stdoutColors := colorsFor(opts.Color, stdout)
	errColors := colorsFor(opts.Color, stderr)

	target := "pawnfmt.toml"
	if len(opts.Paths) > 0 {
		target = opts.Paths[0]
	}

	f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) //nolint:gosec // config file is meant to be readable/committed, not a secret
	if err != nil {
		if os.IsExist(err) {
			writeOptionErrorf(opts, stderr, errColors, "config", target, "already exists; remove it or pass a different path")
		} else {
			writeOptionErrorf(opts, stderr, errColors, "io", target, "%v", err)
		}

		return exitConfigError
	}

	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(config.DefaultTOML()); err != nil {
		writeOptionErrorf(opts, stderr, errColors, "io", target, "write: %v", err)
		return exitInternalError
	}

	_, _ = fmt.Fprintf(stdout, "%s %s\n", stdoutColors.green("wrote"), target)

	return exitOK
}
