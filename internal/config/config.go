// Package config defines pawnfmt's configuration schema, defaults, and loading.
package config

// IndentStyle selects spaces or tabs for indentation.
type IndentStyle string

// Values for IndentStyle.
const (
	IndentStyleSpace IndentStyle = "space"
	IndentStyleTab   IndentStyle = "tab"
)

// NewlineStyle selects the output line ending.
type NewlineStyle string

// Values for NewlineStyle.
const (
	NewlineStyleAuto NewlineStyle = "auto"
	NewlineStyleLF   NewlineStyle = "lf"
	NewlineStyleCRLF NewlineStyle = "crlf"
)

// BraceStyle selects where opening braces are placed.
type BraceStyle string

// Values for BraceStyle.
const (
	BraceStyle1TBS        BraceStyle = "1tbs"
	BraceStyleAllman      BraceStyle = "allman"
	BraceStyleWhitesmiths BraceStyle = "whitesmiths"
)

// SemicolonMode controls optional enum trailing semicolons.
type SemicolonMode string

// Values for SemicolonMode.
const (
	SemicolonsPreserve SemicolonMode = "preserve"
	SemicolonsAlways   SemicolonMode = "always"
)

// SingleStatementBraces controls braces around single-statement control bodies.
type SingleStatementBraces string

// Values for SingleStatementBraces.
const (
	SingleStatementBracesPreserve SingleStatementBraces = "preserve"
	SingleStatementBracesAlways   SingleStatementBraces = "always"
	SingleStatementBracesNever    SingleStatementBraces = "never"
)

// DirectiveIndent controls preprocessor line indentation.
type DirectiveIndent string

// Values for DirectiveIndent.
const (
	DirectiveIndentNone        DirectiveIndent = "none"
	DirectiveIndentKeepInBlock DirectiveIndent = "keep_in_block"
)

// EnumTrailingComma controls the final comma in enum bodies.
type EnumTrailingComma string

// Values for EnumTrailingComma.
const (
	EnumTrailingCommaPreserve EnumTrailingComma = "preserve"
	EnumTrailingCommaAlways   EnumTrailingComma = "always"
)

// TagColonSpacing controls spacing around a tag prefix's colon.
type TagColonSpacing string

// Values for TagColonSpacing.
const (
	TagColonSpacingTight    TagColonSpacing = "tight"
	TagColonSpacingPreserve TagColonSpacing = "preserve"
	TagColonSpacingCompact  TagColonSpacing = "compact"
)

// MultilineListStyle controls how wrapped lists are laid out.
type MultilineListStyle string

// Values for MultilineListStyle.
const (
	MultilineListAuto       MultilineListStyle = "auto"
	MultilineListOnePerLine MultilineListStyle = "one_per_line"
	MultilineListBinPack    MultilineListStyle = "bin_pack"
)

// BinaryOperatorBreak controls where a wrapped binary operator is placed.
type BinaryOperatorBreak string

// Values for BinaryOperatorBreak.
const (
	BinaryOperatorBreakAfter  BinaryOperatorBreak = "after"
	BinaryOperatorBreakBefore BinaryOperatorBreak = "before"
)

// Config holds all formatting options.
type Config struct {
	LineWidth                      int                   `json:"line_width" yaml:"line_width" toml:"line_width"`
	IndentStyle                    IndentStyle           `json:"indent_style" yaml:"indent_style" toml:"indent_style"`
	IndentWidth                    int                   `json:"indent_width" yaml:"indent_width" toml:"indent_width"`
	NewlineStyle                   NewlineStyle          `json:"newline_style" yaml:"newline_style" toml:"newline_style"`
	InsertFinalNewline             bool                  `json:"insert_final_newline" yaml:"insert_final_newline" toml:"insert_final_newline"`
	TrimTrailingWhitespace         bool                  `json:"trim_trailing_whitespace" yaml:"trim_trailing_whitespace" toml:"trim_trailing_whitespace"`
	BraceStyle                     BraceStyle            `json:"brace_style" yaml:"brace_style" toml:"brace_style"`
	KeepSimpleStatementsSingleLine bool                  `json:"keep_simple_statements_single_line" yaml:"keep_simple_statements_single_line" toml:"keep_simple_statements_single_line"`
	IndentCaseContents             bool                  `json:"indent_case_contents" yaml:"indent_case_contents" toml:"indent_case_contents"`
	EmptyLineBetweenTopLevelDecls  bool                  `json:"empty_line_between_top_level_declarations" yaml:"empty_line_between_top_level_declarations" toml:"empty_line_between_top_level_declarations"`
	SpaceAroundOperators           bool                  `json:"space_around_operators" yaml:"space_around_operators" toml:"space_around_operators"`
	SpaceAfterComma                bool                  `json:"space_after_comma" yaml:"space_after_comma" toml:"space_after_comma"`
	SpaceInsideParens              bool                  `json:"space_inside_parens" yaml:"space_inside_parens" toml:"space_inside_parens"`
	SpaceInsideBrackets            bool                  `json:"space_inside_brackets" yaml:"space_inside_brackets" toml:"space_inside_brackets"`
	SpaceInsideBraces              bool                  `json:"space_inside_braces" yaml:"space_inside_braces" toml:"space_inside_braces"`
	SpaceBeforeFunctionParen       bool                  `json:"space_before_function_paren" yaml:"space_before_function_paren" toml:"space_before_function_paren"`
	Semicolons                     SemicolonMode         `json:"semicolons" yaml:"semicolons" toml:"semicolons"`
	SingleStatementBraces          SingleStatementBraces `json:"single_statement_braces" yaml:"single_statement_braces" toml:"single_statement_braces"`
	DirectiveIndent                DirectiveIndent       `json:"directive_indent" yaml:"directive_indent" toml:"directive_indent"`
	DirectiveSpacing               bool                  `json:"directive_spacing" yaml:"directive_spacing" toml:"directive_spacing"`
	IndentNestedDirectives         bool                  `json:"indent_nested_directives" yaml:"indent_nested_directives" toml:"indent_nested_directives"`
	AlignEnumFields                bool                  `json:"align_enum_fields" yaml:"align_enum_fields" toml:"align_enum_fields"`
	AlignConsecutiveDeclarations   bool                  `json:"align_consecutive_declarations" yaml:"align_consecutive_declarations" toml:"align_consecutive_declarations"`
	AlignConsecutiveMacros         bool                  `json:"align_consecutive_macros" yaml:"align_consecutive_macros" toml:"align_consecutive_macros"`
	AlignTrailingComments          bool                  `json:"align_trailing_comments" yaml:"align_trailing_comments" toml:"align_trailing_comments"`
	EnumTrailingComma              EnumTrailingComma     `json:"enum_trailing_comma" yaml:"enum_trailing_comma" toml:"enum_trailing_comma"`
	TagColonSpacing                TagColonSpacing       `json:"tag_colon_spacing" yaml:"tag_colon_spacing" toml:"tag_colon_spacing"`
	SpaceBeforeArrayBrackets       bool                  `json:"space_before_array_brackets" yaml:"space_before_array_brackets" toml:"space_before_array_brackets"`
	MultilineFunctionParams        MultilineListStyle    `json:"multiline_function_params" yaml:"multiline_function_params" toml:"multiline_function_params"`
	MultilineCallArgs              MultilineListStyle    `json:"multiline_call_args" yaml:"multiline_call_args" toml:"multiline_call_args"`
	FormatDisabledRegions          bool                  `json:"format_disabled_regions" yaml:"format_disabled_regions" toml:"format_disabled_regions"`
	BlankLinesAfterIncludeBlock    bool                  `json:"blank_lines_after_include_block" yaml:"blank_lines_after_include_block" toml:"blank_lines_after_include_block"`
	BlankLinesBetweenPublics       bool                  `json:"blank_lines_between_publics" yaml:"blank_lines_between_publics" toml:"blank_lines_between_publics"`
	SortIncludes                   bool                  `json:"sort_includes" yaml:"sort_includes" toml:"sort_includes"`
	GroupIncludesByBrackets        bool                  `json:"group_includes_by_brackets" yaml:"group_includes_by_brackets" toml:"group_includes_by_brackets"`
	CollapseBlankLines             bool                  `json:"collapse_blank_lines" yaml:"collapse_blank_lines" toml:"collapse_blank_lines"`
	MaxBlankLines                  int                   `json:"max_blank_lines" yaml:"max_blank_lines" toml:"max_blank_lines"`
	ContinuationIndentWidth        int                   `json:"continuation_indent_width" yaml:"continuation_indent_width" toml:"continuation_indent_width"`
	BreakBinaryOperator            BinaryOperatorBreak   `json:"break_binary_operator" yaml:"break_binary_operator" toml:"break_binary_operator"`
	IndentCaseLabels               bool                  `json:"indent_case_labels" yaml:"indent_case_labels" toml:"indent_case_labels"`
	IndentGotoLabels               bool                  `json:"indent_goto_labels" yaml:"indent_goto_labels" toml:"indent_goto_labels"`
	Include                        []string              `json:"include" yaml:"include" toml:"include"`
	Exclude                        []string              `json:"exclude" yaml:"exclude" toml:"exclude"`
}
