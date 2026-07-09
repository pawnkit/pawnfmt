package format

import (
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/config"
)

func (s *state) formatLiteral(n *parser.Node) string {
	return normalizeNumericLiteralCase(n.Text(s.source), n.Tok.Kind, s.config.NumericLiteralCase)
}

func normalizeNumericLiteralCase(text string, kind token.Kind, mode config.NumericLiteralCase) string {
	if mode == config.NumericLiteralCasePreserve {
		return text
	}

	//nolint:exhaustive // only numeric literal kinds carry a case-sensitive digit or exponent
	switch kind {
	case token.IntLiteral:
		return normalizeHexDigitCase(text, mode)
	case token.FloatLiteral:
		return normalizeExponentCase(text, mode)
	default:
		return text
	}
}

func normalizeHexDigitCase(text string, mode config.NumericLiteralCase) string {
	if len(text) < 2 || text[0] != '0' || (text[1] != 'x' && text[1] != 'X') {
		return text
	}

	if mode == config.NumericLiteralCaseUpper {
		return text[:2] + strings.ToUpper(text[2:])
	}

	return text[:2] + strings.ToLower(text[2:])
}

func normalizeExponentCase(text string, mode config.NumericLiteralCase) string {
	idx := strings.IndexAny(text, "eE")
	if idx < 0 {
		return text
	}

	letter := "e"
	if mode == config.NumericLiteralCaseUpper {
		letter = "E"
	}

	return text[:idx] + letter + text[idx+1:]
}
