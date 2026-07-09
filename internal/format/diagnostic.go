package format

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/token"
)

func parseDiagnostic(source []byte, parsed *parser.File, subject string) error {
	offset := parseErrorOffset(source, parsed)
	line, column, lineText, marker := sourceLocation(source, offset)
	detail := parseErrorDetail(source, parsed, offset)
	lineNumberWidth := len(strconv.Itoa(line))

	return fmt.Errorf("%s does not parse cleanly at line %d, column %d%s\n%*d | %s\n%s | %s^",
		subject, line, column, detail,
		lineNumberWidth, line, lineText,
		strings.Repeat(" ", lineNumberWidth), marker)
}

func parseErrorOffset(source []byte, parsed *parser.File) int {
	if parsed == nil {
		return 0
	}

	for _, tok := range parsed.Tokens {
		if tok.Kind == token.Unknown {
			return tok.Start.Offset
		}
	}

	if offset, ok := firstErrorNodeOffset(parsed.Root); ok {
		return min(max(offset, 0), len(source))
	}

	if len(parsed.Tokens) > 0 {
		return min(max(parsed.Tokens[len(parsed.Tokens)-1].Start.Offset, 0), len(source))
	}

	return 0
}

func firstErrorNodeOffset(n *parser.Node) (int, bool) {
	if n == nil || !n.HasError {
		return 0, false
	}

	for _, child := range n.Children {
		if offset, ok := firstErrorNodeOffset(child); ok {
			return offset, true
		}
	}

	return n.Start, true
}

func parseErrorDetail(source []byte, parsed *parser.File, offset int) string {
	if parsed == nil {
		return ""
	}

	for _, tok := range parsed.Tokens {
		if tok.Kind == token.EOF && tok.Start.Offset == offset {
			return " at end of file"
		}
		if tok.Start.Offset <= offset && offset < tok.End.Offset {
			text := tok.Text(source)
			if utf8.RuneCountInString(text) > 32 {
				text = string([]rune(text)[:31]) + "…"
			}
			return fmt.Sprintf(" near token %q", text)
		}
	}

	return ""
}

func sourceLocation(source []byte, offset int) (line, column int, lineText, marker string) {
	offset = min(max(offset, 0), len(source))
	line = 1
	lineStart := 0
	for i := 0; i < offset; i++ {
		if source[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	lineEnd := len(source)
	if i := strings.IndexByte(string(source[lineStart:]), '\n'); i >= 0 {
		lineEnd = lineStart + i
	}
	if lineEnd > lineStart && source[lineEnd-1] == '\r' {
		lineEnd--
	}

	prefix := source[lineStart:offset]
	column = utf8.RuneCount(prefix) + 1
	lineText = string(source[lineStart:lineEnd])
	marker = strings.Map(func(r rune) rune {
		if r == '\t' {
			return '\t'
		}
		return ' '
	}, string(prefix))

	return line, column, lineText, marker
}
