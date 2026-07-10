package format_test

import (
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestSharedBraceConditionalFormatting(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		stockFuncOpen,
		"#if defined A",
		"\tif(x) {",
		elseDirective,
		"\tif(y) {",
		endifDirective,
		"\t\treturn 1;",
		"\t}",
		"}",
		"",
	}, "\n"))
	want := strings.Join([]string{
		stockFuncSig,
		"{",
		"    #if defined A",
		"        if (x)",
		"        {",
		elseDirectiveIndented,
		"        if (y)",
		"        {",
		endifDirectiveIndented,
		"            return 1;",
		"        }",
		"}",
		"",
	}, "\n")

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("shared conditional formatting mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared conditional formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedAllmanSplitsOtherControlOpeners(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nwhile(first) {\n#else\nfor(new i; i < 2; i++) {\n#endif\ndo {\nCall();\n} while(false);\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, joined := range []string{"while (first) {", "for (new i; i < 2; i++) {", "do {"} {
		if strings.Contains(text, joined) {
			t.Fatalf("shared Allman output retained joined opener %q:\n%s", joined, text)
		}
	}

	for _, split := range []string{"while (first)\n", "for (new i; i < 2; i++)\n", "do\n"} {
		if !strings.Contains(text, split) {
			t.Fatalf("shared Allman output missing split opener %q:\n%s", split, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared Allman control openers are not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestConditionalFunctionHeadersFormatting(t *testing.T) {
	t.Parallel()

	source := []byte("#if defined LONG\npublic F(value,extra)\n#else\npublic F(value)\n#endif\n{ return value; }\n")
	want := strings.Join([]string{
		"#if defined LONG",
		"    public F(value, extra)",
		elseDirective,
		"    public F(value)",
		endifDirective,
		"{",
		"    return value;",
		"}",
		"",
	}, "\n")

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("conditional function formatting mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatal("conditional function formatting is not idempotent")
	}

	oneTBS := config.Default()
	oneTBS.BraceStyle = config.BraceStyle1TBS

	oneTBSFormatted := mustFormat(t, source, oneTBS)
	if strings.Contains(string(oneTBSFormatted), "#endif {") || !strings.Contains(string(oneTBSFormatted), "#endif\n{") {
		t.Fatalf("1TBS must not place a shared body brace on a directive line:\n%s", oneTBSFormatted)
	}

	whitesmiths := config.Default()
	whitesmiths.BraceStyle = config.BraceStyleWhitesmiths

	whitesmithsFormatted := mustFormat(t, source, whitesmiths)
	if !strings.Contains(string(whitesmithsFormatted), "#endif\n    {") {
		t.Fatalf("Whitesmiths must indent a conditional function body brace:\n%s", whitesmithsFormatted)
	}
}

func TestCommentAndDeclaratorReadability(t *testing.T) {
	t.Parallel()

	source := []byte("//Header\nnew first[] = {1, 2}, second[] = {3, 4}; //0 values\n")
	cfg := config.Default()
	cfg.LineWidth = 35
	want := strings.Join([]string{
		"// Header",
		"new first[] = {1, 2},",
		"    second[] = {3, 4}; // 0 values",
		"",
	}, "\n")

	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("comment/declarator readability mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestSharedConditionalTokenAwareWrapping(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		stockFuncOpen,
		"#if A",
		"\tif(very_long_condition && another_long_condition) {",
		elseDirective,
		"\tif(other_long_condition && final_long_condition) {",
		endifDirective,
		"\t\tCall(Float:value, first_argument, second_argument); //note",
		"\t}",
		"}",
		"",
	}, "\n"))
	requireSharedConditionalPath(t, source)

	cfg := config.Default()
	cfg.LineWidth = 50
	formatted := mustFormat(t, source, cfg)

	text := string(formatted)
	for _, want := range []string{"if (very_long_condition", "&& another_long_condition", "// note"} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared conditional output missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != text {
		t.Fatalf("shared token-aware wrapping is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedWrappingCountsTabVisualWidth(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(very_long_condition && another_long_condition && final_condition) {\n#else\nif(other_long_condition && another_long_condition && final_condition) {\n#endif\nreturn 1;\n}\n}\n")
	requireSharedConditionalPath(t, source)

	cfg := config.Default()
	cfg.IndentStyle = config.IndentStyleTab
	cfg.IndentWidth = 4
	cfg.LineWidth = 44

	formatted := mustFormat(t, source, cfg)
	for line := range strings.SplitSeq(string(formatted), "\n") {
		columns := 0

		for _, ch := range line {
			if ch == '\t' {
				columns += cfg.IndentWidth
			} else {
				columns++
			}
		}

		if columns > cfg.LineWidth && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			t.Fatalf("tab-indented shared line exceeds visual width (%d > %d): %q\n%s", columns, cfg.LineWidth, line, formatted)
		}
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("tab-width-aware shared wrapping is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestWrappingCountsUnicodeCharactersInsteadOfBytes(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(first) {\nCall(\"éééééééé\", value);\n#else\nif(second) {\n#endif\nreturn 1;\n}\n}\n")
	requireSharedConditionalPath(t, source)

	cfg := config.Default()
	cfg.LineWidth = 40

	formatted := mustFormat(t, source, cfg)
	if !strings.Contains(string(formatted), "Call(\"éééééééé\", value);") {
		t.Fatalf("Unicode text wrapped according to UTF-8 byte length instead of characters:\n%s", formatted)
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("Unicode-aware wrapping is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedConditionalRespectsSpacingOptions(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(Float :value, other) {\n#else\nif(bool :value, other) {\n#endif\nreturn 1;\n}\n}\n")
	requireSharedConditionalPath(t, source)

	tightCfg := config.Default()
	tightCfg.TagColonSpacing = config.TagColonSpacingTight

	tight := mustFormat(t, source, tightCfg)
	if !strings.Contains(string(tight), "if (Float: value, other)") {
		t.Fatalf("shared conditional did not apply tight tag/comma spacing:\n%s", tight)
	}

	cfg := config.Default()
	cfg.TagColonSpacing = config.TagColonSpacingPreserve
	cfg.SpaceAfterComma = false

	preserved := mustFormat(t, source, cfg)
	if !strings.Contains(string(preserved), "if (Float :value,other)") {
		t.Fatalf("shared conditional did not preserve tag spacing/remove comma spacing:\n%s", preserved)
	}
}

func TestSharedConditionalDistinguishesPrefixAndBinaryOperators(t *testing.T) {
	t.Parallel()

	source := []byte("#if A\nstock F(& Float:value, & other)\n#else\nstock F(& Float:value, & other)\n#endif\n{ return value == - 1 && ! other ? value & other : + other; }\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{"&Float:value", "&other", "value == -1", "!other", "value & other", ": +other"} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared conditional operator formatting missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared conditional operator formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedConditionalPreservesWrappedBinaryOperatorSpacing(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(very_long_value - another_long_value > limit) {\n#else\nif(other_long_value - another_long_value > limit) {\n#endif\nreturn 1;\n}\n}\n")
	requireSharedConditionalPath(t, source)

	cfg := config.Default()
	cfg.LineWidth = 42
	formatted := mustFormat(t, source, cfg)

	text := string(formatted)
	if strings.Contains(text, "-another_long_value") || !strings.Contains(text, "- another_long_value") {
		t.Fatalf("wrapped binary operator lost its spacing:\n%s", text)
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != text {
		t.Fatalf("wrapped binary operator formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedConditionalIndentsExistingContinuationLines(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(first_condition &&\nsecond_condition &&\nthird_condition) {\n#else\nif(other_condition ||\nfinal_condition) {\n#endif\nCall(first_argument,\nsecond_argument);\n}\n}\n")
	requireSharedConditionalPath(t, source)
	formatted := mustFormat(t, source, config.Default())
	text := string(formatted)

	want := "    if (first_condition &&\n        second_condition &&\n        third_condition)"
	if !strings.Contains(text, want) {
		t.Fatalf("shared continuation indentation missing %q:\n%s", want, text)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared continuation indentation is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedConditionalAppliesAlwaysBracesToCompleteInlineControls(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(first) return 1;\nelse Call(first);\n#else\nif(second) return 2;\nelse if(third) Call(second);\n#endif\nreturn 0;\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{
		"if (first)\n    {\n        return 1;\n    }",
		"else\n    {\n        Call(first);\n    }",
		"if (second)\n    {\n        return 2;\n    }",
		"else if (third)\n    {\n        Call(second);\n    }",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared inline control did not receive default braces, missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared inline control braces are not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}

	preserve := config.Default()
	preserve.SingleStatementBraces = config.SingleStatementBracesPreserve

	preserved := mustFormat(t, source, preserve)
	if !strings.Contains(string(preserved), "if (first) return 1;") ||
		!strings.Contains(string(preserved), "else Call(first);") {
		t.Fatalf("shared inline control ignored braces=preserve:\n%s", preserved)
	}
}

func TestSharedInlineControlRespectsMultilineWithoutBraceSynthesis(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(first) Call(first);\nelse Call(second);\n#else\nif(third) Call(third);\n#endif\n}\n")

	for _, braces := range []config.SingleStatementBraces{
		config.SingleStatementBracesPreserve,
		config.SingleStatementBracesNever,
	} {
		cfg := config.Default()
		cfg.SingleStatementBraces = braces
		cfg.KeepSimpleStatementsSingleLine = false
		formatted := mustFormat(t, source, cfg)

		text := string(formatted)
		if strings.Contains(text, "if (first) Call(first);") || strings.Contains(text, "else Call(second);") ||
			!strings.Contains(text, "if (first)\n        Call(first);") ||
			!strings.Contains(text, "else\n        Call(second);") {
			t.Fatalf("shared inline control ignored multiline policy with braces=%s:\n%s", braces, text)
		}

		second := mustFormat(t, formatted, cfg)
		if string(second) != text {
			t.Fatalf("shared multiline unbraced control is not idempotent with braces=%s\nfirst:\n%s\nsecond:\n%s", braces, text, second)
		}
	}
}

func TestSharedInlineControlBracesConvergeAcrossStyles(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(first) Call(first);\nelse Call(second);\n#else\nif(third) Call(third);\n#endif\n}\n")

	cases := []struct {
		name   string
		mutate func(*config.Config)
	}{
		{name: "1tbs", mutate: func(c *config.Config) { c.BraceStyle = config.BraceStyle1TBS }},
		{name: "whitesmiths", mutate: func(c *config.Config) { c.BraceStyle = config.BraceStyleWhitesmiths }},
		{name: "tabs", mutate: func(c *config.Config) { c.IndentStyle = config.IndentStyleTab }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Default()
			tc.mutate(&cfg)
			first := mustFormat(t, source, cfg)

			second := mustFormat(t, first, cfg)
			if string(second) != string(first) {
				t.Fatalf("shared inline braces did not converge\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}

func TestSharedDirectiveContinuationsUseStableIndent(t *testing.T) {
	t.Parallel()

	source := []byte("stock F() {\n#if A\nif(first) {\n#if FIRST\\\n|| SECOND\\\n|| THIRD\nCall();\n#endif\n#else\nif(second) {\n#endif\nCall();\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())
	text := string(formatted)
	continuationIndents := []int{}

	for line := range strings.SplitSeq(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "|| SECOND\\" || trimmed == "|| THIRD" {
			continuationIndents = append(continuationIndents, len(line)-len(strings.TrimLeft(line, " \t")))
		}
	}

	if len(continuationIndents) != 2 || continuationIndents[0] != continuationIndents[1] {
		t.Fatalf("directive continuations did not use one stable indent (%v):\n%s", continuationIndents, text)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("directive continuation indentation is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedConditionalSynthesizesSoleSharedOpenBrace(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		stockFuncOpen,
		"#if defined isnull",
		"if (isnull(s))",
		elseDirective,
		"if (s[0] == 0)",
		endifDirective,
		"{",
		returnOneStatement,
		"}",
		"}",
		"",
	}, "\n"))
	requireSharedConditionalPath(t, source)
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	if !strings.Contains(text, "    {\n") || !strings.Contains(text, "    }\n") {
		t.Fatalf("expected a synthesized, correctly-indented brace pair:\n%s", text)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("synthesized brace formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedConditionalRendersTrailingElse(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		stockFuncOpen,
		"#if defined isnull",
		"if (isnull(s))",
		elseDirective,
		"if (s[0] == 0)",
		endifDirective,
		"{",
		returnOneStatement,
		"}",
		"else return 0;",
		"}",
		"",
	}, "\n"))
	requireSharedConditionalPath(t, source)
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	if !strings.Contains(text, "else") || !strings.Contains(text, "return 0;") {
		t.Fatalf("trailing else after a shared conditional was dropped:\n%s", text)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("trailing-else formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestIfStatementRendersConditionalElseIfExtension(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		"stock F(a) {",
		"if (a == 1)",
		"{",
		returnOneStatement,
		"}",
		"#if defined B",
		"else if (a == 2)",
		"{",
		"return 2;",
		"}",
		endifDirective,
		"else return 0;",
		"}",
		"",
	}, "\n"))
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{"defined B", "else if (a == 2)", "return 2;", "return 0;"} {
		if !strings.Contains(text, want) {
			t.Fatalf("conditional else-if extension missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("conditional else-if extension formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestConditionalRegionRendersSharedTrailingElseOnce(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		"stock F(lc) {",
		"#if defined A",
		"if(lc == 1) lc = 10;",
		elseDirective,
		"if(lc == 2) lc = 10;",
		endifDirective,
		"else if(s[0])",
		"{",
		"trim(s);",
		"}",
		"}",
		"",
	}, "\n"))
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	if strings.Count(text, "else if (s[0])") != 1 {
		t.Fatalf("shared trailing else was duplicated once per branch instead of rendered once:\n%s", text)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared trailing else formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedConditionalSynthesizesBraceDespiteBalancedNestedBlocks(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		"stock F(lid, newkeys) {",
		"#if defined A",
		"new clab;",
		"if (newkeys & KEY_WALK)",
		"{",
		"if (x)",
		"clab = 1;",
		"else",
		"clab = 2;",
		"}",
		"else clab = 3;",
		"if (clab != -1)",
		elseDirective,
		"new clab = 4;",
		"if (clab != -2)",
		endifDirective,
		"{",
		returnOneStatement,
		"}",
		"else return 0;",
		"}",
		"",
	}, "\n"))
	requireSharedConditionalPath(t, source)
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	if !strings.Contains(text, returnOneStatement) {
		t.Fatalf("shared body's opening brace was dropped despite balanced nested blocks in the prefix:\n%s", text)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedInlineControlSplitsWithoutBracesWhenConfigured(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.SingleStatementBraces = config.SingleStatementBracesNever
	cfg.KeepSimpleStatementsSingleLine = false
	source := []byte(strings.Join([]string{
		stockFuncOpen,
		"#if A",
		"if(first) {",
		elseDirective,
		"if(second) Call(second);",
		endifDirective,
		"Common();",
		"}",
		"}",
		"",
	}, "\n"))
	want := strings.Join([]string{
		stockFuncSig,
		"{",
		"    #if A",
		"    if (first)",
		openBraceIndented,
		elseDirectiveIndented,
		"    if (second)",
		"        Call(second);",
		endifDirectiveIndented,
		"        Common();",
		closingBraceIndented,
		"}",
		"",
	}, "\n")

	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("shared inline control did not split without braces\nexpected:\n%s\nactual:\n%s", want, formatted)
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("split inline control is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedControlStatementWrapsConditionAndBodySeparately(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.SingleStatementBraces = config.SingleStatementBracesPreserve
	cfg.LineWidth = 60
	source := []byte(strings.Join([]string{
		stockFuncOpen,
		"#if A",
		"if(first) {",
		elseDirective,
		"if(firstConditionIsVeryLong && secondConditionIsAlsoVeryLong) CallSomeFunctionHere(argumentOne, argumentTwo);",
		endifDirective,
		"Common();",
		"}",
		"}",
		"",
	}, "\n"))
	want := strings.Join([]string{
		stockFuncSig,
		"{",
		"    #if A",
		"    if (first)",
		openBraceIndented,
		elseDirectiveIndented,
		"    if (firstConditionIsVeryLong",
		"        && secondConditionIsAlsoVeryLong)",
		"        CallSomeFunctionHere(argumentOne, argumentTwo);",
		endifDirectiveIndented,
		"        Common();",
		closingBraceIndented,
		"}",
		"",
	}, "\n")

	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("shared control statement did not wrap condition and body separately\nexpected:\n%s\nactual:\n%s", want, formatted)
	}

	for line := range strings.SplitSeq(string(formatted), "\n") {
		if len(line) > cfg.LineWidth {
			t.Fatalf("a line exceeds LineWidth despite needing to wrap:\n%s", formatted)
		}
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("wrapped control statement is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}
