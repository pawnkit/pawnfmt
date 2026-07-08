package format

import (
	"sort"
	"strings"

	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/config"
)

type textEdit struct {
	start int
	end   int
	text  string
}

func normalizeSharedLine(line string, cfg config.Config) string {
	trimmedLine := strings.TrimSpace(line)
	if strings.HasPrefix(trimmedLine, "#") {
		if !sharedExpressionDirective(trimmedLine) {
			return strings.TrimRight(line, " \t")
		}

		line = ensureDirectiveKeywordSpacing(line)
	}

	line = normalizeSharedComments(line)
	tokens := lexer.Tokenize([]byte(line))
	arrayBraces := sharedArrayBraceTokens(tokens)
	declarationLike := sharedDeclarationLine(tokens)

	var edits []textEdit

	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i].Kind == token.EOF {
			continue
		}

		if normalizeSharedPunctuation(line, cfg, tokens, declarationLike, i, &edits) ||
			normalizeSharedDelimiters(line, cfg, tokens, arrayBraces, i, &edits) ||
			normalizeSharedOperatorsAndAdjacency(line, cfg, tokens, declarationLike, i, &edits) {
			continue
		}
	}

	return strings.TrimRight(applySharedTextEdits(line, edits), " \t")
}

func normalizeSharedPunctuation(line string, cfg config.Config, tokens []token.Token, declarationLike bool, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]

	ternarySeparator := cur.Kind == token.Question || cur.Kind == token.Colon && sharedTernaryColon(tokens, i)
	if ternarySeparator {
		if i > 0 {
			prev := tokens[i-1]
			if prev.End.Offset <= cur.Start.Offset && horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
				*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset, text: " "})
			}
		}

		if next.Kind != token.EOF && cur.End.Offset <= next.Start.Offset &&
			horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		}

		return true
	}

	caseOrLabelColon := cur.Kind == token.Colon && sharedCaseOrLabelColon(tokens, i, []byte(line))
	if next.Kind == token.EOF {
		if caseOrLabelColon && i > 0 {
			prev := tokens[i-1]
			if prev.End.Offset < cur.Start.Offset && horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
				*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset})
			}
		}

		return true
	}

	if caseOrLabelColon {
		prev := tokens[i-1]
		if prev.End.Offset < cur.Start.Offset && horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
			*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset})
		}

		if cur.End.Offset <= next.Start.Offset && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		}

		return true
	}

	if cur.Kind == token.DotDot {
		if i > 0 {
			prev := tokens[i-1]
			if prev.End.Offset <= cur.Start.Offset && horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
				*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset, text: " "})
			}
		}

		if cur.End.Offset <= next.Start.Offset && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		}

		return true
	}

	if cur.Kind == token.Dot || cur.Kind == token.ColonColon {
		if i > 0 {
			prev := tokens[i-1]

			containerSpacing := cur.Kind == token.Dot && (prev.Kind == token.Comma || prev.Kind == token.LParen)
			if !containerSpacing && prev.End.Offset < cur.Start.Offset && horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
				*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset})
			}
		}

		if cur.End.Offset < next.Start.Offset && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		}

		return true
	}

	if (next.Kind == token.Comma || next.Kind == token.Semicolon) &&
		cur.End.Offset < next.Start.Offset && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	if cur.Kind == token.Semicolon && next.Kind != token.RParen && next.Kind != token.Semicolon &&
		cur.End.Offset <= next.Start.Offset &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		return true
	}

	if sharedStringLiteral(cur.Kind) && sharedStringLiteral(next.Kind) &&
		cur.End.Offset <= next.Start.Offset && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		return true
	}

	if sharedKeywordNeedsSpace(cur.Kind, next.Kind) && cur.End.Offset < next.Start.Offset &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		return true
	}

	if declarationLike && sharedBeforeInitializer(tokens, i) &&
		cur.Kind == token.Identifier && next.Kind == token.Identifier &&
		cur.End.Offset < next.Start.Offset && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		return true
	}

	return false
}

func normalizeSharedDelimiters(line string, cfg config.Config, tokens []token.Token, arrayBraces map[int]bool, i int, edits *[]textEdit) bool {
	return normalizeSharedTagAndComma(line, cfg, tokens, i, edits) ||
		normalizeSharedParentheses(line, cfg, tokens, i, edits) ||
		normalizeSharedContainers(line, cfg, tokens, arrayBraces, i, edits)
}

func normalizeSharedTagAndComma(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	declarationTag := cur.Kind == token.Colon && sharedDeclarationTagColon(tokens, i)

	castTag := cur.Kind == token.Colon && i > 0 && tokens[i-1].Kind == token.Identifier &&
		sharedTagName(tokens[i-1].Text([]byte(line))) && !sharedTernaryColon(tokens, i)
	if cur.Kind == token.Colon && next.Kind != token.Colon && i > 0 &&
		(declarationTag || castTag) && cfg.TagColonSpacing == config.TagColonSpacingTight {
		prev := tokens[i-1]
		if horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
			*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset})
		}

		if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			after := ""
			if declarationTag {
				after = " "
			}

			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: after})
		}

		return true
	}

	if cur.Kind == token.Comma && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceAfterComma {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	return false
}

func normalizeSharedParentheses(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if cur.Kind == token.LParen && next.Kind == token.RParen &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	if cur.Kind == token.LParen && sharedForOpeningParen(tokens, i) &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	if cur.Kind == token.LParen && sharedTightKeywordOpeningParen(tokens, i) &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	if cur.Kind == token.LParen && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideParens {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if next.Kind == token.RParen && sharedForClosingParen(tokens, i+1) &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cur.Kind == token.Semicolon {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if next.Kind == token.RParen && sharedTightKeywordClosingParen(tokens, i+1) &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	if next.Kind == token.RParen && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideParens {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	return false
}

func normalizeSharedContainers(line string, cfg config.Config, tokens []token.Token, arrayBraces map[int]bool, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if cur.Kind == token.LBracket && next.Kind == token.RBracket &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBrackets {
			space = "  "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if cur.Kind == token.LBracket && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBrackets {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if next.Kind == token.RBracket && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBrackets {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if cur.Kind == token.LBrace && arrayBraces[i] && next.Kind == token.RBrace &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBraces {
			space = "  "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if cur.Kind == token.LBrace && arrayBraces[i] &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBraces {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if next.Kind == token.RBrace && arrayBraces[i+1] &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBraces {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	return false
}

func normalizeSharedOperatorsAndAdjacency(line string, cfg config.Config, tokens []token.Token, declarationLike bool, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if sharedBinaryOperator(tokens, i) {
		space := ""
		if cfg.SpaceAroundOperators {
			space = " "
		}

		if i > 0 {
			prev := tokens[i-1]
			if horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
				*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset, text: space})
			}
		}

		if next.Kind != token.Comma && next.Kind != token.Semicolon &&
			horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})
		}

		return true
	}

	if sharedPostfixOperator(tokens, i) && i > 0 {
		prev := tokens[i-1]
		if horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
			*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset})
		}

		return true
	}

	if sharedPrefixOperator(tokens, i) &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	if next.Kind == token.LParen {
		if declarationLike && sharedBeforeInitializer(tokens, i) && cur.Kind == token.Identifier &&
			horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			space := ""
			if cfg.SpaceBeforeFunctionParen {
				space = " "
			}

			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

			return true
		}

		switch cur.Kind {
		case token.KwIf, token.KwFor, token.KwWhile, token.KwSwitch:
			if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
				*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
			}
		case token.KwReturn:
			if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
				*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
			}
		case token.KwSizeof, token.KwTagof, token.KwDefined:
			if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
				*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
			}
		case token.Identifier, token.RParen, token.RBracket:
			if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
				*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
			}
		}
	}

	declaratorBracket := declarationLike && sharedBeforeInitializer(tokens, i)
	if next.Kind == token.LBracket && !declaratorBracket &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	if declaratorBracket && cur.Kind == token.Identifier && next.Kind == token.LBracket &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceBeforeArrayBrackets {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	return false
}

func applySharedTextEdits(line string, edits []textEdit) string {
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start == edits[j].start {
			return edits[i].end < edits[j].end
		}

		return edits[i].start < edits[j].start
	})
	seen := make(map[[2]int]bool, len(edits))

	rightmostStart := len(line) + 1
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]

		key := [2]int{edit.start, edit.end}
		if seen[key] || edit.start < 0 || edit.start > edit.end || edit.end > len(line) || edit.end > rightmostStart {
			continue
		}

		seen[key] = true
		line = line[:edit.start] + edit.text + line[edit.end:]
		rightmostStart = edit.start
	}

	return line
}
