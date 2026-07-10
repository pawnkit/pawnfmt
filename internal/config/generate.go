package config

import "fmt"

// DefaultTOML renders the default config as a fully commented TOML file.
func DefaultTOML() string {
	d := Default()
	return defaultTOMLHeader(d) + defaultTOMLDirectivesAndAlignment(d) + defaultTOMLMiscAndFooter(d)
}

func defaultTOMLHeader(d Config) string {
	return fmt.Sprintf(`# pawnfmt configuration.
# Every key is shown with its default value. Change the options you care
# about, or delete a line to keep using its default.

# Optional parent config path, resolved relative to this file.
extends = %q

# Parser handling: "strict" rejects parser-broken input; "tolerant" formats
# clean regions while preserving error regions byte-for-byte.
parse_mode = %q

# Line length pawnfmt tries to stay within.
line_width = %d

# Use "space" or "tab" for indentation.
indent_style = %q
# Number of spaces per indent level. Ignored when using tabs.
indent_width = %d
# Extra indent for wrapped lines. 0 uses indent_width.
continuation_indent_width = %d

# Line endings: "auto" keeps the input style, or use "lf"/"crlf".
newline_style = %q
# End each formatted file with one newline.
insert_final_newline = %t
# Remove trailing spaces and tabs.
trim_trailing_whitespace = %t

# Brace placement: "1tbs", "allman", or "whitesmiths".
brace_style = %q
# Keep short unbraced if/while/for bodies on one line.
keep_simple_statements_single_line = %t
# Indent the statements inside each switch case/default.
indent_case_contents = %t
# Indent case/default labels inside the switch block.
indent_case_labels = %t
# Keep goto labels at the current indent. false outdents them one level.
indent_goto_labels = %t
# Add a blank line between separate top-level declarations.
empty_line_between_top_level_declarations = %t

# Add spaces around operators, like "a + b".
space_around_operators = %t
# Add a space after a unary operator, like "! x" instead of "!x".
space_after_unary_operator = %t
# Add a space after commas.
space_after_comma = %t
# Add spaces just inside parentheses.
space_inside_parens = %t
# Add spaces just inside brackets.
space_inside_brackets = %t
# Add spaces just inside array literal braces.
space_inside_braces = %t
# Add a space before a function's parameter list.
space_before_function_paren = %t
# Add a space before array brackets.
space_before_array_brackets = %t

`,
		d.Extends,
		string(d.ParseMode),
		d.LineWidth,
		string(d.IndentStyle),
		d.IndentWidth,
		d.ContinuationIndentWidth,
		string(d.NewlineStyle),
		d.InsertFinalNewline,
		d.TrimTrailingWhitespace,
		string(d.BraceStyle),
		d.KeepSimpleStatementsSingleLine,
		d.IndentCaseContents,
		d.IndentCaseLabels,
		d.IndentGotoLabels,
		d.EmptyLineBetweenTopLevelDecls,
		d.SpaceAroundOperators,
		d.SpaceAfterUnaryOperator,
		d.SpaceAfterComma,
		d.SpaceInsideParens,
		d.SpaceInsideBrackets,
		d.SpaceInsideBraces,
		d.SpaceBeforeFunctionParen,
		d.SpaceBeforeArrayBrackets,
	)
}

func defaultTOMLDirectivesAndAlignment(d Config) string {
	return fmt.Sprintf(`# Enum trailing semicolons: "preserve" or "always".
semicolons = %q
# Single-statement if/while/for/else braces: "preserve", "always", or "never".
single_statement_braces = %q

# Preprocessor indentation: "none" or "keep_in_block".
directive_indent = %q
# Add a space after '#' in directives.
directive_spacing = %t
# Indent a top-level "#if" branch's contents, including nested "#if"s.
indent_nested_directives = %t

# Align enum values in a column.
align_enum_fields = %t
# Align nearby initialized declarations.
align_consecutive_declarations = %t
# Align nearby #define values.
align_consecutive_macros = %t
# Align nearby trailing // comments.
align_trailing_comments = %t
# Enum trailing commas: "preserve" or "always".
enum_trailing_comma = %q
# Tag prefix spacing: "tight" normalizes "Float: x", "compact" normalizes
# "Float:x", "preserve" keeps input.
tag_colon_spacing = %q
# Casing for hex digits and float exponents: "upper", "lower", or "preserve".
numeric_literal_case = %q

`,
		string(d.Semicolons),
		string(d.SingleStatementBraces),
		string(d.DirectiveIndent),
		d.DirectiveSpacing,
		d.IndentNestedDirectives,
		d.AlignEnumFields,
		d.AlignConsecutiveDeclarations,
		d.AlignConsecutiveMacros,
		d.AlignTrailingComments,
		string(d.EnumTrailingComma),
		string(d.TagColonSpacing),
		string(d.NumericLiteralCase),
	)
}

func defaultTOMLMiscAndFooter(d Config) string {
	return fmt.Sprintf(`# Function parameter wrapping: "auto", "one_per_line", or "bin_pack".
multiline_function_params = %q
# Call argument wrapping: "auto", "one_per_line", or "bin_pack".
multiline_call_args = %q

# Wrapped binary operators: "after" keeps the operator at line end;
# "before" moves it to the next line.
break_binary_operator = %q

# Format code inside pawnfmt off/on regions.
format_disabled_regions = %t
# Add one blank line after the top include block.
blank_lines_after_include_block = %t
# Add one blank line between adjacent public functions.
blank_lines_between_publics = %t
# Sort top-level include blocks by path.
sort_includes = %t
# When sorting includes, place <...> includes before "..." includes.
group_includes_by_brackets = %t

# Limit repeated blank lines.
collapse_blank_lines = %t
# Maximum blank lines to keep when collapse_blank_lines is true.
max_blank_lines = %d

# Glob patterns for directory formatting. exclude wins over include.
include = []
exclude = []
`,
		string(d.MultilineFunctionParams),
		string(d.MultilineCallArgs),
		string(d.BreakBinaryOperator),
		d.FormatDisabledRegions,
		d.BlankLinesAfterIncludeBlock,
		d.BlankLinesBetweenPublics,
		d.SortIncludes,
		d.GroupIncludesByBrackets,
		d.CollapseBlankLines,
		d.MaxBlankLines,
	)
}
