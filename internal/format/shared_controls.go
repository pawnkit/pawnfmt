package format

import (
	"slices"
	"strings"

	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/config"
)

func expandSharedSimpleControl(line string, cfg config.Config) []string {
	tokens := lexer.Tokenize([]byte(line))
	if cfg.BraceStyle == config.BraceStyleAllman && len(tokens) >= 2 &&
		tokens[0].Kind == token.KwDo && strings.TrimSpace(line[tokens[0].End.Offset:]) == "{" {
		return []string{"do", "{"}
	}

	if len(tokens) < 4 {
		return []string{line}
	}

	if tokens[0].Kind == token.KwElse && tokens[1].Kind != token.KwIf {
		body := strings.TrimSpace(line[tokens[0].End.Offset:])
		return expandSharedControlBody(line, "else", body, cfg)
	}

	closeOffset := sharedControlHeaderEnd(line, tokens)
	if closeOffset < 0 {
		return []string{line}
	}

	header := strings.TrimSpace(line[:closeOffset])
	body := strings.TrimSpace(line[closeOffset:])

	return expandSharedControlBody(line, header, body, cfg)
}

func sharedControlHeaderEnd(line string, tokens []token.Token) int {
	//nolint:exhaustive // only control-keyword token kinds matter here
	switch tokens[0].Kind {
	case token.KwIf, token.KwFor, token.KwWhile, token.KwSwitch:
	case token.KwElse:
		if tokens[1].Kind != token.KwIf {
			return -1
		}
	case token.Identifier:
		if tokens[0].Text([]byte(line)) != "foreach" {
			return -1
		}
	default:
		return -1
	}

	open := -1

	for i, tok := range tokens {
		if tok.Kind == token.LParen {
			open = i
			break
		}
	}

	if open < 0 {
		return -1
	}

	depth := 0

	for _, tok := range tokens[open:] {
		//nolint:exhaustive // only paren depth tokens matter here
		switch tok.Kind {
		case token.LParen:
			depth++
		case token.RParen:
			depth--
			if depth == 0 {
				return tok.End.Offset
			}
		}
	}

	return -1
}

func expandSharedControlBody(original, header, body string, cfg config.Config) []string {
	if body == "{" && cfg.BraceStyle == config.BraceStyleAllman {
		return []string{header, "{"}
	}

	if !sharedCompleteSimpleBody(body) {
		return []string{original}
	}

	if cfg.SingleStatementBraces == config.SingleStatementBracesAlways {
		return sharedBracedControl(header, body, cfg)
	}

	if !cfg.KeepSimpleStatementsSingleLine {
		return sharedSplitSimpleControl(header, body, cfg)
	}

	return []string{original}
}

func sharedSplitSimpleControl(header, body string, cfg config.Config) []string {
	unit := strings.Repeat(" ", cfg.IndentWidth)
	if cfg.IndentStyle == config.IndentStyleTab {
		unit = "\t"
	}

	return []string{header, unit + body}
}

func sharedCompleteSimpleBody(body string) bool {
	return body != "" && body != ";" && body != "{" &&
		strings.HasSuffix(body, ";") && !strings.Contains(body, " else ")
}

func sharedBracedControl(header, body string, cfg config.Config) []string {
	unit := strings.Repeat(" ", cfg.IndentWidth)
	if cfg.IndentStyle == config.IndentStyleTab {
		unit = "\t"
	}

	switch cfg.BraceStyle {
	case config.BraceStyle1TBS:
		return []string{header + " {", unit + body, "}"}
	case config.BraceStyleWhitesmiths:
		return []string{header, unit + "{", unit + unit + body, unit + "}"}
	case config.BraceStyleAllman:
		return []string{header, "{", unit + body, "}"}
	default:
		return []string{header, "{", unit + body, "}"}
	}
}

func sharedLineStartsContinuation(line string) bool {
	tokens := lexer.Tokenize([]byte(line))
	if len(tokens) == 0 {
		return false
	}

	return sharedContinuationOperator(tokens[0].Kind)
}

func sharedLineEndsContinuation(line string) bool {
	tokens := lexer.Tokenize([]byte(line))
	for _, v := range slices.Backward(tokens) {
		if v.Kind == token.EOF {
			continue
		}

		return v.Kind == token.Comma || sharedContinuationOperator(v.Kind)
	}

	return false
}

func sharedContinuationOperator(kind token.Kind) bool {
	//nolint:exhaustive // only operator token kinds that continue a line matter here
	switch kind {
	case token.Assign, token.PlusAssign, token.MinusAssign, token.StarAssign, token.SlashAssign,
		token.PercentAssign, token.ShlAssign, token.ShrAssign, token.UshrAssign,
		token.AndAssign, token.OrAssign, token.XorAssign,
		token.Plus, token.Minus, token.Star, token.Slash, token.Percent,
		token.Eq, token.NotEq, token.Lt, token.Gt, token.LtEq, token.GtEq,
		token.Shl, token.Shr, token.Ushr, token.AndAnd, token.OrOr,
		token.Amp, token.Pipe, token.Caret, token.Question:
		return true
	default:
		return false
	}
}
