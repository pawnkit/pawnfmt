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

	switch definition.IndentStyle {
	case editorconfig.IndentStyleSpaces:
		cfg.IndentStyle = IndentStyleSpace
	case editorconfig.IndentStyleTab:
		cfg.IndentStyle = IndentStyleTab
	}

	if definition.IndentSize != "" && definition.IndentSize != editorconfig.UnsetValue && definition.IndentSize != "tab" {
		width, err := strconv.Atoi(definition.IndentSize)
		if err != nil || width < 1 {
			return fmt.Errorf("load EditorConfig for %s: invalid indent_size %q", filename, definition.IndentSize)
		}
		cfg.IndentWidth = width
	}

	switch definition.EndOfLine {
	case editorconfig.EndOfLineLf:
		cfg.NewlineStyle = NewlineStyleLF
	case editorconfig.EndOfLineCrLf:
		cfg.NewlineStyle = NewlineStyleCRLF
	}

	if definition.InsertFinalNewline != nil {
		cfg.InsertFinalNewline = *definition.InsertFinalNewline
	}
	if definition.TrimTrailingWhitespace != nil {
		cfg.TrimTrailingWhitespace = *definition.TrimTrailingWhitespace
	}

	if value := strings.ToLower(definition.Raw["max_line_length"]); value != "" && value != "off" && value != editorconfig.UnsetValue {
		width, err := strconv.Atoi(value)
		if err != nil || width < 20 {
			return fmt.Errorf("load EditorConfig for %s: max_line_length must be at least 20 or off", filename)
		}
		cfg.LineWidth = width
	}

	return cfg.Validate()
}
