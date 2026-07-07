package config

import "fmt"

func DefaultTOML() string {
	d := Default()
	return fmt.Sprintf(`# pawnfmt configuration.
# Every key below is shown at its default value — edit in place. Delete a
# line (or the whole file) to fall back to that default. See
# https://github.com/pawnkit/pawnfmt for the full option reference.

# Column to wrap lines at.
line_width = %d

# Indentation: "space" or "tab".
indent_style = %q
# Spaces per indent level. Ignored when indent_style = "tab".
indent_width = %d
# Spaces used to indent a wrapped line's continuation (e.g. a long binary
# expression or call argument list that spilled onto more than one line).
# 0 means "same as indent_width" — set a larger value to visually
# distinguish a wrapped expression from a nested block.
continuation_indent_width = %d

# Line ending: "auto" (match the input file), "lf", or "crlf".
newline_style = %q
# Ensure the output ends with exactly one newline.
insert_final_newline = %t
# Strip trailing spaces/tabs from every line, including inside multi-line
# comments and raw-preserved regions.
trim_trailing_whitespace = %t

# Opening brace placement for functions, if/while/for/do, switch, enum,
# and macro-loop blocks: "1tbs" (same line), "allman" (own line), or
# "whitesmiths" (own line, indented one level further).
brace_style = %q
# Keep an unbraced if/while/for body on the same line as its header
# ("if (x) return 1;") instead of always breaking it onto its own line.
keep_simple_statements_single_line = %t
# Indent a switch case/default clause's body one level past its label.
indent_case_contents = %t
# Indent "case"/"default" labels themselves one level under their
# enclosing switch's own brace. false keeps labels flush with the switch
# statement's own column instead.
indent_case_labels = %t
# Indent a goto label at the current statement's ambient indent (true) or
# outdent it one level, e.g. flush with the enclosing block's brace
# (false) — goto is rare in idiomatic Pawn, but some styles still outdent
# labels to make them stand out.
indent_goto_labels = %t
# Separate top-level declaration groups with one blank line. Related runs of
# variables and forward/native declarations remain compact.
empty_line_between_top_level_declarations = %t

# Space on both sides of binary/assignment operators: "a + b" vs "a+b".
space_around_operators = %t
# Space after ',' in lists, parameters, and arguments: "a, b" vs "a,b".
space_after_comma = %t
# Pad "(...)" with inner spaces: "( x )" vs "(x)".
space_inside_parens = %t
# Pad "[...]" with inner spaces: "[ 0 ]" vs "[0]".
space_inside_brackets = %t
# Pad array-literal "{...}" with inner spaces: "{ 1, 2 }" vs "{1, 2}".
space_inside_braces = %t
# Space between a function's name and its parameter list: "Foo (x)" vs
# "Foo(x)".
space_before_function_paren = %t
# Space between a name and its array dimensions: "arr []" vs "arr[]".
space_before_array_brackets = %t

# An enum declaration's trailing ';' — the one genuinely optional
# semicolon in Pawn: "preserve" (keep what the source had) or "always".
semicolons = %q
# An unbraced if/while/for/else body: "preserve" (keep as parsed),
# "always" (wrap in "{ }"), or "never" (unwrap a single-statement block).
single_statement_braces = %q

# A preprocessor line's column inside a nested block: "none" (always
# column 0) or "keep_in_block" (indent to match the surrounding code).
directive_indent = %q
# Space after '#' before the directive keyword: "#include <a>" vs
# "#include<a>".
directive_spacing = %t

# Pad each enum entry's '=' to a common column, aligning the values.
align_enum_fields = %t
# Pad the '=' of a single-declarator, initialized variable declaration
# ("new x = 1;") to align with its immediate neighbors' '=' — but only
# within one contiguous run of such declarations at the same block level; a
# blank line, a leading comment, or a non-matching statement in between
# starts a new (unaligned, unless it's also a run of 2+) group.
align_consecutive_declarations = %t
# Pad a "#define NAME value" macro's value to align with its immediate
# neighbors' values — same contiguous-run rule as
# align_consecutive_declarations (a blank line, leading comment, or
# non-matching directive breaks the run).
align_consecutive_macros = %t
# Pad a trailing end-of-line "// comment" to align with its immediate
# neighbors' comments — same contiguous-run rule as the other align_*
# options. Only applies to statements/directives that render on a single
# line; a run is broken by a blank line or an item with no trailing
# comment of its own.
align_trailing_comments = %t
# The comma after an enum body's last entry: "preserve" or "always".
enum_trailing_comma = %q
# Spacing around a declaration tag prefix's ':': "tight" removes space
# before it and inserts one after ("Float: x"); "preserve" keeps source spacing.
tag_colon_spacing = %q

# How a function's parameter list wraps once it doesn't fit line_width:
# "auto" (break only when needed, one item per line), "one_per_line"
# (always one item per line), or "bin_pack" (greedily fill each line).
multiline_function_params = %q
# multiline_function_params's counterpart for a call's argument list.
multiline_call_args = %q

# Where a binary operator lands when its expression must wrap across
# lines: "after" (operator ends the line before the wrapped operand,
# "x +\n    y") or "before" (operator leads the next line, "x\n    + y").
break_binary_operator = %q

# The comma after the last item of an array literal, call argument list, or
# parameter list, when that list is wrapped across multiple lines (never
# added when the whole list fits on one line): "never" or "multiline".
#
# A "magic trailing comma" already present in the source (one before the
# closing delimiter) is treated as an explicit request to keep that list
# exploded one item per line, even if it would otherwise fit on one line —
# matching Black/Prettier's convention of the same name. Only honored when
# this is "multiline"; with "never" the comma is simply stripped and the
# list collapses normally, since honoring the explosion while also removing
# the signal that caused it would not be a stable, idempotent choice.
trailing_comma = %q

# Format code inside a "// pawnfmt off" / "// pawnfmt on" region anyway,
# instead of leaving it untouched.
format_disabled_regions = %t
# Force one blank line after the last #include/#tryinclude at the top of
# a file.
blank_lines_after_include_block = %t
# Force one blank line between adjacent "public" function definitions,
# even if empty_line_between_top_level_declarations is false.
blank_lines_between_publics = %t
# Stably sort each contiguous run of top-level #include/#tryinclude lines
# by path.
sort_includes = %t
# Within a sorted run, place "#include <a>" (angle-bracket) lines before
# "#include \"b\"" (quoted) lines. Only takes effect when sort_includes is
# true.
group_includes_by_brackets = %t

# Cap consecutive blank lines at max_blank_lines. When false, every blank
# line run the source had is preserved as-is.
collapse_blank_lines = %t
# The cap collapse_blank_lines enforces. 0 means never preserve blank
# lines at all.
max_blank_lines = %d

# Glob patterns (matched against both the base name and full path) for
# directory formatting. exclude always wins; when include is non-empty,
# only matching files are formatted. Not used by --stdin or single-file
# invocations.
include = []
exclude = []
`,
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
		d.SpaceAfterComma,
		d.SpaceInsideParens,
		d.SpaceInsideBrackets,
		d.SpaceInsideBraces,
		d.SpaceBeforeFunctionParen,
		d.SpaceBeforeArrayBrackets,
		string(d.Semicolons),
		string(d.SingleStatementBraces),
		string(d.DirectiveIndent),
		d.DirectiveSpacing,
		d.AlignEnumFields,
		d.AlignConsecutiveDeclarations,
		d.AlignConsecutiveMacros,
		d.AlignTrailingComments,
		string(d.EnumTrailingComma),
		string(d.TagColonSpacing),
		string(d.MultilineFunctionParams),
		string(d.MultilineCallArgs),
		string(d.BreakBinaryOperator),
		string(d.TrailingComma),
		d.FormatDisabledRegions,
		d.BlankLinesAfterIncludeBlock,
		d.BlankLinesBetweenPublics,
		d.SortIncludes,
		d.GroupIncludesByBrackets,
		d.CollapseBlankLines,
		d.MaxBlankLines,
	)
}
