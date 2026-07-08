package format_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestMacroBodyFallbackPreservesOwnBackslashes(t *testing.T) {
	source := []byte(strings.Join([]string{
		`#define CHAIN_FORWARD:%0_%2(%1)=%3; \`,
		"\tforward %0_%2(%1); \\",
		"\tpublic %0_%2(%1) <_ALS : _ALS_x0, _ALS : _ALS_x1> { return (%3); } \\",
		"\tpublic %0_%2(%1) <> { return (%3); }",
		"",
	}, "\n"))
	formatted := mustFormat(t, source, config.Default())
	if strings.Contains(string(formatted), `\ \`) {
		t.Fatalf("macro body backslash continuation was duplicated:\n%s", formatted)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("macro body formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSubscriptPreservesWildcardBraceDelimiter(t *testing.T) {
	source := []byte("stock F(playerid) { if (!items[playerid][i]{0}) Call(); }\n")
	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "items[playerid][i]{0}") {
		t.Fatalf("wildcard-tag subscript delimiter was not preserved:\n%s", formatted)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("wildcard subscript formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestVariableDeclaratorRendersCapacitySuffix(t *testing.T) {
	source := []byte("#if defined foreach\nnew\n\tIterator:FCNPC<MAX_PLAYERS>;\n#endif\n")
	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "<MAX_PLAYERS>") {
		t.Fatalf("declarator's capacity suffix was dropped:\n%s", formatted)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("capacity suffix formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestFunctionRendersCallingConventionDimension(t *testing.T) {
	source := []byte("native ArgTag:[2]pawn_arg_pack(AnyTag:value, tag_id = tagof value);\n")
	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "[2]") {
		t.Fatalf("return-type array dimension was dropped:\n%s", formatted)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("calling_convention dimension formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestConditionalSpliceGetsLineNormalization(t *testing.T) {
	source := readFile(t, filepath.Join(testdataDir(), "regressions", "conditional_splice_creator", "source.pwn"))
	requireSharedConditionalPath(t, source)
	formatted := mustFormat(t, source, config.Default())
	text := string(formatted)
	if strings.Contains(text, "\t\t\t\t\t\tif(keys & KEY_WALK && keys & KEY_JUMP) GetDynamicObjectRot") {
		t.Fatalf("conditional_splice content was passed through verbatim instead of normalized:\n%s", text)
	}
	if !strings.Contains(text, "if (keys & KEY_WALK && keys & KEY_JUMP)") {
		t.Fatalf("conditional_splice content did not get normal operator spacing:\n%s", text)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("conditional_splice formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestLongConditionDirectiveWrapsWithContinuation(t *testing.T) {
	source := []byte("stock F() {\n\t#if defined IsValidDynamicObject && defined IsDynamicObjectMaterialTextUsed && defined GetDynamicObjectMaterialText && defined SetDynamicObjectMaterialText\n\tnew x;\n\t#endif\n}\n")
	formatted := mustFormat(t, source, config.Default())
	text := string(formatted)
	for line := range strings.SplitSeq(text, "\n") {
		if len(line) > 100 {
			t.Fatalf("condition directive line exceeds LineWidth despite needing to wrap:\n%s", text)
		}
	}
	if !strings.Contains(text, "\\\n") {
		t.Fatalf("long condition directive did not wrap with \"\\\" continuation:\n%s", text)
	}
	if strings.Contains(text, "\n"+"defined") {
		t.Fatalf("continuation line lost its indentation:\n%s", text)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("condition directive wrapping is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestNestedConditionalDirectivesAlignWithEnclosingBrace(t *testing.T) {
	topLevel := []byte(strings.Join([]string{
		"#if OUTER",
		"#if INNER",
		"new x = 1;",
		"#endif",
		"#endif",
		"",
	}, "\n"))
	formatted := mustFormat(t, topLevel, config.Default())

	want := "#if OUTER\n    #if INNER\n        new x = 1;\n    #endif\n#endif\n"
	if string(formatted) != want {
		t.Fatalf("top-level nested directives did not indent by nesting depth\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("nested directive indent is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}

	nested := []byte(strings.Join([]string{
		"stock F() {",
		"\tif (cond) {",
		"\t\t#if OUTER",
		"\t\tnew x;",
		"\t\t#if INNER",
		"\t\tnew y;",
		"\t\t#endif",
		"\t\t#endif",
		"\t}",
		"}",
		"",
	}, "\n"))
	nestedFormatted := mustFormat(t, nested, config.Default())
	nestedWant := strings.Join([]string{
		"stock F()",
		"{",
		"    if (cond)",
		"    {",
		"    #if OUTER",
		"        new x;",
		"    #if INNER",
		"        new y;",
		"    #endif",
		"    #endif",
		"    }",
		"}",
		"",
	}, "\n")
	if string(nestedFormatted) != nestedWant {
		t.Fatalf("nested directives did not align with the enclosing if-block's brace\nexpected:\n%s\nactual:\n%s", nestedWant, nestedFormatted)
	}
	nestedSecond := mustFormat(t, nestedFormatted, config.Default())
	if string(nestedSecond) != string(nestedFormatted) {
		t.Fatalf("nested directive indent (in a block) is not idempotent\nfirst:\n%s\nsecond:\n%s", nestedFormatted, nestedSecond)
	}

	cfg := config.Default()
	cfg.DirectiveIndent = config.DirectiveIndentNone
	noneFormatted := mustFormat(t, nested, cfg)
	for line := range strings.SplitSeq(string(noneFormatted), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.HasPrefix(line, " ") {
			t.Fatalf("DirectiveIndentNone should keep every directive at column 0:\n%s", noneFormatted)
		}
	}
}

func TestOnlyEnumsGetTrailingCommas(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "parameter_list",
			source: "stock F(a, b,) {\n    return a + b;\n}\n",
			want:   "stock F(a, b)\n{\n    return a + b;\n}\n",
		},
		{
			name:   "call_arguments",
			source: "stock F() {\n    Call(a, b,);\n}\n",
			want:   "stock F()\n{\n    Call(a, b);\n}\n",
		},
		{
			name:   "array_literal",
			source: "new arr[] = {1, 2,};\n",
			want:   "new arr[] = {1, 2};\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			formatted := mustFormat(t, []byte(tc.source), config.Default())
			if string(formatted) != tc.want {
				t.Fatalf("expected the trailing comma to be stripped, not preserved or forced into an explosion\nexpected:\n%s\nactual:\n%s", tc.want, formatted)
			}

			second := mustFormat(t, formatted, config.Default())
			if string(second) != string(formatted) {
				t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
			}
		})
	}
}

func TestNoTrailingCommaStillCollapsesToOneLine(t *testing.T) {
	source := []byte("stock F() {\n    Call(a, b);\n}\n")
	want := "stock F()\n{\n    Call(a, b);\n}\n"
	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("list without a trailing comma should still collapse\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestAlignConsecutiveDeclarationsAlignsWithinARun(t *testing.T) {
	cfg := config.Default()
	cfg.AlignConsecutiveDeclarations = true
	source := []byte(strings.Join([]string{
		"new gPlayerHealth = 100;",
		"new gPlayerArmor = 0;",
		"new gPlayerScoreLong = 500;",
		"",
		"new gUnrelated = 1;",
		"",
		"stock F() {",
		"    new x = 1;",
		"    static Float:yyyy = 2.0;",
		"    new zz = 3;",
		"    Call();",
		"    new after = 4;",
		"}",
		"",
	}, "\n"))
	want := strings.Join([]string{
		"new gPlayerHealth    = 100;",
		"new gPlayerArmor     = 0;",
		"new gPlayerScoreLong = 500;",
		"",
		"new gUnrelated = 1;",
		"",
		"stock F()",
		"{",
		"    new x              = 1;",
		"    static Float: yyyy = 2.0;",
		"    new zz             = 3;",
		"    Call();",
		"    new after = 4;",
		"}",
		"",
	}, "\n")
	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("consecutive declarations were not aligned as expected\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("aligned declaration output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestAlignConsecutiveDeclarationsDisabledByDefault(t *testing.T) {
	source := []byte("new gShort = 1;\nnew gVeryLongName = 2;\n")
	want := "new gShort = 1;\nnew gVeryLongName = 2;\n"
	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("declarations should stay unaligned by default\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestAlignConsecutiveDeclarationsBreaksOnLeadingComment(t *testing.T) {
	cfg := config.Default()
	cfg.AlignConsecutiveDeclarations = true
	source := []byte("new gShort = 1;\n// a comment\nnew gVeryLongName = 2;\n")
	want := "new gShort = 1;\n// a comment\nnew gVeryLongName = 2;\n"
	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("a leading comment should break the alignment run\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestAlignConsecutiveDeclarationsSkipsMultiDeclaratorStatements(t *testing.T) {
	cfg := config.Default()
	cfg.AlignConsecutiveDeclarations = true
	source := []byte(strings.Join([]string{
		"new gShort = 1;",
		"new gVeryLongName = 2;",
		"new a = 1, b = 2;",
		"new gX = 3;",
		"new gVeryLongTail = 4;",
		"",
	}, "\n"))
	formatted := mustFormat(t, source, cfg)
	text := string(formatted)
	if !strings.Contains(text, "new gShort        = 1;\nnew gVeryLongName = 2;") {
		t.Fatalf("first run before the multi-declarator statement did not align:\n%s", text)
	}
	if !strings.Contains(text, "new a = 1, b = 2;") {
		t.Fatalf("multi-declarator statement should be left untouched:\n%s", text)
	}
	if !strings.Contains(text, "new gX            = 3;\nnew gVeryLongTail = 4;") {
		t.Fatalf("second run after the multi-declarator statement did not align:\n%s", text)
	}
	second := mustFormat(t, formatted, cfg)
	if string(second) != text {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestAssignmentChainStaysOnOneLineWhenItFits(t *testing.T) {
	source := []byte("stock F() {\n    a = b = c = 1;\n}\n")
	want := "stock F()\n{\n    a = b = c = 1;\n}\n"
	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("assignment chain mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("assignment chain output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestAssignmentChainWrapsOneAssignmentPerLineWhenTooLong(t *testing.T) {
	cfg := config.Default()
	cfg.LineWidth = 40
	source := []byte("stock F() {\n    firstVariable = secondVariable = thirdVariable = 1;\n}\n")
	want := strings.Join([]string{
		"stock F()",
		"{",
		"    firstVariable =",
		"        secondVariable =",
		"        thirdVariable = 1;",
		"}",
		"",
	}, "\n")
	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("wrapped assignment chain mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("wrapped assignment chain is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestAssignmentChainRespectsSpaceAroundOperatorsFalse(t *testing.T) {
	cfg := config.Default()
	cfg.SpaceAroundOperators = false
	source := []byte("stock F() {\n    a = b = c = 1;\n}\n")
	want := "stock F()\n{\n    a= b= c=1;\n}\n"
	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("assignment chain with SpaceAroundOperators=false mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestStringConcatJoinsAdjacentLiteralsWithASpace(t *testing.T) {
	source := []byte("new s[64] = \"a\" \"b\" \"c\";\n")
	want := "new s[64] = \"a\" \"b\" \"c\";\n"
	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("string concat mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("string concat output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestStringConcatNormalizesExtraWhitespaceBetweenPieces(t *testing.T) {
	source := []byte("new s[64] = \"a\"    \"b\";\n")
	want := "new s[64] = \"a\" \"b\";\n"
	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("string concat whitespace normalization mismatch\nexpected:\n%s\nactual:\n%s", want, formatted)
	}
}

func TestBareIdentifierBeforeParameterIsNotTreatedAsItsOwnParameter(t *testing.T) {
	t.Parallel()

	source := []byte("#define WC_CONST\nstock F(playerid, WC_CONST animlib[])\n{\n    return 1;\n}\n")
	want := "#define WC_CONST\n\nstock F(playerid, WC_CONST animlib[])\n{\n    return 1;\n}\n"

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("expected the macro qualifier and parameter to stay one parameter, with no comma between them\nexpected:\n%s\nactual:\n%s", want, formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestStateStatementKeepsTagQualifiedTarget(t *testing.T) {
	t.Parallel()

	source := []byte("public F()\n{\n    state _ALS : _ALS_go;\n    return 1;\n}\n")
	want := "public F()\n{\n    state _ALS: _ALS_go;\n    return 1;\n}\n"

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != want {
		t.Fatalf("expected the tag-qualified state target to survive formatting\nexpected:\n%s\nactual:\n%s", want, formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestTrailingLineCommentInGroupForcesABreakInsteadOfSwallowingCode(t *testing.T) {
	t.Parallel()

	source := []byte("stock F(x)\n{\n    if (x == 1 // a\n    || x == 2)\n    {\n        return 1;\n    }\n    return 0;\n}\n")

	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	if !strings.Contains(text, "x == 2") {
		t.Fatalf("the second operand was swallowed into the preceding line comment:\n%s", text)
	}

	if idx := strings.Index(text, "// a"); idx >= 0 {
		lineEnd := strings.IndexByte(text[idx:], '\n')
		if lineEnd < 0 {
			lineEnd = len(text) - idx
		}

		if strings.TrimSpace(text[idx+len("// a"):idx+lineEnd]) != "" {
			t.Fatalf("content was printed on the same line as the trailing \"//\" comment:\n%s", text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestEmitDirectiveKeepsStatementIndentInsteadOfAligningWithBrace(t *testing.T) {
	t.Parallel()

	source := []byte("stock F()\n{\n    while (1)\n    {\n        #emit LCTRL 5\n        #emit LOAD.alt 1\n    }\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != string(source) {
		t.Fatalf("expected #emit to keep its statement indent\nexpected:\n%s\nactual:\n%s", source, formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestIndentNestedDirectivesIndentsTopLevelBranchContents(t *testing.T) {
	t.Parallel()

	source := []byte(strings.Join([]string{
		"#if defined _INC_y_va",
		"#if defined _INC_open_mp",
		"stock F() {}",
		"#else",
		"stock F() {}",
		"#endif",
		"#else",
		"stock F() {}",
		"#endif",
		"",
	}, "\n"))
	want := strings.Join([]string{
		"#if defined _INC_y_va",
		"    #if defined _INC_open_mp",
		"        stock F()",
		"        { }",
		"    #else",
		"        stock F()",
		"        { }",
		"    #endif",
		"#else",
		"    stock F()",
		"    { }",
		"#endif",
		"",
	}, "\n")

	cfg := config.Default()
	cfg.IndentNestedDirectives = true

	formatted := mustFormat(t, source, cfg)
	if string(formatted) != want {
		t.Fatalf("expected nested top-level directive branches to be indented\nexpected:\n%s\nactual:\n%s", want, formatted)
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}

	off := config.Default()
	off.IndentNestedDirectives = false

	offFormatted := mustFormat(t, source, off)

	offWant := strings.Join([]string{
		"#if defined _INC_y_va",
		"#if defined _INC_open_mp",
		"stock F()",
		"{ }",
		"#else",
		"stock F()",
		"{ }",
		"#endif",
		"#else",
		"stock F()",
		"{ }",
		"#endif",
		"",
	}, "\n")
	if string(offFormatted) != offWant {
		t.Fatalf("expected IndentNestedDirectives=false to keep top-level directives flat\nexpected:\n%s\nactual:\n%s", offWant, offFormatted)
	}
}

func TestIndentNestedDirectivesDoesNotDoubleIndentInsideABlock(t *testing.T) {
	t.Parallel()

	source := []byte("stock F()\n{\n    if (cond)\n    {\n    #if OUTER\n        new x;\n    #if INNER\n        new y;\n    #endif\n    #endif\n    }\n}\n")

	cfg := config.Default()
	cfg.IndentNestedDirectives = true

	formatted := mustFormat(t, source, cfg)
	if string(formatted) != string(source) {
		t.Fatalf("expected in-block directive alignment to be unaffected\nexpected:\n%s\nactual:\n%s", source, formatted)
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestDirectiveAfterBlankLineStillAlignsWithEnclosingBrace(t *testing.T) {
	t.Parallel()

	source := []byte("stock F()\n{\n    if (outer)\n    {\n        if (before)\n        {\n            Before();\n        }\n\n    #if FEATURE\n        Call();\n    #endif\n    }\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != string(source) {
		t.Fatalf("expected the #if to stay aligned with its #endif despite the preceding blank line\nexpected:\n%s\nactual:\n%s", source, formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("output is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}
