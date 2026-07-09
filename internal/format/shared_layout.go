package format

import (
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatSharedConditional(n *parser.Node) doc.Doc {
	if prefix, body := n.Field("prefix"), n.Field("body"); prefix != nil && body != nil {
		return s.formatStructuredSharedConditional(prefix, body, n.Field("alternative"))
	}

	text := strings.TrimRight(n.Text(s.source), " \t\r\n")
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines = indentSharedDirectiveFreeBlocks(lines, s.config)
	lines = expandSharedMultilineControls(lines, s.config)

	baseIndent := sharedBaseIndent(s.source, n.Start, s.config.IndentWidth)
	if s.config.DirectiveIndent == config.DirectiveIndentNone {
		baseIndent = sharedMinimumCodeIndent(lines, s.config.IndentWidth)
	}

	return s.formatSharedLines(lines, baseIndent)
}

func (s *state) formatStructuredSharedConditional(prefix, body, alternative *parser.Node) doc.Doc {
	parts := []doc.Doc{s.formatNode(prefix)}

	braceLevel := sharedPrefixBraceLevel(s.source, prefix, s.config.IndentWidth)
	if bracesBalanced(prefix.Text(s.source)) {
		parts = append(parts, indentLevels(s.joinBraceStyle(doc.Text("{")), braceLevel+1))
	}

	if len(body.Children) > 0 {
		var items []doc.Doc

		for i, item := range body.Children {
			if i == 0 {
				items = append(items, s.itemSeparatorBefore(item))
			} else {
				separator := blankLineSeparator(s.blankLinesBefore(item.LeadingTrivia()))
				items = append(items, s.separatorForItem(separator, item))
			}

			items = append(items, s.formatNode(item))
		}

		parts = append(parts, indentLevels(doc.Concat(items...), braceLevel+1))
	}

	parts = append(parts, indentLevels(doc.Concat(doc.HardLine(), doc.Text("}")), braceLevel))
	if alternative != nil {
		parts = append(parts, s.blockTrailingKeyword("else"))
		if alternative.Kind == parser.KindIfStatement {
			parts = append(parts, s.chainContinuation(alternative))
		} else {
			parts = append(parts, s.formatBranchBody(alternative))
		}
	}

	return doc.Concat(parts...)
}

func indentLevels(content doc.Doc, levels int) doc.Doc {
	for range levels {
		content = doc.Indent(content)
	}

	return content
}

func bracesBalanced(text string) bool {
	depth := 0

	for _, tok := range lexer.Tokenize([]byte(text)) {
		//nolint:exhaustive // only brace depth tokens matter here
		switch tok.Kind {
		case token.LBrace:
			depth++
		case token.RBrace:
			depth--
		}
	}

	return depth <= 0
}

// sharedLineContinuation tracks the continuation state threaded across lines
// in formatSharedLines: whether the previous line implies the current one
// continues it, and at what column continued lines should sit.
type sharedLineContinuation struct {
	previousTrimmed string
	previousColumns int
	active          bool
	columns         int
}

func (s *state) formatSharedLines(lines []string, baseIndent int) doc.Doc {
	parts := make([]doc.Doc, 0, len(lines)*2)

	var cont sharedLineContinuation

	for i, line := range lines {
		if i > 0 {
			parts = append(parts, s.sharedLineBreak(line))
		}

		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}

		level := cont.resolveLevel(s, line, trimmed, i, baseIndent)

		trimmed = normalizeSharedLine(trimmed, s.config)
		parts = append(parts, s.formatSharedLogicalLines(trimmed, level, baseIndent)...)

		cont.previousTrimmed = trimmed

		cont.previousColumns = level * s.config.IndentWidth
		if i > 0 {
			cont.previousColumns += baseIndent
		}
	}

	return doc.Concat(parts...)
}

func (s *state) sharedLineBreak(line string) doc.Doc {
	lineBreak := doc.HardLine()
	if s.config.DirectiveIndent == config.DirectiveIndentNone && strings.HasPrefix(strings.TrimSpace(line), "#") {
		lineBreak = doc.ResetIndent(lineBreak)
	}

	return lineBreak
}

// resolveColumns computes the raw (pre-baseIndent) column for the current
// line, updating whether a continuation is active in cont.
func (cont *sharedLineContinuation) resolveColumns(s *state, line, trimmed string, i int) int {
	columns := sharedIndentColumns(line, s.config.IndentWidth)

	continues := i > 0 && !strings.HasPrefix(trimmed, "#") &&
		(sharedLineEndsContinuation(cont.previousTrimmed) || sharedLineStartsContinuation(trimmed))
	if strings.HasPrefix(cont.previousTrimmed, "#") {
		continues = strings.HasSuffix(cont.previousTrimmed, "\\") && sharedLineStartsContinuation(trimmed)
	}

	if !continues {
		cont.active = false
		return columns
	}

	if !cont.active {
		cont.columns = cont.previousColumns + s.continuationIndentWidth()
	}

	if columns < cont.columns {
		columns = cont.columns
	}

	cont.active = true

	return columns
}

// resolveLevel computes the indent level for the current line, updating the
// continuation state in place, and returns the level in indent-width units.
func (cont *sharedLineContinuation) resolveLevel(s *state, line, trimmed string, i, baseIndent int) int {
	columns := cont.resolveColumns(s, line, trimmed, i)
	if i > 0 {
		columns -= baseIndent
	}

	if s.config.DirectiveIndent == config.DirectiveIndentNone && strings.HasPrefix(trimmed, "#") {
		columns = 0
	}

	if columns < 0 {
		columns = 0
	}

	level := columns / s.config.IndentWidth
	if columns%s.config.IndentWidth != 0 {
		level++
	}

	return level
}

func (s *state) formatSharedLogicalLines(trimmed string, level, baseIndent int) []doc.Doc {
	indent := strings.Repeat(" ", level*s.config.IndentWidth)
	if s.config.IndentStyle == config.IndentStyleTab {
		indent = strings.Repeat("\t", level)
	}

	var parts []doc.Doc

	logicalLines := expandSharedSimpleControl(trimmed, s.config)
	for li, logicalLine := range logicalLines {
		if li > 0 {
			parts = append(parts, doc.HardLine())
		}

		logicalCode := strings.TrimLeft(logicalLine, " \t")
		logicalIndent := logicalLine[:len(logicalLine)-len(logicalCode)]
		logicalIndentColumns := sharedIndentColumns(logicalLine, s.config.IndentWidth)
		indentColumns := level*s.config.IndentWidth + logicalIndentColumns

		wrapped := wrapSharedLine(logicalCode, s.config.LineWidth-baseIndent-indentColumns, s.config.IndentWidth, s.continuationIndentWidth(), s.config.IndentStyle == config.IndentStyleTab)
		for wi, segment := range wrapped {
			if wi > 0 {
				parts = append(parts, s.sharedLineBreak(segment))
			}

			parts = append(parts, doc.Text(indent+logicalIndent+segment))
		}
	}

	return parts
}
