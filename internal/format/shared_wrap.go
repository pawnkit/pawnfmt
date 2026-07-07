package format

import (
	"strings"
	"unicode/utf8"

	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func wrapSharedLine(line string, width, indentWidth, continuationWidth int, useTabs bool) []string {
	if utf8.RuneCountInString(line) <= width || width < 20 || strings.HasPrefix(strings.TrimSpace(line), "#") || strings.HasSuffix(strings.TrimSpace(line), "\\") {
		return []string{line}
	}
	tokens := lexer.Tokenize([]byte(line))
	candidates, controlEnd := sharedWrapCandidates(line, tokens)
	if controlEnd > 0 {
		suffix := strings.TrimSpace(line[controlEnd:])
		if suffix != "" && suffix != "{" {
			return wrapSharedControlStatement(line, tokens, controlEnd, width, indentWidth, continuationWidth, useTabs)
		}
	}
	if len(candidates) == 0 {
		return []string{line}
	}
	continuation, continuationColumns := sharedContinuationIndent(continuationWidth, useTabs)
	return wrapSharedAtCandidates(line, candidates, width, continuationColumns, continuation)
}

func sharedWrapCandidates(line string, tokens []token.Token) ([]int, int) {
	var candidates []int
	parenDepth := 0
	ternaryDepth := 0
	controlHeader := strings.HasPrefix(strings.TrimSpace(line), "if (") || strings.HasPrefix(strings.TrimSpace(line), "else if (")
	controlEnd := -1
	for i, tok := range tokens {
		switch tok.Kind {
		case token.Comma:
			candidates = append(candidates, tok.End.Offset)
		case token.Assign, token.PlusAssign, token.MinusAssign, token.StarAssign, token.SlashAssign,
			token.PercentAssign:
			candidates = append(candidates, tok.End.Offset)
		case token.AndAnd, token.OrOr, token.Plus, token.Minus,
			token.Star, token.Slash, token.Percent,
			token.Eq, token.NotEq, token.Lt, token.Gt, token.LtEq, token.GtEq:
			candidates = append(candidates, tok.Start.Offset)
		case token.Question:
			ternaryDepth++
			candidates = append(candidates, tok.Start.Offset)
		case token.Colon:
			if ternaryDepth > 0 {
				ternaryDepth--
				candidates = append(candidates, tok.Start.Offset)
			}
		case token.LParen:
			parenDepth++
		case token.RParen:
			parenDepth--
			if controlHeader && parenDepth == 0 && tok.End.Offset < len(line) {
				candidates = append(candidates, tok.End.Offset)
				controlEnd = tok.End.Offset
				controlHeader = false
			} else if i+1 < len(tokens) && tokens[i+1].Kind != token.EOF &&
				tok.End.Offset < tokens[i+1].Start.Offset {
				candidates = append(candidates, tok.End.Offset)
			}
		}
	}
	return candidates, controlEnd
}

func sharedContinuationIndent(indentWidth int, useTabs bool) (string, int) {
	continuation := strings.Repeat(" ", indentWidth)
	if useTabs {
		return "\t", indentWidth
	}
	return continuation, len(continuation)
}

func wrapSharedAtCandidates(line string, candidates []int, width, continuationColumns int, continuation string) []string {
	contentWidth := width - continuationColumns
	if contentWidth < 20 {
		contentWidth = width
	}
	var out []string
	start := 0
	for utf8.RuneCountInString(line[start:]) > contentWidth {
		choice := -1
		for _, candidate := range candidates {
			if candidate <= start {
				continue
			}
			if utf8.RuneCountInString(line[start:candidate]) <= contentWidth {
				choice = candidate
				continue
			}
			if choice < 0 {
				choice = candidate
			}
			break
		}
		if choice <= start || choice >= len(line) {
			break
		}
		out = append(out, strings.TrimSpace(line[start:choice]))
		start = choice
		for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
			start++
		}
	}
	if len(out) == 0 {
		return []string{line}
	}
	out = append(out, strings.TrimSpace(line[start:]))
	for i := 1; i < len(out); i++ {
		out[i] = continuation + out[i]
	}
	return out
}

func wrapSharedControlStatement(line string, tokens []token.Token, controlEnd, width, indentWidth, continuationWidth int, useTabs bool) []string {
	header := strings.TrimSpace(line[:controlEnd])
	suffixEnd := len(line)
	elseStart := -1
	for _, tok := range tokens {
		if tok.Start.Offset >= controlEnd && tok.Kind == token.KwElse {
			elseStart = tok.Start.Offset
			suffixEnd = elseStart
			break
		}
	}
	result := wrapSharedLine(header, width, indentWidth, continuationWidth, useTabs)
	continuation := strings.Repeat(" ", indentWidth)
	if useTabs {
		continuation = "\t"
	}
	body := strings.TrimSpace(line[controlEnd:suffixEnd])
	if body != "" {
		bodyLines := wrapSharedLine(body, width-indentWidth, indentWidth, continuationWidth, useTabs)
		for _, bodyLine := range bodyLines {
			result = append(result, continuation+bodyLine)
		}
	}
	if elseStart >= 0 {
		elseLine := strings.TrimSpace(line[elseStart:])
		result = append(result, wrapSharedLine(elseLine, width, indentWidth, continuationWidth, useTabs)...)
	}
	return result
}

func (s *state) formatConditionalFunctionDefinition(n *parser.Node) doc.Doc {
	headers := n.Field("headers")
	body := n.Field("body")
	if headers == nil || body == nil {
		return s.raw(n)
	}
	bodyDoc := doc.Concat(doc.HardLine(), s.formatNode(body))
	if s.config.BraceStyle == config.BraceStyleWhitesmiths {
		bodyDoc = doc.Indent(bodyDoc)
	}
	return doc.Concat(s.formatNode(headers), bodyDoc)
}

func sharedBaseIndent(source []byte, start, width int) int {
	lineStart := start
	for lineStart > 0 && source[lineStart-1] != '\n' && source[lineStart-1] != '\r' {
		lineStart--
	}
	return sharedIndentColumns(string(source[lineStart:start]), width)
}

func sharedMinimumCodeIndent(lines []string, width int) int {
	minimum := -1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		columns := sharedIndentColumns(line, width)
		if minimum < 0 || columns < minimum {
			minimum = columns
		}
	}
	if minimum < 0 {
		return 0
	}
	return minimum
}

func sharedIndentColumns(line string, width int) int {
	columns := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			columns++
		case '\t':
			columns += width
		default:
			return columns
		}
	}
	return columns
}
