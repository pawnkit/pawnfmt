package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Validate checks that cfg's fields hold legal values.
func (cfg Config) Validate() error {
	if err := cfg.validateNumbers(); err != nil {
		return err
	}

	return cfg.validateEnums()
}

func (cfg Config) validateNumbers() error {
	if cfg.LineWidth < 20 {
		return errors.New("line_width must be at least 20")
	}

	if cfg.IndentWidth < 1 {
		return errors.New("indent_width must be at least 1")
	}

	if cfg.MaxBlankLines < 0 {
		return errors.New("max_blank_lines must be at least 0")
	}

	if cfg.ContinuationIndentWidth < 0 {
		return errors.New("continuation_indent_width must be at least 0 (0 means match indent_width)")
	}

	return nil
}

func (cfg Config) validateEnums() error {
	if err := oneOf("indent_style", string(cfg.IndentStyle), string(IndentStyleSpace), string(IndentStyleTab)); err != nil {
		return err
	}

	if err := oneOf("newline_style", string(cfg.NewlineStyle), string(NewlineStyleAuto), string(NewlineStyleLF), string(NewlineStyleCRLF)); err != nil {
		return err
	}

	if err := oneOf("brace_style", string(cfg.BraceStyle), string(BraceStyle1TBS), string(BraceStyleAllman), string(BraceStyleWhitesmiths)); err != nil {
		return err
	}

	if err := oneOf("semicolons", string(cfg.Semicolons), string(SemicolonsPreserve), string(SemicolonsAlways)); err != nil {
		return err
	}

	if err := oneOf("single_statement_braces", string(cfg.SingleStatementBraces), string(SingleStatementBracesPreserve), string(SingleStatementBracesAlways), string(SingleStatementBracesNever)); err != nil {
		return err
	}

	if err := oneOf("directive_indent", string(cfg.DirectiveIndent), string(DirectiveIndentNone), string(DirectiveIndentKeepInBlock)); err != nil {
		return err
	}

	if err := oneOf("enum_trailing_comma", string(cfg.EnumTrailingComma), string(EnumTrailingCommaPreserve), string(EnumTrailingCommaAlways)); err != nil {
		return err
	}

	if err := oneOf("tag_colon_spacing", string(cfg.TagColonSpacing), string(TagColonSpacingTight), string(TagColonSpacingPreserve), string(TagColonSpacingCompact)); err != nil {
		return err
	}

	if err := oneOf("numeric_literal_case", string(cfg.NumericLiteralCase), string(NumericLiteralCasePreserve), string(NumericLiteralCaseUpper), string(NumericLiteralCaseLower)); err != nil {
		return err
	}

	if err := oneOf("multiline_function_params", string(cfg.MultilineFunctionParams), string(MultilineListAuto), string(MultilineListOnePerLine), string(MultilineListBinPack)); err != nil {
		return err
	}

	if err := oneOf("multiline_call_args", string(cfg.MultilineCallArgs), string(MultilineListAuto), string(MultilineListOnePerLine), string(MultilineListBinPack)); err != nil {
		return err
	}

	return oneOf("break_binary_operator", string(cfg.BreakBinaryOperator), string(BinaryOperatorBreakAfter), string(BinaryOperatorBreakBefore))
}

func oneOf(name, value string, options ...string) error {
	if slices.Contains(options, value) {
		return nil
	}

	return fmt.Errorf("%s must be one of %s", name, strings.Join(options, ", "))
}
