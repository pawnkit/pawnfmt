package format_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func TestNarrowWidthMultilineLists(t *testing.T) {
	t.Parallel()

	source := []byte("forward VeryLongCallbackName(playerid, const playerName[], Float:spawnX, Float:spawnY, Float:spawnZ);\n")
	cfg := config.Default()
	cfg.LineWidth = 40

	formatted := mustFormat(t, source, cfg)
	if !strings.Contains(string(formatted), "\n") || strings.Count(string(formatted), "\n") < 3 {
		t.Fatalf("expected multiline output for narrow width, got:\n%s", formatted)
	}

	second := mustFormat(t, formatted, cfg)
	if string(formatted) != string(second) {
		t.Fatalf("narrow-width multiline formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestTopLevelSemanticGrouping(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		"#define FEATURE 1",
		"new first;",
		"new second;",
		"forward OnFirst();",
		"forward OnSecond();",
		"stock First() { return 1; }",
		"stock Second() { return 2; }",
	}, "\n"))
	want := strings.Join([]string{
		"#define FEATURE 1",
		"",
		"new first;",
		"new second;",
		"",
		"forward OnFirst();",
		"forward OnSecond();",
		"",
		"stock First()",
		"{",
		"    return 1;",
		"}",
		"",
		"stock Second()",
		"{",
		"    return 2;",
		"}",
		"",
	}, "\n")

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("semantic top-level grouping mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestElseIfRemainsChainedWhenSimpleStatementsBreak(t *testing.T) {
	t.Parallel()

	source := []byte("stock F(x) { if (x == 1) return 1; else if (x == 2) return 2; else return 0; }\n")
	cfg := config.Default()
	cfg.KeepSimpleStatementsSingleLine = false
	cfg.SingleStatementBraces = config.SingleStatementBracesPreserve

	formatted := mustFormat(t, source, cfg)
	if !strings.Contains(string(formatted), "else if (x == 2)") || strings.Contains(string(formatted), "else\n        if") {
		t.Fatalf("else-if chain was dismantled:\n%s", formatted)
	}
}

func TestNeverBracesPreservesDanglingElseSemantics(t *testing.T) {
	t.Parallel()

	source := []byte("stock F(a, b) { if (a) { if (b) return 1; } else return 2; return 0; }\n")
	cfg := config.Default()
	cfg.SingleStatementBraces = config.SingleStatementBracesNever

	formatted := mustFormat(t, source, cfg)
	if !strings.Contains(string(formatted), "if (a)\n    {") {
		t.Fatalf("outer consequence braces were removed and would rebind else:\n%s", formatted)
	}
}

func TestSortIncludesKeepsFileHeaderAtTop(t *testing.T) {
	t.Parallel()

	source := []byte("// file header\n#include <zeta>\n// alpha note\n#include <alpha>\n#include <middle>\n")
	cfg := config.Default()
	cfg.SortIncludes = true
	formatted := mustFormat(t, source, cfg)

	want := "// file header\n// alpha note\n#include <alpha>\n#include <middle>\n#include <zeta>\n"
	if string(formatted) != want {
		t.Fatalf("include sorting moved the file header or detached an include comment\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestDisabledRegionIsNotSortedOrSeparated(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		"// pawnfmt off",
		"#include <zeta>",
		"#include <alpha>",
		"new   first;",
		"new   second;",
		"// pawnfmt on",
		"",
	}, "\n"))
	cfg := config.Default()
	cfg.SortIncludes = true

	formatted := mustFormat(t, source, cfg)
	if string(formatted) != string(source) {
		t.Fatalf("disabled region was reordered, separated, or formatted\nexpected:\n%s\nactual:\n%s", source, formatted)
	}
}

func TestGroupIncludesByBracketsPutsAngleBracketsFirst(t *testing.T) {
	t.Parallel()

	source := []byte("#include \"local.inc\"\n#include <a_samp>\n")
	cfg := config.Default()
	cfg.SortIncludes = true
	cfg.GroupIncludesByBrackets = true
	formatted := mustFormat(t, source, cfg)

	want := "#include <a_samp>\n#include \"local.inc\"\n"
	if string(formatted) != want {
		t.Fatalf("group_includes_by_brackets did not put <> includes before \"\" includes\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestAlignConsecutiveMacrosPadsValues(t *testing.T) {
	t.Parallel()

	source := []byte("#define SHORT 1\n#define MUCH_LONGER 2\n")
	cfg := config.Default()
	cfg.AlignConsecutiveMacros = true
	formatted := mustFormat(t, source, cfg)

	want := "#define SHORT       1\n#define MUCH_LONGER 2\n"
	if string(formatted) != want {
		t.Fatalf("align_consecutive_macros did not pad values to a common column\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestAlignConsecutiveMacrosBreaksRunOnBlankLine(t *testing.T) {
	t.Parallel()

	source := []byte("#define SHORT 1\n\n#define MUCH_LONGER 2\n")
	cfg := config.Default()
	cfg.AlignConsecutiveMacros = true
	formatted := mustFormat(t, source, cfg)

	want := "#define SHORT 1\n\n#define MUCH_LONGER 2\n"
	if string(formatted) != want {
		t.Fatalf("a blank line should break an align_consecutive_macros run\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestAlignTrailingCommentsPadsComments(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n    new x = 1; // a\n    new muchLongerName = 2; // b\n}\n")
	cfg := config.Default()
	cfg.AlignTrailingComments = true
	formatted := mustFormat(t, source, cfg)

	want := "stock F()\n{\n    new x = 1;              // a\n    new muchLongerName = 2; // b\n}\n"
	if string(formatted) != want {
		t.Fatalf("align_trailing_comments did not pad comments to a common column\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestAlignTrailingCommentsSkipsItemsWithoutTrailingComments(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n    new x = 1; // a\n    new muchLongerName = 2;\n    new y = 1; // c\n}\n")
	cfg := config.Default()
	cfg.AlignTrailingComments = true
	formatted := mustFormat(t, source, cfg)
	// muchLongerName has no trailing comment, so it splits the run: "x" and
	// "y" are not adjacent and must not be aligned with each other.
	if strings.Contains(string(formatted), "1;              // a") || strings.Contains(string(formatted), "1;              // c") {
		t.Fatalf("align_trailing_comments aligned across a non-adjacent item with no comment of its own:\n%s", formatted)
	}
}

func TestBreakBinaryOperatorBeforePlacesOperatorOnContinuationLine(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n    return aaaaaaaaaa + bbbbbbbbbb + cccccccccc;\n}\n")
	cfg := config.Default()
	cfg.LineWidth = 30
	cfg.BreakBinaryOperator = config.BinaryOperatorBreakBefore

	formatted := mustFormat(t, source, cfg)
	if !strings.Contains(string(formatted), "\n        + ") {
		t.Fatalf("break_binary_operator=before should place '+' leading a continuation line:\n%s", formatted)
	}

	if strings.Contains(string(formatted), "+\n") {
		t.Fatalf("break_binary_operator=before should not leave '+' trailing a line:\n%s", formatted)
	}
}

func TestIndentCaseLabelsFalseKeepsLabelsFlushWithSwitch(t *testing.T) {
	t.Parallel()

	source := []byte("stock F(x) {\n    switch (x) {\n        case 1: return 1;\n        default: return 0;\n    }\n}\n")
	cfg := config.Default()
	cfg.IndentCaseLabels = false
	formatted := mustFormat(t, source, cfg)

	want := "stock F(x)\n{\n    switch (x)\n    {\n    case 1:\n        return 1;\n    default:\n        return 0;\n    }\n}\n"
	if string(formatted) != want {
		t.Fatalf("indent_case_labels=false should keep case/default flush with switch's own brace\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestIndentGotoLabelsFalseOutdentsLabel(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n    goto Skip;\n    new x = 1;\n    Skip:\n    return x;\n}\n")
	cfg := config.Default()
	cfg.IndentGotoLabels = false

	formatted := mustFormat(t, source, cfg)
	if !strings.Contains(string(formatted), "\nSkip:\n") {
		t.Fatalf("indent_goto_labels=false should outdent the label to column 0:\n%s", formatted)
	}
}

func TestOperatorSpacingAppliesToDeclarations(t *testing.T) {
	t.Parallel()

	source := []byte("new value = 1;\nforward F(arg = 2);\nnative N() = Alias;\nenum E { Field = 3 };\n")
	cfg := config.Default()
	cfg.SpaceAroundOperators = false

	formatted := mustFormat(t, source, cfg)
	for _, want := range []string{"value=1", "arg=2", "N()=Alias", "Field=3"} {
		if !strings.Contains(string(formatted), want) {
			t.Fatalf("compact operator spacing missing %q:\n%s", want, formatted)
		}
	}
}

func TestCommaSpacingAppliesToArrayAndCaseLists(t *testing.T) {
	t.Parallel()

	source := []byte("new values[] = {1, 2, 3};\nstock F(x) { switch (x) { case 1, 2, 3: return 1; } }\n")
	cfg := config.Default()
	cfg.SpaceAfterComma = false

	formatted := mustFormat(t, source, cfg)
	for _, want := range []string{"{1,2,3}", "case 1,2,3:"} {
		if !strings.Contains(string(formatted), want) {
			t.Fatalf("compact comma spacing missing %q:\n%s", want, formatted)
		}
	}
}

func TestDirectiveIndentNone(t *testing.T) {
	t.Parallel()
	source := readFile(t, filepath.Join(testdataDir(), "input", "wrapped_inline_labels.pwn"))
	cfg := config.Default()
	cfg.DirectiveIndent = config.DirectiveIndentNone
	formatted := mustFormat(t, source, cfg)

	expected := ensureTrailingNewline([]byte(strings.Join([]string{
		"stock Wrapped(value)",
		"{",
		"#if defined FEATURE",
		"    label_a: return 1;",
		"#elseif defined ALT",
		"    while (value > 0)",
		"    {",
		"#if defined INNER",
		"        retry: continue;",
		"#endif",
		closingBraceIndented,
		elseDirective,
		"    guard: #warning fallback branch",
		"#endif",
		"}",
	}, "\n")))
	if string(formatted) != string(expected) {
		t.Fatalf("directive_indent=none mismatch\nexpected:\n%s\nactual:\n%s", expected, formatted)
	}
}

func TestDirectiveIndentNoneAcrossNestedContainers(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		"stock F(value) {",
		"    switch (value) {",
		"        case 1: return 1;",
		"#if FEATURE",
		"        case 2: return 2;",
		"#endif",
		closingBraceIndented,
		"#if OUTER",
		"    if (value) {",
		elseDirective,
		"    if (!value) {",
		"#endif",
		"        return 3;",
		closingBraceIndented,
		"}",
		"",
	}, "\n"))
	cfg := config.Default()
	cfg.DirectiveIndent = config.DirectiveIndentNone

	formatted := mustFormat(t, source, cfg)
	for line := range strings.SplitSeq(string(formatted), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.HasPrefix(line, " ") {
			t.Fatalf("directive was not reset to column zero: %q\n%s", line, formatted)
		}
	}
}

func TestDirectiveIndentKeepInBlock(t *testing.T) {
	t.Parallel()
	source := readFile(t, filepath.Join(testdataDir(), "input", "wrapped_inline_labels.pwn"))
	cfg := config.Default()
	cfg.DirectiveIndent = config.DirectiveIndentKeepInBlock
	formatted := mustFormat(t, source, cfg)

	expected := ensureTrailingNewline(readFile(t, filepath.Join(testdataDir(), "expected", "wrapped_inline_labels.pwn")))
	if string(formatted) != string(expected) {
		t.Fatalf("directive_indent=keep_in_block mismatch\nexpected:\n%s\nactual:\n%s", expected, formatted)
	}
}

// configOptionCase is one case in TestConfigOptionsChangeOutput's table:
// formatting source with reference (or the default config, if nil) must
// differ from formatting it with mutate applied.
type configOptionCase struct {
	name      string
	source    string
	reference func(*config.Config)
	mutate    func(*config.Config)
}

func configOptionCases() []configOptionCase {
	return []configOptionCase{
		{name: "indent_style_tab", source: "stock F() {\n\tnew x = 1;\n}\n", mutate: func(c *config.Config) { c.IndentStyle = config.IndentStyleTab }},
		{name: "newline_style_crlf", source: "new x;\nnew y;\n", mutate: func(c *config.Config) { c.NewlineStyle = config.NewlineStyleCRLF }},
		{name: "insert_final_newline_false", source: "new x;\n", mutate: func(c *config.Config) { c.InsertFinalNewline = false }},
		{name: "brace_style_1tbs", source: "stock F(x) {\n    if (x) {\n        y = 1;\n    }\n}\n", mutate: func(c *config.Config) { c.BraceStyle = config.BraceStyle1TBS }},
		{name: "brace_style_whitesmiths", source: "stock F(x) {\n    if (x) {\n        y = 1;\n    }\n}\n", mutate: func(c *config.Config) { c.BraceStyle = config.BraceStyleWhitesmiths }},
		{
			name:      "keep_simple_statements_single_line_false",
			source:    "stock F(x) {\n    if (x) return 1;\n    return 0;\n}\n",
			reference: func(c *config.Config) { c.SingleStatementBraces = config.SingleStatementBracesPreserve },
			mutate: func(c *config.Config) {
				c.SingleStatementBraces = config.SingleStatementBracesPreserve
				c.KeepSimpleStatementsSingleLine = false
			},
		},
		{name: "indent_case_contents_false", source: "stock F(x) {\n    switch (x) {\n        case 1: return 1;\n        default: return 0;\n    }\n}\n", mutate: func(c *config.Config) { c.IndentCaseContents = false }},
		{name: "empty_line_between_top_level_decls_false", source: "stock A() {\n    return 1;\n}\nstock B() {\n    return 2;\n}\n", mutate: func(c *config.Config) { c.EmptyLineBetweenTopLevelDecls = false }},
		{name: "space_around_operators_false", source: "stock F() {\n    new y = 1;\n    y = y + 1;\n    return y;\n}\n", mutate: func(c *config.Config) { c.SpaceAroundOperators = false }},
		{name: "space_after_comma_false", source: "stock F(a, b) {\n    return a + b;\n}\n", mutate: func(c *config.Config) { c.SpaceAfterComma = false }},
		{name: "space_inside_parens_true", source: "stock F(x) {\n    if (x) return 1;\n    return 0;\n}\n", mutate: func(c *config.Config) { c.SpaceInsideParens = true }},
		{name: "space_inside_brackets_true", source: "new a[4];\nstock F() {\n    return a[0];\n}\n", mutate: func(c *config.Config) { c.SpaceInsideBrackets = true }},
		{name: "space_before_function_paren_true", source: "stock F(x) {\n    return x;\n}\n", mutate: func(c *config.Config) { c.SpaceBeforeFunctionParen = true }},
		{name: "semicolons_always", source: "enum X {\n    A\n}\n", mutate: func(c *config.Config) { c.Semicolons = config.SemicolonsAlways }},
		{
			name:      "single_statement_braces_always",
			source:    "stock F(x) {\n    if (x) return 1;\n}\n",
			reference: func(c *config.Config) { c.SingleStatementBraces = config.SingleStatementBracesPreserve },
			mutate:    func(c *config.Config) { c.SingleStatementBraces = config.SingleStatementBracesAlways },
		},
		{name: "single_statement_braces_never", source: "stock F(x) {\n    if (x) {\n        return 1;\n    }\n}\n", mutate: func(c *config.Config) { c.SingleStatementBraces = config.SingleStatementBracesNever }},
		{name: "directive_indent_none", source: "stock F(x) {\n    if (x) {\n        #if defined X\n        return 1;\n        #endif\n    }\n    return 0;\n}\n", mutate: func(c *config.Config) { c.DirectiveIndent = config.DirectiveIndentNone }},
		{name: "directive_spacing_false", source: "#include <a>\n", mutate: func(c *config.Config) { c.DirectiveSpacing = false }},
		{name: "align_enum_fields_true", source: "enum X {\n    SHORT = 1,\n    MUCH_LONGER = 2\n};\n", mutate: func(c *config.Config) { c.AlignEnumFields = true }},
		{name: "align_consecutive_declarations_true", source: "new gShort = 1;\nnew gVeryLongName = 2;\n", mutate: func(c *config.Config) { c.AlignConsecutiveDeclarations = true }},
		{
			name:      "enum_trailing_comma_always",
			source:    "enum X {\n    A,\n    B\n};\n",
			reference: func(c *config.Config) { c.EnumTrailingComma = config.EnumTrailingCommaPreserve },
			mutate:    func(c *config.Config) { c.EnumTrailingComma = config.EnumTrailingCommaAlways },
		},
		{name: "tag_colon_spacing_preserve", source: "new Float : x;\n", mutate: func(c *config.Config) { c.TagColonSpacing = config.TagColonSpacingPreserve }},
		{name: "tag_colon_spacing_compact", source: "new Float: x;\n", reference: func(c *config.Config) { c.TagColonSpacing = config.TagColonSpacingTight }, mutate: func(c *config.Config) { c.TagColonSpacing = config.TagColonSpacingCompact }},
		{name: "space_before_array_brackets_true", source: "new x[4];\n", mutate: func(c *config.Config) { c.SpaceBeforeArrayBrackets = true }},
		{
			name:   "multiline_function_params_bin_pack",
			source: "forward Foo(aaaaaaaaaa, bbbbbbbbbb, cccccccccc, dddddddddd, eeeeeeeeee);\n",
			mutate: func(c *config.Config) { c.LineWidth = 40; c.MultilineFunctionParams = config.MultilineListBinPack },
		},
		{
			name:   "multiline_call_args_bin_pack",
			source: "stock F() {\n    Call(aaaaaaaaaa, bbbbbbbbbb, cccccccccc, dddddddddd, eeeeeeeeee);\n}\n",
			mutate: func(c *config.Config) { c.LineWidth = 40; c.MultilineCallArgs = config.MultilineListBinPack },
		},
		{
			name:   "multiline_function_params_one_per_line",
			source: "forward Foo(a, b);\n",
			mutate: func(c *config.Config) { c.MultilineFunctionParams = config.MultilineListOnePerLine },
		},
		{
			name:   "multiline_call_args_one_per_line",
			source: "stock F() {\n    Call(a, b);\n}\n",
			mutate: func(c *config.Config) { c.MultilineCallArgs = config.MultilineListOnePerLine },
		},
		{name: "format_disabled_regions_true", source: "// pawnfmt off\nnew   x = 1;\n// pawnfmt on\n", mutate: func(c *config.Config) { c.FormatDisabledRegions = true }},
		{name: "blank_lines_after_include_block_false", source: "#include <a>\nnew x;\n", mutate: func(c *config.Config) { c.BlankLinesAfterIncludeBlock = false }},
		{
			name:      "blank_lines_between_publics_false",
			source:    "public A() {\n    return 1;\n}\npublic B() {\n    return 2;\n}\n",
			reference: func(c *config.Config) { c.EmptyLineBetweenTopLevelDecls = false },
			mutate:    func(c *config.Config) { c.EmptyLineBetweenTopLevelDecls = false; c.BlankLinesBetweenPublics = false },
		},
		{name: "sort_includes_true", source: "#include <zeta>\n#include <alpha>\n", mutate: func(c *config.Config) { c.SortIncludes = true }},
		{name: "space_inside_braces_true", source: "new a[3] = {1, 2, 3};\n", mutate: func(c *config.Config) { c.SpaceInsideBraces = true }},
		{
			name:   "trim_trailing_whitespace_false",
			source: "/* line one   \nline two */\nnew x;\n",
			mutate: func(c *config.Config) { c.TrimTrailingWhitespace = false },
		},
		{
			name:   "collapse_blank_lines_false",
			source: "stock F() {\n    new x;\n\n\n\n\n\n\n    new y;\n}\n",
			mutate: func(c *config.Config) { c.CollapseBlankLines = false },
		},
		{
			name:   "align_consecutive_macros_true",
			source: "#define SHORT 1\n#define MUCH_LONGER 2\n",
			mutate: func(c *config.Config) { c.AlignConsecutiveMacros = true },
		},
		{
			name:   "align_trailing_comments_true",
			source: "stock F() {\n    new x = 1; // a\n    new muchLongerName = 2; // b\n}\n",
			mutate: func(c *config.Config) { c.AlignTrailingComments = true },
		},
		{
			name:   "break_binary_operator_before",
			source: "stock F() {\n    return aaaaaaaaaa + bbbbbbbbbb + cccccccccc + dddddddddd + eeeeeeeeee;\n}\n",
			mutate: func(c *config.Config) { c.LineWidth = 40; c.BreakBinaryOperator = config.BinaryOperatorBreakBefore },
		},
		{
			name:   "indent_case_labels_false",
			source: "stock F(x) {\n    switch (x) {\n        case 1: return 1;\n        default: return 0;\n    }\n}\n",
			mutate: func(c *config.Config) { c.IndentCaseLabels = false },
		},
		{
			name:   "indent_goto_labels_false",
			source: "stock F() {\n    goto Skip;\n    new x = 1;\n    Skip:\n    return x;\n}\n",
			mutate: func(c *config.Config) { c.IndentGotoLabels = false },
		},
		{
			name:      "group_includes_by_brackets_true",
			source:    "#include \"local.inc\"\n#include <a_samp>\n",
			reference: func(c *config.Config) { c.SortIncludes = true },
			mutate: func(c *config.Config) {
				c.SortIncludes = true
				c.GroupIncludesByBrackets = true
			},
		},
	}
}

func TestConfigOptionsChangeOutput(t *testing.T) {
	t.Parallel()

	for _, tc := range configOptionCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseCfg := config.Default()
			if tc.reference != nil {
				tc.reference(&baseCfg)
			}

			baseline := mustFormat(t, []byte(tc.source), baseCfg)
			cfg := config.Default()
			tc.mutate(&cfg)

			mutated := mustFormat(t, []byte(tc.source), cfg)
			if string(baseline) == string(mutated) {
				t.Fatalf("changing this option produced output identical to the reference config; the option may be unimplemented\nsource:\n%s\noutput:\n%s", tc.source, baseline)
			}
		})
	}
}

func TestWhitesmithsBracesMatch(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.BraceStyle = config.BraceStyleWhitesmiths
	source := []byte("stock F(x) {\n    if (x) {\n        return 1;\n    }\n}\n")
	formatted := mustFormat(t, source, cfg)

	lines := strings.Split(string(formatted), "\n")
	indentOf := func(line string) int {
		return len(line) - len(strings.TrimLeft(line, " "))
	}

	var openIndent, closeIndent []int

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "{" {
			openIndent = append(openIndent, indentOf(line))
		}

		if trimmed == "}" {
			closeIndent = append(closeIndent, indentOf(line))
		}
	}

	if len(openIndent) != 2 || len(closeIndent) != 2 {
		t.Fatalf("expected 2 opening and 2 closing braces on their own line, got %v / %v\noutput:\n%s", openIndent, closeIndent, formatted)
	}

	for i := range openIndent {
		if openIndent[i] != closeIndent[len(closeIndent)-1-i] {
			t.Fatalf("brace pair %d: opening indent %d != matching closing indent %d\noutput:\n%s", i, openIndent[i], closeIndent[len(closeIndent)-1-i], formatted)
		}
	}
}

func TestConditionalSplitHeaderNoBodyIsIdempotent(t *testing.T) {
	t.Parallel()

	source := []byte("stock F()\n{\n" +
		"#if defined foreach\n" +
		"foreach(new i : Player)\n" +
		"#else\n" +
		"for(new i = 0; i < MAX; i++)\n" +
		"#endif\n" +
		"{\n" +
		"if (i) continue;\n" +
		"}\n}\n")
	first := mustFormat(t, source, config.Default())

	second := mustFormat(t, first, config.Default())
	if string(first) != string(second) {
		t.Fatalf("not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	if strings.Contains(string(first), "\n\n") {
		t.Fatalf("expected no blank lines to appear around the split header, got:\n%s", first)
	}
}

func TestCRLFLineCommentIsIdempotent(t *testing.T) {
	t.Parallel()

	source := []byte("// note\r\nnew x;\r\n")
	first := mustFormat(t, source, config.Default())

	second := mustFormat(t, first, config.Default())
	if string(first) != string(second) {
		t.Fatalf("not idempotent\nfirst:\n%q\nsecond:\n%q", first, second)
	}

	if strings.Contains(string(first), "\r\r") {
		t.Fatalf("comment gained an extra \\r:\n%q", first)
	}
}

func TestBraceStyleAppliesToEveryConstruct(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{"function", "stock F() {\n    return 1;\n}\n"},
		{"enum", "enum X {\n    A,\n    B\n};\n"},
		{"switch", "stock F(x) {\n    switch (x) {\n        case 1: return 1;\n    }\n}\n"},
		{"macro_invocation_block", "stock F() {\n    each_player(i) {\n        x = i;\n    }\n}\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Default() // Allman by default

			formatted := mustFormat(t, []byte(tc.source), cfg)
			for line := range strings.SplitSeq(string(formatted), "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed == "{" || trimmed == "" || !strings.Contains(trimmed, "{") {
					continue
				}

				if trimmed == emptyBraceBody {
					continue
				}

				t.Fatalf("%s: found a non-own-line '{' under BraceStyleAllman: %q\nfull output:\n%s", tc.name, line, formatted)
			}
		})
	}
}

func TestRejectsParseInvalidInput(t *testing.T) {
	t.Parallel()

	source := []byte("}\n")

	_, err := formatter.Source(source, config.Default())
	if err == nil {
		t.Fatal("expected parse-invalid input to be rejected")
	}

	if !strings.Contains(err.Error(), "source does not parse cleanly") {
		t.Fatalf("unexpected error: %v", err)
	}
}
