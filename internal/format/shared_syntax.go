package format

import (
	"slices"
	"strings"

	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawn-parser/token"
)

func sharedExpressionDirective(line string) bool {
	for _, keyword := range []string{"if", "elseif", "assert"} {
		prefix := "#" + keyword
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		if len(line) == len(prefix) || !isIdentByte(line[len(prefix)]) {
			return true
		}
	}

	return false
}

func horizontalGap(gap string) bool {
	for _, ch := range gap {
		if ch != ' ' && ch != '\t' {
			return false
		}
	}

	return true
}

func sharedTagName(name string) bool {
	if name == "_" || name == "bool" || name == "Float" || name == "File" {
		return true
	}

	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

func sharedKeywordNeedsSpace(current, next token.Kind) bool {
	if next == token.LParen && (current == token.KwSizeof || current == token.KwTagof || current == token.KwDefined) {
		return false
	}

	//nolint:exhaustive // only the relevant token kinds matter here
	switch current {
	case token.KwPublic, token.KwStock, token.KwStatic, token.KwNative,
		token.KwForward, token.KwConst, token.KwNew, token.KwDecl,
		token.KwReturn, token.KwGoto, token.KwState,
		token.KwSizeof, token.KwTagof, token.KwDefined,
		token.KwCase, token.KwElse:
		return true
	default:
		return false
	}
}

func sharedStringLiteral(kind token.Kind) bool {
	return kind == token.StringLiteral || kind == token.PackedString
}

func sharedForOpeningParen(tokens []token.Token, i int) bool {
	return i > 0 && i < len(tokens) && tokens[i].Kind == token.LParen && tokens[i-1].Kind == token.KwFor
}

func sharedForClosingParen(tokens []token.Token, i int) bool {
	if i < 0 || i >= len(tokens) || tokens[i].Kind != token.RParen {
		return false
	}

	depth := 0

	for j := i; j >= 0; j-- {
		//nolint:exhaustive // only the relevant token kinds matter here
		switch tokens[j].Kind {
		case token.RParen:
			depth++
		case token.LParen:
			depth--
			if depth == 0 {
				return sharedForOpeningParen(tokens, j)
			}
		}
	}

	return false
}

func sharedTightKeywordOpeningParen(tokens []token.Token, i int) bool {
	if i <= 0 || i >= len(tokens) || tokens[i].Kind != token.LParen {
		return false
	}

	//nolint:exhaustive // only the relevant token kinds matter here
	switch tokens[i-1].Kind {
	case token.KwSizeof, token.KwTagof, token.KwDefined:
		return true
	default:
		return false
	}
}

func sharedTightKeywordClosingParen(tokens []token.Token, i int) bool {
	if i < 0 || i >= len(tokens) || tokens[i].Kind != token.RParen {
		return false
	}

	depth := 0

	for j := i; j >= 0; j-- {
		//nolint:exhaustive // only the relevant token kinds matter here
		switch tokens[j].Kind {
		case token.RParen:
			depth++
		case token.LParen:
			depth--
			if depth == 0 {
				return sharedTightKeywordOpeningParen(tokens, j)
			}
		}
	}

	return false
}

func sharedDeclarationTagColon(tokens []token.Token, i int) bool {
	if i < 1 || i+1 >= len(tokens) || tokens[i-1].Kind != token.Identifier || tokens[i+1].Kind != token.Identifier {
		return false
	}

	if i == 1 {
		for _, tok := range tokens[i+2:] {
			if tok.Kind == token.Comma || tok.Kind == token.RParen {
				return true
			}
		}

		return false
	}

	//nolint:exhaustive // only the relevant token kinds matter here
	switch tokens[i-2].Kind {
	case token.KwPublic, token.KwStock, token.KwStatic, token.KwNative,
		token.KwForward, token.KwConst, token.KwNew,
		token.LParen, token.Comma, token.Amp:
		return true
	default:
		return false
	}
}

func sharedTernaryColon(tokens []token.Token, i int) bool {
	depth := 0

	for j := 0; j <= i && j < len(tokens); j++ {
		//nolint:exhaustive // only the relevant token kinds matter here
		switch tokens[j].Kind {
		case token.Question:
			depth++
		case token.Colon:
			if j == i {
				return depth > 0
			}

			if depth > 0 {
				depth--
			}
		}
	}

	return false
}

func sharedCaseOrLabelColon(tokens []token.Token, i int, source []byte) bool {
	if i <= 0 || i >= len(tokens) || tokens[i].Kind != token.Colon || sharedTernaryColon(tokens, i) {
		return false
	}

	if tokens[0].Kind == token.KwCase || tokens[0].Kind == token.KwDefault {
		return true
	}

	if i != 1 || tokens[0].Kind != token.Identifier {
		return false
	}

	if i+1 < len(tokens) && tokens[i+1].Kind != token.EOF &&
		sharedTagName(tokens[0].Text(source)) {
		return false
	}

	return true
}

func sharedDeclarationLine(tokens []token.Token) bool {
	for _, tok := range tokens {
		if tok.Kind == token.EOF {
			return false
		}

		//nolint:exhaustive // only the relevant token kinds matter here
		switch tok.Kind {
		case token.KwPublic, token.KwStock, token.KwStatic, token.KwNative,
			token.KwForward, token.KwConst, token.KwNew:
			return true
		case token.Identifier:
			return len(tokens) > 1 && tokens[1].Kind == token.Identifier
		default:
			return false
		}
	}

	return false
}

func sharedBeforeInitializer(tokens []token.Token, i int) bool {
	depth := 0
	initializerDepth := -1

	for j := 0; j < i && j < len(tokens); j++ {
		//nolint:exhaustive // only the relevant token kinds matter here
		switch tokens[j].Kind {
		case token.RParen, token.RBracket, token.RBrace:
			if depth > 0 {
				depth--
			}
		case token.Comma:
			if initializerDepth == depth {
				initializerDepth = -1
			}
		case token.Assign, token.PlusAssign, token.MinusAssign, token.StarAssign,
			token.SlashAssign, token.PercentAssign, token.ShlAssign, token.ShrAssign,
			token.UshrAssign, token.AndAssign, token.OrAssign, token.XorAssign:
			if initializerDepth < 0 {
				initializerDepth = depth
			}
		case token.Semicolon:
			initializerDepth = -1
		case token.LParen, token.LBracket, token.LBrace:
			depth++
		}
	}

	return initializerDepth < 0
}

func sharedArrayBraceTokens(tokens []token.Token) map[int]bool {
	type brace struct {
		index int
		array bool
	}

	var stack []brace

	result := make(map[int]bool)

	for i, tok := range tokens {
		//nolint:exhaustive // only the relevant token kinds matter here
		switch tok.Kind {
		case token.LBrace:
			array := false

			if i > 0 {
				//nolint:exhaustive // only the relevant token kinds matter here
				switch tokens[i-1].Kind {
				case token.Assign, token.Comma, token.LParen:
					array = true
				case token.LBrace:
					array = result[i-1]
				}
			}

			result[i] = array
			stack = append(stack, brace{index: i, array: array})
		case token.RBrace:
			if len(stack) == 0 {
				continue
			}

			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if open.array {
				result[i] = true
			}
		}
	}

	return result
}

func sharedBinaryOperator(tokens []token.Token, i int) bool {
	if i < 0 || i >= len(tokens) {
		return false
	}

	//nolint:exhaustive // only the relevant token kinds matter here
	switch tokens[i].Kind {
	case token.Assign, token.PlusAssign, token.MinusAssign, token.StarAssign, token.SlashAssign,
		token.PercentAssign, token.ShlAssign, token.ShrAssign, token.UshrAssign,
		token.AndAssign, token.OrAssign, token.XorAssign,
		token.Eq, token.NotEq, token.Lt, token.Gt, token.LtEq, token.GtEq,
		token.AndAnd, token.OrOr, token.Pipe, token.Caret,
		token.Shl, token.Shr, token.Ushr:
		return true
	case token.Plus, token.Minus, token.Star, token.Slash, token.Percent, token.Amp:
		return i > 0 && sharedTokenEndsOperand(tokens[i-1].Kind)
	default:
		return false
	}
}

func sharedPrefixOperator(tokens []token.Token, i int) bool {
	if i < 0 || i >= len(tokens) {
		return false
	}

	//nolint:exhaustive // only the relevant token kinds matter here
	switch tokens[i].Kind {
	case token.Bang, token.Tilde:
		return true
	case token.PlusPlus, token.MinusMinus:
		return i == 0 || !sharedTokenEndsOperand(tokens[i-1].Kind)
	case token.Plus, token.Minus, token.Star, token.Amp:
		return i > 0 && !sharedTokenEndsOperand(tokens[i-1].Kind)
	default:
		return false
	}
}

func sharedPostfixOperator(tokens []token.Token, i int) bool {
	if i <= 0 || i >= len(tokens) {
		return false
	}

	return (tokens[i].Kind == token.PlusPlus || tokens[i].Kind == token.MinusMinus) &&
		sharedTokenEndsOperand(tokens[i-1].Kind)
}

func sharedTokenEndsOperand(kind token.Kind) bool {
	//nolint:exhaustive // only the relevant token kinds matter here
	switch kind {
	case token.Identifier,
		token.IntLiteral, token.FloatLiteral, token.CharLiteral, token.StringLiteral,
		token.PackedString, token.MacroParam, token.KwNull,
		token.RParen, token.RBracket, token.RBrace,
		token.PlusPlus, token.MinusMinus:
		return true
	default:
		return false
	}
}

func normalizeSharedComments(line string) string {
	tokens := lexer.Tokenize([]byte(line))
	seen := make(map[int]bool)

	var offsets []int

	collect := func(items []token.Trivia) {
		for _, item := range items {
			if item.Kind != token.Comment || seen[item.Start.Offset] {
				continue
			}

			seen[item.Start.Offset] = true

			start := item.Start.Offset
			if start+2 < len(line) && line[start:start+2] == "//" {
				ch := line[start+2]
				if ch != ' ' && ch != '\t' && ch != '/' && ch != '!' && ch != '#' {
					offsets = append(offsets, start+2)
				}
			}
		}
	}
	for _, tok := range tokens {
		collect(tok.LeadingTrivia)
		collect(tok.TrailingTrivia)
	}

	for _, v := range slices.Backward(offsets) {
		at := v
		line = line[:at] + " " + line[at:]
	}

	return line
}
