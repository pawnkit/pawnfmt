package config

import (
	"fmt"
	"strconv"
	"strings"

	editorconfig "github.com/editorconfig/editorconfig-core-go/v2"
)

// ApplyEditorConfig applies supported EditorConfig properties for filename to
// cfg. Pawn-specific config files are expected to be loaded over the result.
func ApplyEditorConfig(filename string, cfg *Config, parser editorconfig.Parser) error {
	if filename == "" {
		return nil
	}

	loader := editorconfig.Config{Parser: parser}

	definition, err := loader.Load(filename)
	if err != nil {
		return fmt.Errorf("load EditorConfig for %s: %w", filename, err)
	}

	applyEditorConfigIndentStyle(definition, cfg)
	applyEditorConfigEndOfLine(definition, cfg)
	applyEditorConfigNewlineFlags(definition, cfg)

	if err := applyEditorConfigIndentWidth(filename, definition, cfg); err != nil {
		return err
	}

	if err := applyEditorConfigLineWidth(filename, definition, cfg); err != nil {
		return err
	}

	return cfg.Validate()
}

func applyEditorConfigIndentStyle(definition *editorconfig.Definition, cfg *Config) {
	switch definition.IndentStyle {
	case editorconfig.IndentStyleSpaces:
		cfg.IndentStyle = IndentStyleSpace
	case editorconfig.IndentStyleTab:
		cfg.IndentStyle = IndentStyleTab
	}
}

func applyEditorConfigEndOfLine(definition *editorconfig.Definition, cfg *Config) {
	switch definition.EndOfLine {
	case editorconfig.EndOfLineLf:
		cfg.NewlineStyle = NewlineStyleLF
	case editorconfig.EndOfLineCrLf:
		cfg.NewlineStyle = NewlineStyleCRLF
	}
}

func applyEditorConfigNewlineFlags(definition *editorconfig.Definition, cfg *Config) {
	if definition.InsertFinalNewline != nil {
		cfg.InsertFinalNewline = *definition.InsertFinalNewline
	}

	if definition.TrimTrailingWhitespace != nil {
		cfg.TrimTrailingWhitespace = *definition.TrimTrailingWhitespace
	}
}

func applyEditorConfigIndentWidth(filename string, definition *editorconfig.Definition, cfg *Config) error {
	if definition.IndentSize == "" || definition.IndentSize == editorconfig.UnsetValue || definition.IndentSize == "tab" {
		return nil
	}

	width, err := strconv.Atoi(definition.IndentSize)
	if err != nil || width < 1 {
		return fmt.Errorf("load EditorConfig for %s: invalid indent_size %q", filename, definition.IndentSize)
	}

	cfg.IndentWidth = width

	return nil
}

func applyEditorConfigLineWidth(filename string, definition *editorconfig.Definition, cfg *Config) error {
	value := strings.ToLower(definition.Raw["max_line_length"])
	if value == "" || value == "off" || value == editorconfig.UnsetValue {
		return nil
	}

	width, err := strconv.Atoi(value)
	if err != nil || width < 20 {
		return fmt.Errorf("load EditorConfig for %s: max_line_length must be at least 20 or off", filename)
	}

	cfg.LineWidth = width

	return nil
}
