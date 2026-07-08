package format

import (
	"strings"

	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/config"
)

func indentSharedDirectiveFreeBlocks(lines []string, cfg config.Config) []string {
	lines = append([]string(nil), lines...)

	var stack []int

	unit := strings.Repeat(" ", cfg.IndentWidth)
	if cfg.IndentStyle == config.IndentStyleTab {
		unit = "\t"
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "{" {
			stack = append(stack, i)
			continue
		}

		if trimmed != "}" && trimmed != "};" || len(stack) == 0 {
			continue
		}

		open := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		openerIndent := sharedIndentColumns(lines[open], cfg.IndentWidth)
		minimum := -1
		hasDirective := false

		for j := open + 1; j < i; j++ {
			body := strings.TrimSpace(lines[j])
			if body == "" {
				continue
			}

			if strings.HasPrefix(body, "#") {
				hasDirective = true
				break
			}

			indent := sharedIndentColumns(lines[j], cfg.IndentWidth)
			if minimum < 0 || indent < minimum {
				minimum = indent
			}
		}

		if hasDirective || minimum < 0 || minimum > openerIndent {
			continue
		}

		for j := open + 1; j < i; j++ {
			if strings.TrimSpace(lines[j]) != "" {
				lines[j] = unit + lines[j]
			}
		}
	}

	return lines
}

func expandSharedMultilineControls(lines []string, cfg config.Config) []string {
	if cfg.SingleStatementBraces != config.SingleStatementBracesAlways {
		return lines
	}

	var out []string

	for i := 0; i < len(lines); i++ {
		startTokens := lexer.Tokenize([]byte(strings.TrimSpace(lines[i])))

		start := 0
		if len(startTokens) > 1 && startTokens[0].Kind == token.KwElse && startTokens[1].Kind == token.KwIf {
			start = 1
		}

		if start >= len(startTokens) || (startTokens[start].Kind != token.KwIf &&
			startTokens[start].Kind != token.KwFor && startTokens[start].Kind != token.KwWhile) {
			out = append(out, lines[i])
			continue
		}

		depth, opened, closeLine, closeOffset := 0, false, -1, -1

		for j := i; j < len(lines); j++ {
			if j > i && strings.HasPrefix(strings.TrimSpace(lines[j]), "#") {
				break
			}

			for _, tok := range lexer.Tokenize([]byte(lines[j])) {
				switch tok.Kind {
				case token.LParen:
					depth++
					opened = true
				case token.RParen:
					if opened {
						depth--
						if depth == 0 {
							closeLine, closeOffset = j, tok.End.Offset
						}
					}
				}

				if closeLine >= 0 {
					break
				}
			}

			if closeLine >= 0 {
				break
			}
		}

		if closeLine <= i {
			out = append(out, lines[i])
			continue
		}

		body := strings.TrimSpace(lines[closeLine][closeOffset:])

		bodyLine := closeLine
		if body == "" {
			bodyLine++
			if bodyLine >= len(lines) {
				out = append(out, lines[i])
				continue
			}

			body = strings.TrimSpace(lines[bodyLine])
		}

		if strings.HasPrefix(body, "#") || !sharedCompleteSimpleBody(body) {
			out = append(out, lines[i])
			continue
		}

		indentText := lines[i][:len(lines[i])-len(strings.TrimLeft(lines[i], " \t"))]

		unit := strings.Repeat(" ", cfg.IndentWidth)
		if cfg.IndentStyle == config.IndentStyleTab {
			unit = "\t"
		}

		out = append(out, lines[i:closeLine+1]...)
		if bodyLine == closeLine {
			out[len(out)-1] = strings.TrimRight(lines[closeLine][:closeOffset], " \t")
		}

		switch cfg.BraceStyle {
		case config.BraceStyle1TBS:
			out[len(out)-1] += " {"
			out = append(out, indentText+unit+body, indentText+"}")
		case config.BraceStyleWhitesmiths:
			out = append(out, indentText+unit+"{", indentText+unit+unit+body, indentText+unit+"}")
		default:
			out = append(out, indentText+"{", indentText+unit+body, indentText+"}")
		}

		i = bodyLine
	}

	return out
}
