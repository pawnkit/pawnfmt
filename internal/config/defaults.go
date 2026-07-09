package config

// Default returns pawnfmt's built-in default configuration.
func Default() Config {
	return Config{
		LineWidth:                      100,
		IndentStyle:                    IndentStyleSpace,
		IndentWidth:                    4,
		NewlineStyle:                   NewlineStyleAuto,
		InsertFinalNewline:             true,
		TrimTrailingWhitespace:         true,
		BraceStyle:                     BraceStyleAllman,
		KeepSimpleStatementsSingleLine: true,
		IndentCaseContents:             true,
		EmptyLineBetweenTopLevelDecls:  true,
		SpaceAroundOperators:           true,
		SpaceAfterComma:                true,
		SpaceInsideParens:              false,
		SpaceInsideBrackets:            false,
		SpaceInsideBraces:              false,
		SpaceBeforeFunctionParen:       false,
		Semicolons:                     SemicolonsPreserve,
		SingleStatementBraces:          SingleStatementBracesAlways,
		DirectiveIndent:                DirectiveIndentKeepInBlock,
		DirectiveSpacing:               true,
		IndentNestedDirectives:         true,
		AlignEnumFields:                false,
		AlignConsecutiveDeclarations:   false,
		AlignConsecutiveMacros:         false,
		AlignTrailingComments:          false,
		EnumTrailingComma:              EnumTrailingCommaAlways,
		TagColonSpacing:                TagColonSpacingCompact,
		SpaceBeforeArrayBrackets:       false,
		MultilineFunctionParams:        MultilineListAuto,
		MultilineCallArgs:              MultilineListAuto,
		FormatDisabledRegions:          false,
		BlankLinesAfterIncludeBlock:    true,
		BlankLinesBetweenPublics:       true,
		SortIncludes:                   false,
		GroupIncludesByBrackets:        false,
		CollapseBlankLines:             true,
		MaxBlankLines:                  2,
		ContinuationIndentWidth:        0,
		BreakBinaryOperator:            BinaryOperatorBreakAfter,
		IndentCaseLabels:               true,
		IndentGotoLabels:               true,
	}
}

// ApplyDefaults fills zero-valued fields with their defaults.
func (cfg *Config) ApplyDefaults() {
	defaults := Default()
	cfg.applyLayoutDefaults(defaults)
	cfg.applyStyleDefaults(defaults)
}

func (cfg *Config) applyLayoutDefaults(defaults Config) {
	if cfg.LineWidth == 0 {
		cfg.LineWidth = defaults.LineWidth
	}

	if cfg.IndentStyle == "" {
		cfg.IndentStyle = defaults.IndentStyle
	}

	if cfg.IndentWidth == 0 {
		cfg.IndentWidth = defaults.IndentWidth
	}

	if cfg.NewlineStyle == "" {
		cfg.NewlineStyle = defaults.NewlineStyle
	}

	if cfg.BraceStyle == "" {
		cfg.BraceStyle = defaults.BraceStyle
	}

	if cfg.MultilineFunctionParams == "" {
		cfg.MultilineFunctionParams = defaults.MultilineFunctionParams
	}

	if cfg.MultilineCallArgs == "" {
		cfg.MultilineCallArgs = defaults.MultilineCallArgs
	}

	if cfg.BreakBinaryOperator == "" {
		cfg.BreakBinaryOperator = defaults.BreakBinaryOperator
	}
}

func (cfg *Config) applyStyleDefaults(defaults Config) {
	if cfg.Semicolons == "" {
		cfg.Semicolons = defaults.Semicolons
	}

	if cfg.SingleStatementBraces == "" {
		cfg.SingleStatementBraces = defaults.SingleStatementBraces
	}

	if cfg.DirectiveIndent == "" {
		cfg.DirectiveIndent = defaults.DirectiveIndent
	}

	if cfg.EnumTrailingComma == "" {
		cfg.EnumTrailingComma = defaults.EnumTrailingComma
	}

	if cfg.TagColonSpacing == "" {
		cfg.TagColonSpacing = defaults.TagColonSpacing
	}
}
