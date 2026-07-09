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

		indentDirectiveFreeBracePair(lines, open, i, cfg.IndentWidth, unit)
	}

	return lines
}

func indentDirectiveFreeBracePair(lines []string, open, closeLine, indentWidth int, unit string) {
	openerIndent := sharedIndentColumns(lines[open], indentWidth)
	minimum := -1
	hasDirective := false

	for j := open + 1; j < closeLine; j++ {
		body := strings.TrimSpace(lines[j])
		if body == "" {
			continue
		}

		if strings.HasPrefix(body, "#") {
			hasDirective = true
			break
		}

		indent := sharedIndentColumns(lines[j], indentWidth)
		if minimum < 0 || indent < minimum {
			minimum = indent
		}
	}

	if hasDirective || minimum < 0 || minimum > openerIndent {
		return
	}

	for j := open + 1; j < closeLine; j++ {
		if strings.TrimSpace(lines[j]) != "" {
			lines[j] = unit + lines[j]
		}
	}
}

func expandSharedMultilineControls(lines []string, cfg config.Config) []string {
	if cfg.SingleStatementBraces != config.SingleStatementBracesAlways {
		return lines
	}

	var out []string

	for i := 0; i < len(lines); i++ {
		next, consumed := expandSharedMultilineControlAt(lines, i, cfg)
		out = append(out, next...)
		i += consumed
	}

	return out
}

// expandSharedMultilineControlAt expands the control statement starting at
// lines[i] if it has a brace-worthy single-statement body, returning the
// replacement lines and how many extra input lines were consumed.
func expandSharedMultilineControlAt(lines []string, i int, cfg config.Config) ([]string, int) {
	if !isSharedControlHeaderLine(lines[i]) {
		return []string{lines[i]}, 0
	}

	closeLine, closeOffset := findSharedControlParenClose(lines, i)
	if closeLine <= i {
		return []string{lines[i]}, 0
	}

	body := strings.TrimSpace(lines[closeLine][closeOffset:])

	bodyLine := closeLine
	if body == "" {
		bodyLine++
		if bodyLine >= len(lines) {
			return []string{lines[i]}, 0
		}

		body = strings.TrimSpace(lines[bodyLine])
	}

	if strings.HasPrefix(body, "#") || !sharedCompleteSimpleBody(body) {
		return []string{lines[i]}, 0
	}

	return expandSharedControlBrace(lines, i, closeLine, closeOffset, bodyLine, body, cfg), bodyLine - i
}

func isSharedControlHeaderLine(line string) bool {
	startTokens := lexer.Tokenize([]byte(strings.TrimSpace(line)))

	start := 0
	if len(startTokens) > 1 && startTokens[0].Kind == token.KwElse && startTokens[1].Kind == token.KwIf {
		start = 1
	}

	return start < len(startTokens) && (startTokens[start].Kind == token.KwIf ||
		startTokens[start].Kind == token.KwFor || startTokens[start].Kind == token.KwWhile)
}

// findSharedControlParenClose scans from lines[i] for the closing paren that
// matches the control statement's condition, stopping at any directive line.
func findSharedControlParenClose(lines []string, i int) (closeLine, closeOffset int) {
	depth, opened := 0, false
	closeLine, closeOffset = -1, -1

	for j := i; j < len(lines); j++ {
		if j > i && strings.HasPrefix(strings.TrimSpace(lines[j]), "#") {
			break
		}

		for _, tok := range lexer.Tokenize([]byte(lines[j])) {
			//nolint:exhaustive // only paren depth tokens matter here
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

	return closeLine, closeOffset
}

func expandSharedControlBrace(lines []string, i, closeLine, closeOffset, bodyLine int, body string, cfg config.Config) []string {
	indentText := lines[i][:len(lines[i])-len(strings.TrimLeft(lines[i], " \t"))]

	unit := strings.Repeat(" ", cfg.IndentWidth)
	if cfg.IndentStyle == config.IndentStyleTab {
		unit = "\t"
	}

	out := append([]string(nil), lines[i:closeLine+1]...)
	if bodyLine == closeLine {
		out[len(out)-1] = strings.TrimRight(lines[closeLine][:closeOffset], " \t")
	}

	switch cfg.BraceStyle {
	case config.BraceStyle1TBS:
		out[len(out)-1] += " {"
		out = append(out, indentText+unit+body, indentText+"}")
	case config.BraceStyleWhitesmiths:
		out = append(out, indentText+unit+"{", indentText+unit+unit+body, indentText+unit+"}")
	case config.BraceStyleAllman:
		out = append(out, indentText+"{", indentText+unit+body, indentText+"}")
	default:
		out = append(out, indentText+"{", indentText+unit+body, indentText+"}")
	}

	return out
}
