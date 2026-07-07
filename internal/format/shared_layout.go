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
	if bracesBalanced(prefix.Text(s.source)) {
		parts = append(parts, doc.Indent(s.joinBraceStyle(doc.Text("{"))))
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
		parts = append(parts, doc.Indent(doc.Indent(doc.Concat(items...))))
	}
	parts = append(parts, doc.Indent(doc.Concat(doc.HardLine(), doc.Text("}"))))
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

func bracesBalanced(text string) bool {
	depth := 0
	for _, tok := range lexer.Tokenize([]byte(text)) {
		switch tok.Kind {
		case token.LBrace:
			depth++
		case token.RBrace:
			depth--
		}
	}
	return depth <= 0
}

func (s *state) formatSharedLines(lines []string, baseIndent int) doc.Doc {
	parts := make([]doc.Doc, 0, len(lines)*2)
	previousTrimmed := ""
	previousColumns := 0
	continuationActive := false
	continuationColumns := 0
	for i, line := range lines {
		if i > 0 {
			lineBreak := doc.HardLine()
			if s.config.DirectiveIndent == config.DirectiveIndentNone && strings.HasPrefix(strings.TrimSpace(line), "#") {
				lineBreak = doc.ResetIndent(lineBreak)
			}
			parts = append(parts, lineBreak)
		}
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}
		columns := sharedIndentColumns(line, s.config.IndentWidth)
		continues := i > 0 && !strings.HasPrefix(trimmed, "#") &&
			(sharedLineEndsContinuation(previousTrimmed) || sharedLineStartsContinuation(trimmed))
		if strings.HasPrefix(previousTrimmed, "#") {
			continues = strings.HasSuffix(previousTrimmed, "\\") && sharedLineStartsContinuation(trimmed)
		}
		if continues {
			if !continuationActive {
				continuationColumns = previousColumns + s.continuationIndentWidth()
			}
			if columns < continuationColumns {
				columns = continuationColumns
			}
			continuationActive = true
		} else {
			continuationActive = false
		}
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
		indent := strings.Repeat(" ", level*s.config.IndentWidth)
		if s.config.IndentStyle == config.IndentStyleTab {
			indent = strings.Repeat("\t", level)
		}
		trimmed = normalizeSharedLine(trimmed, s.config)
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
					lineBreak := doc.HardLine()
					if s.config.DirectiveIndent == config.DirectiveIndentNone && strings.HasPrefix(strings.TrimSpace(segment), "#") {
						lineBreak = doc.ResetIndent(lineBreak)
					}
					parts = append(parts, lineBreak)
				}
				parts = append(parts, doc.Text(indent+logicalIndent+segment))
			}
		}
		previousTrimmed = trimmed
		previousColumns = level * s.config.IndentWidth
		if i > 0 {
			previousColumns += baseIndent
		}
	}
	return doc.Concat(parts...)
}
