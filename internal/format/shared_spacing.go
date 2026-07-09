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
	return normalizeSharedTernaryAndCaseColon(line, tokens, i, edits) ||
		normalizeSharedRangeAndAccess(line, tokens, i, edits) ||
		normalizeSharedPunctuationAdjacency(line, tokens, i, edits) ||
		normalizeSharedLiteralAndKeywordAdjacency(line, cfg, tokens, declarationLike, i, edits)
}

func normalizeSharedTernaryAndCaseColon(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	return normalizeSharedTernarySeparator(line, tokens, i, edits) ||
		normalizeSharedCaseOrLabelColon(line, tokens, i, edits)
}

func normalizeSharedTernarySeparator(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]

	ternarySeparator := cur.Kind == token.Question || cur.Kind == token.Colon && sharedTernaryColon(tokens, i)
	if !ternarySeparator {
		return false
	}

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

func normalizeSharedCaseOrLabelColon(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]

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

	if !caseOrLabelColon {
		return false
	}

	prev := tokens[i-1]
	if prev.End.Offset < cur.Start.Offset && horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
		*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset})
	}

	if cur.End.Offset <= next.Start.Offset && horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
	}

	return true
}

func normalizeSharedRangeAndAccess(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	return normalizeSharedDotDotRange(line, tokens, i, edits) ||
		normalizeSharedMemberAccess(line, tokens, i, edits)
}

func normalizeSharedDotDotRange(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if cur.Kind != token.DotDot {
		return false
	}

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

func normalizeSharedMemberAccess(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if cur.Kind != token.Dot && cur.Kind != token.ColonColon {
		return false
	}

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

func normalizeSharedPunctuationAdjacency(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]

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

	return false
}

func normalizeSharedLiteralAndKeywordAdjacency(line string, cfg config.Config, tokens []token.Token, declarationLike bool, i int, edits *[]textEdit) bool {
	return normalizeSharedLiteralOrKeywordGap(line, tokens, i, edits) ||
		normalizeSharedDeclaratorIdentifierGap(line, cfg, tokens, declarationLike, i, edits)
}

func normalizeSharedLiteralOrKeywordGap(line string, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]

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

	return false
}

func normalizeSharedDeclaratorIdentifierGap(line string, _ config.Config, tokens []token.Token, declarationLike bool, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]

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
	return normalizeSharedTagColon(line, cfg, tokens, i, edits) ||
		normalizeSharedCommaSpacing(line, cfg, tokens, i, edits)
}

// isSharedDeclarationTagColon reports whether tokens[i] is a "Tag:" colon
// belonging to a declaration (e.g. "Float:") rather than a cast expression.
func isSharedCastTagColon(line string, tokens []token.Token, i int) bool {
	return tokens[i].Kind == token.Colon && i > 0 && tokens[i-1].Kind == token.Identifier &&
		sharedTagName(tokens[i-1].Text([]byte(line))) && !sharedTernaryColon(tokens, i)
}

func normalizeSharedTagColon(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	declarationTag := cur.Kind == token.Colon && sharedDeclarationTagColon(tokens, i)
	castTag := isSharedCastTagColon(line, tokens, i)

	tight := cfg.TagColonSpacing == config.TagColonSpacingTight
	compact := cfg.TagColonSpacing == config.TagColonSpacingCompact

	if cur.Kind != token.Colon || next.Kind == token.Colon || i == 0 ||
		!declarationTag && !castTag || !tight && !compact {
		return false
	}

	prev := tokens[i-1]
	if horizontalGap(line[prev.End.Offset:cur.Start.Offset]) {
		*edits = append(*edits, textEdit{start: prev.End.Offset, end: cur.Start.Offset})
	}

	if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		after := ""
		if declarationTag && tight {
			after = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: after})
	}

	return true
}

func normalizeSharedCommaSpacing(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]

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
	return normalizeSharedAfterOpenParen(line, cfg, tokens, i, edits) ||
		normalizeSharedBeforeCloseParen(line, cfg, tokens, i, edits)
}

func normalizeSharedAfterOpenParen(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if cur.Kind != token.LParen || !horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		return false
	}

	if next.Kind == token.RParen || sharedForOpeningParen(tokens, i) || sharedTightKeywordOpeningParen(tokens, i) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	space := ""
	if cfg.SpaceInsideParens {
		space = " "
	}

	*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

	return true
}

func normalizeSharedBeforeCloseParen(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if next.Kind != token.RParen || !horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		return false
	}

	if sharedForClosingParen(tokens, i+1) {
		space := ""
		if cur.Kind == token.Semicolon {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if sharedTightKeywordClosingParen(tokens, i+1) {
		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		return true
	}

	space := ""
	if cfg.SpaceInsideParens {
		space = " "
	}

	*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

	return true
}

func normalizeSharedContainers(line string, cfg config.Config, tokens []token.Token, arrayBraces map[int]bool, i int, edits *[]textEdit) bool {
	return normalizeSharedBrackets(line, cfg, tokens, i, edits) ||
		normalizeSharedArrayBraces(line, cfg, tokens, arrayBraces, i, edits)
}

func normalizeSharedBrackets(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
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

	if (cur.Kind == token.LBracket || next.Kind == token.RBracket) &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBrackets {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	return false
}

func normalizeSharedArrayBraces(line string, cfg config.Config, tokens []token.Token, arrayBraces map[int]bool, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if cur.Kind == token.LBrace && arrayBraces[i] && next.Kind == token.RBrace &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceInsideBraces {
			space = "  "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	if (cur.Kind == token.LBrace && arrayBraces[i] || next.Kind == token.RBrace && arrayBraces[i+1]) &&
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
	if normalizeSharedOperatorAdjacency(line, cfg, tokens, i, edits) {
		return true
	}

	if normalizeSharedParenAdjacency(line, cfg, tokens, declarationLike, i, edits) {
		return true
	}

	return normalizeSharedBracketAdjacency(line, cfg, tokens, declarationLike, i, edits)
}

// applySharedBinaryOperatorSpacing normalizes the gaps on both sides of a
// binary operator token, respecting cfg.SpaceAroundOperators.
func applySharedBinaryOperatorSpacing(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) {
	cur, next := tokens[i], tokens[i+1]

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
}

func normalizeSharedOperatorAdjacency(line string, cfg config.Config, tokens []token.Token, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if sharedBinaryOperator(tokens, i) {
		applySharedBinaryOperatorSpacing(line, cfg, tokens, i, edits)
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

	return false
}

func normalizeSharedParenAdjacency(line string, cfg config.Config, tokens []token.Token, declarationLike bool, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
	if next.Kind != token.LParen {
		return false
	}

	if declarationLike && sharedBeforeInitializer(tokens, i) && cur.Kind == token.Identifier &&
		horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
		space := ""
		if cfg.SpaceBeforeFunctionParen {
			space = " "
		}

		*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: space})

		return true
	}

	normalizeSharedKeywordParenGap(line, tokens, i, edits)

	return false
}

// normalizeSharedKeywordParenGap handles the space (or lack of one) before an
// opening paren that follows a keyword, identifier, or closing bracket/paren.
func normalizeSharedKeywordParenGap(line string, tokens []token.Token, i int, edits *[]textEdit) {
	cur, next := tokens[i], tokens[i+1]

	//nolint:exhaustive // only control-keyword token kinds matter here
	switch cur.Kind {
	case token.KwIf, token.KwFor, token.KwWhile, token.KwSwitch, token.KwReturn:
		if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset, text: " "})
		}
	case token.KwSizeof, token.KwTagof, token.KwDefined, token.Identifier, token.RParen, token.RBracket:
		if horizontalGap(line[cur.End.Offset:next.Start.Offset]) {
			*edits = append(*edits, textEdit{start: cur.End.Offset, end: next.Start.Offset})
		}
	}
}

func normalizeSharedBracketAdjacency(line string, cfg config.Config, tokens []token.Token, declarationLike bool, i int, edits *[]textEdit) bool {
	cur, next := tokens[i], tokens[i+1]
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
