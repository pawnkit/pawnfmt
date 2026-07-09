package format_test

import (
	"bytes"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func configFromVariant(variant uint8) config.Config {
	cfg := config.Default()

	if variant&0x01 != 0 {
		cfg.SpaceInsideParens = true
		cfg.SpaceInsideBrackets = true
	}

	if variant&0x02 != 0 {
		cfg.SpaceInsideBraces = true
	}

	if variant&0x04 != 0 {
		cfg.AlignConsecutiveDeclarations = true
		cfg.AlignEnumFields = true
	}

	if variant&0x08 != 0 {
		cfg.MultilineCallArgs = config.MultilineListOnePerLine
		cfg.MultilineFunctionParams = config.MultilineListOnePerLine
	}

	if variant&0x40 != 0 {
		cfg.AlignConsecutiveMacros = true
		cfg.AlignTrailingComments = true
	}

	if variant&0x80 != 0 {
		cfg.BreakBinaryOperator = config.BinaryOperatorBreakBefore
		cfg.IndentCaseLabels = false
		cfg.IndentGotoLabels = false
		cfg.SortIncludes = true
		cfg.GroupIncludesByBrackets = true
		cfg.ContinuationIndentWidth = 8
	}

	switch (variant >> 4) % 6 {
	case 1:
		cfg.BraceStyle = config.BraceStyle1TBS
		cfg.IndentStyle = config.IndentStyleTab
		cfg.SingleStatementBraces = config.SingleStatementBracesPreserve
		cfg.SpaceAroundOperators = false
		cfg.SpaceAfterComma = false
	case 2:
		cfg.BraceStyle = config.BraceStyleWhitesmiths
		cfg.IndentWidth = 2
		cfg.LineWidth = 24
	case 3:
		cfg.SingleStatementBraces = config.SingleStatementBracesNever
		cfg.DirectiveIndent = config.DirectiveIndentNone
		cfg.LineWidth = 40
	case 4:
		cfg.DirectiveIndent = config.DirectiveIndentKeepInBlock
	case 5:
		cfg.BraceStyle = config.BraceStyleAllman
	}

	return cfg
}

func FuzzFormatConverges(f *testing.F) {
	seeds := []string{
		"stock F() { if (x) return 1; }\n",
		"#if A\nnew x = 1;\n#else\nnew x = 2;\n#endif\n",
		"stock F() {\n#if A\nif (x) {\n#else\nif (y) {\n#endif\nreturn 1;\n}\n}\n",
		"#define JOIN%0(%1) forward%0(%1); public%0(%1)\n",
		"stock F() { foreach(new i : Player) Use(i); }\n",
		"stock F() { Call(a, b,); }\n",
		"new arr[] = {1, 2,};\n",
		"stock F(a, b,) { return a + b; }\n",
		"new gShort = 1;\nnew gVeryLongName = 2;\n",
		"stock F() {\n    new x = 1;\n    Call();\n    new y = 2;\n}\n",
		"enum X {\n#if A\n    A_VAL,\n#else\n    B_VAL,\n#endif\n    C_VAL,\n};\n",
		"stock F() {\n    if (cond) {\n        #if A\n        new x;\n        #if B\n        new y;\n        #endif\n        #endif\n    }\n}\n",
		"#define SHORT 1\n#define MUCH_LONGER 2\n",
		"stock F() {\n    new x = 1; // a\n    new muchLongerName = 2; // b\n}\n",
		"stock F() {\n    goto Skip;\n    new x = 1;\n    Skip:\n    return x;\n}\n",
		"stock F(x) {\n    switch (x) {\n        case 1: return 1;\n        default: return 0;\n    }\n}\n",
		"stock F() {\n    return aaaaaaaaaa + bbbbbbbbbb + cccccccccc + dddddddddd + eeeeeeeeee;\n}\n",
		"#include \"local.inc\"\n#include <a_samp>\n",
	}

	variants := []uint8{0x00, 0x0F, 0x10, 0x1F, 0x22, 0x2D, 0x35, 0x3A, 0x40, 0x48, 0x4F, 0x50, 0x5F, 0x80, 0xC0, 0xFF}
	for _, seed := range seeds {
		for _, variant := range variants {
			f.Add(seed, variant)
		}
	}

	f.Fuzz(func(t *testing.T, source string, variant uint8) {
		cfg := configFromVariant(variant)

		first, err := formatter.Source([]byte(source), cfg)
		if err != nil {
			return
		}

		second, err := formatter.Source(first, cfg)
		if err != nil {
			t.Fatalf("second format failed after a successful first pass: %v\noutput:\n%s", err, first)
		}

		if !bytes.Equal(first, second) {
			t.Fatalf("successful formatting did not reach a fixed point\nfirst:\n%s\nsecond:\n%s", first, second)
		}
	})
}
