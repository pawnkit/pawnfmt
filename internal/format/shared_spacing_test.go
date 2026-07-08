package format_test

import (
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestSharedExpressionDirectiveInsertsMissingKeywordSpace(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\n#if(INNER)\nCall();\n#endif\n#else\nif(second) {\n#endif\nreturn 1;\n}\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "#if (INNER)") || strings.Contains(string(formatted), "#if(INNER)") {
		t.Fatalf("shared expression directive retained missing keyword spacing:\n%s", formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared expression directive spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestUnknownRawDirectivePreservesPayloadAdjacency(t *testing.T) {
	source := []byte("#custom(payload)\n")

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != string(source) {
		t.Fatalf("unknown directive payload adjacency changed\nexpected: %q\nactual:   %q", source, formatted)
	}
}

func TestSharedRegionCollapsesRequiredKeywordSpacing(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nnew    value; result = sizeof    value; return    value;\n#if defined    FEATURE\nCall();\n#endif\n#else\nif(second) {\n#endif\nreturn 1;\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{"new value;", "sizeof value;", "return value;", "#if defined FEATURE"} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared required keyword spacing missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared keyword spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedRegionCollapsesCustomQualifierSpacing(t *testing.T) {
	source := []byte("#if A\nac_fpublic    Handler(value)\n{\n#else\nac_fpublic    Handler(other)\n{\n#endif\nreturn 1;\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if strings.Contains(string(formatted), "ac_fpublic    Handler") ||
		!strings.Contains(string(formatted), "ac_fpublic Handler") {
		t.Fatalf("shared custom qualifier spacing was not normalized:\n%s", formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared custom qualifier spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedRegionSeparatesAdjacentStrings(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nvalue = \"first\"\"second\"; packed = !\"third\"!\"fourth\";\n#else\nif(second) {\n#endif\nreturn 1;\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{"\"first\" \"second\"", "!\"third\" !\"fourth\""} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared adjacent strings missing separator %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared adjacent string spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedRegionBracesMultilineControlBody(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first &&\nsecond)\nCallOne();\n#else\nif(third ||\nfourth)\nCallTwo();\n#endif\n}\n")
	formatted := mustFormat(t, source, config.Default())
	text := string(formatted)

	for _, body := range []string{"CallOne();", "CallTwo();"} {
		needle := "{\n        " + body + "\n    }"
		if !strings.Contains(text, needle) {
			t.Fatalf("multiline shared control body %q did not receive braces:\n%s", body, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("multiline shared control braces are not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedDirectiveFreeBlockCorrectsBodyWithoutMovingClose(t *testing.T) {
	source := []byte("stock F() {\n#if A\nswitch(value)\n{\ncase 1:\n{\nCallOne();\nif(flag)\n{\nCallTwo();\n}\n}\n}\nif(first) {\n#else\nif(second) {\n#endif\nCallThree();\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	if !strings.Contains(text, "case 1:\n        {\n            CallOne();") {
		t.Fatalf("directive-free shared block body was not corrected:\n%s", text)
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "case 1:" || i+1 >= len(lines) {
			continue
		}

		openIndent := len(lines[i+1]) - len(strings.TrimLeft(lines[i+1], " \t"))
		depth := 1

		for j := i + 2; j < len(lines); j++ {
			switch strings.TrimSpace(lines[j]) {
			case "{":
				depth++
			case "}":
				depth--
			}

			if depth == 0 {
				closeIndent := len(lines[j]) - len(strings.TrimLeft(lines[j], " \t"))
				if closeIndent != openIndent {
					t.Fatalf("case block braces do not align (%d != %d):\n%s", openIndent, closeIndent, text)
				}

				break
			}
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("directive-free shared block correction is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedRegionNormalizesRangeOperatorSpacing(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nnew value = 14..16;\n#else\nif(second) {\n#endif\nreturn value;\n}\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "new value = 14 .. 16;") {
		t.Fatalf("shared range operator spacing was not normalized:\n%s", formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared range spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}

	compact := config.Default()
	compact.SpaceAroundOperators = false

	compactFormatted := mustFormat(t, source, compact)
	if !strings.Contains(string(compactFormatted), "14 .. 16") {
		t.Fatalf("range spacing must remain consistent when other operators are compact:\n%s", compactFormatted)
	}
}

func TestSharedRegionNormalizesUpdateOperatorSpacing(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nvalue ++; ++ other;\n#else\nif(second) {\n#endif\nreturn value;\n}\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "value++; ++other;") {
		t.Fatalf("shared update operator spacing was not normalized:\n%s", formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared update spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedRegionNormalizesTightPunctuation(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nCall(. target = value, . reliability = 1); result = Namespace :: Symbol;\n#else\nif(second) {\n#endif\nreturn result;\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{".target = value, .reliability = 1", "Namespace::Symbol"} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared tight punctuation missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared tight punctuation is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedRegionNormalizesSeparatorSpacing(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nCall(first , second) ; value = 1;next = 2;\n#else\nif(second) {\n#endif\nreturn value;\n}\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "Call(first, second); value = 1; next = 2;") {
		t.Fatalf("shared separator spacing was not normalized:\n%s", formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared separator spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedRegionNormalizesCallAndSubscriptAdjacency(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nCall (value); result = items [index];\n#else\nif(second) {\n#endif\nreturn result;\n}\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "Call(value); result = items[index];") {
		t.Fatalf("shared call/subscript adjacency was not normalized:\n%s", formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared call/subscript adjacency is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedDeclarationOptionsDoNotLeakIntoInitializer(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nnew result = Call (items [index], values [other]), second [4];\n#else\nif(second) {\n#endif\nreturn result;\n}\n}\n")
	cfg := config.Default()
	cfg.SpaceBeforeFunctionParen = true
	cfg.SpaceBeforeArrayBrackets = true

	formatted := mustFormat(t, source, cfg)
	if !strings.Contains(string(formatted), "Call(items[index], values[other])") ||
		!strings.Contains(string(formatted), "second [4]") {
		t.Fatalf("declaration spacing options leaked into an initializer expression:\n%s", formatted)
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != string(formatted) {
		t.Fatalf("shared initializer adjacency is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}
}

func TestSharedRegionNormalizesKeywordParenAdjacency(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nresult = sizeof (items) + tagof (value); return(value);\n#else\nif(second) {\n#endif\nreturn result;\n}\n}\n")

	formatted := mustFormat(t, source, config.Default())
	if !strings.Contains(string(formatted), "sizeof(items) + tagof(value); return (value);") {
		t.Fatalf("shared keyword/paren adjacency was not normalized:\n%s", formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("shared keyword/paren adjacency is not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, second)
	}

	padded := config.Default()
	padded.SpaceInsideParens = true

	paddedFormatted := mustFormat(t, source, padded)
	if !strings.Contains(string(paddedFormatted), "sizeof(items) + tagof(value)") {
		t.Fatalf("keyword-owned parentheses incorrectly inherited inner padding:\n%s", paddedFormatted)
	}
}

func TestSharedRegionDistinguishesTagCastAndTernaryColons(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\nnew custom : value; result = condition?YES:NO; nested = first?second?A:B:C; result = Float : 1.0;\n#else\nif(second) {\n#endif\nreturn result;\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{"custom: value", "condition ? YES : NO", "first ? second ? A : B : C", "Float:1.0"} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared colon role formatting missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared colon role formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedOverlappingWhitespaceEditsDoNotPanic(t *testing.T) {
	source := []byte("#if ? ? ?   ?")
	formatted := mustFormat(t, source, config.Default())

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("overlapping shared whitespace edits are not idempotent\nfirst:  %q\nsecond: %q", formatted, second)
	}
}

func TestSharedRegionNormalizesCaseAndLabelColons(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif(first) {\ncase 1 : Call();\nretry : Call();\n#else\nif(second) {\n#endif\nreturn 1;\n}\n}\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{"case 1: Call();", "retry: Call();"} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared separator colon missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("shared separator colon formatting is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestOpaqueTokenPastingMacroIsPreservedAndIdempotent(t *testing.T) {
	source := []byte("#define ac_fpublic%0(%1) forward%0(%1); public%0(%1)\n")

	formatted := mustFormat(t, source, config.Default())
	if string(formatted) != string(source) {
		t.Fatalf("opaque token-pasting macro changed\nexpected:\n%s\nactual:\n%s", source, formatted)
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != string(formatted) {
		t.Fatalf("opaque token-pasting macro is not idempotent:\n%s", second)
	}
}

func TestSharedConditionalRespectsInnerSpacingOptions(t *testing.T) {
	source := []byte("stock F() {\n#if A\nif( Check( { -1, 2 } ) ) {\n#else\nif( items[ index ] ) {\n#endif\nreturn 1;\n}\n}\n")
	requireSharedConditionalPath(t, source)

	compact := mustFormat(t, source, config.Default())
	if !strings.Contains(string(compact), "Check({-1, 2})") || !strings.Contains(string(compact), "items[index]") {
		t.Fatalf("shared conditional did not remove inner spacing:\n%s", compact)
	}

	cfg := config.Default()
	cfg.SpaceInsideParens = true
	cfg.SpaceInsideBrackets = true
	cfg.SpaceInsideBraces = true

	padded := mustFormat(t, source, cfg)
	if !strings.Contains(string(padded), "Check( { -1, 2 } )") || !strings.Contains(string(padded), "items[ index ]") {
		t.Fatalf("shared conditional did not add inner spacing:\n%s", padded)
	}
}

func TestSharedForBoundaryIgnoresInnerParenPadding(t *testing.T) {
	source := []byte("stock F() {\n#if A\nfor( ; value; ++value ) {\nfor(;;) Call(value);\n#else\nfor( ; other; ++other ) {\nfor(;other;) Call(other);\n#endif\nCall( value );\n}\n}\n")
	cfg := config.Default()
	cfg.SpaceInsideParens = true
	formatted := mustFormat(t, source, cfg)

	text := string(formatted)
	for _, want := range []string{"for (; value; ++value)", "for (; other; ++other)", "for (;; )", "for (; other; )", "Call( value )"} {
		if !strings.Contains(text, want) {
			t.Fatalf("shared for-boundary spacing missing %q:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, cfg)
	if string(second) != text {
		t.Fatalf("shared for-boundary spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}

func TestSharedRegionKeepsOperatorTightBeforeSemicolonAndComma(t *testing.T) {
	source := []byte("#if ?\nx < y>;\nx < y,\n")
	formatted := mustFormat(t, source, config.Default())

	text := string(formatted)
	for _, want := range []string{"x < y >;", "x < y,"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected tight operator-before-punctuation %q, got:\n%s", want, text)
		}
	}

	second := mustFormat(t, formatted, config.Default())
	if string(second) != text {
		t.Fatalf("operator-before-punctuation spacing is not idempotent\nfirst:\n%s\nsecond:\n%s", text, second)
	}
}
