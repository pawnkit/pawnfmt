// Package format renders a parsed Pawn source file to formatted output.
package format

import (
	"maps"
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func commentText(t token.Trivia, source []byte) string {
	text := t.Text(source)
	if t.Kind == token.Comment && strings.HasPrefix(text, "//") {
		text = strings.TrimRight(text, " \t")
		if len(text) > 2 && text[2] != ' ' && text[2] != '\t' && text[2] != '/' &&
			text[2] != '!' && text[2] != '#' {
			text = "// " + text[2:]
		}

		return text
	}

	return text
}

func (s *state) leadingDocs(lead []token.Trivia) []doc.Doc {
	var parts []doc.Doc

	i := 0
	for i < len(lead) {
		t := lead[i]
		if t.Kind != token.Comment {
			i++
			continue
		}

		followedByNewline := false

		for j := i + 1; j < len(lead); j++ {
			if lead[j].Kind == token.Newline {
				followedByNewline = true
				break
			}

			if lead[j].Kind == token.Comment {
				break
			}
		}

		i++

		if !s.claimComment(t) {
			continue
		}

		text := commentText(t, s.source)
		if strings.HasPrefix(text, "//") || followedByNewline || i == len(lead) {
			parts = append(parts, doc.RawTextBlock(text), doc.HardLine())
		} else {
			parts = append(parts, doc.Text(text), doc.Text(" "))
		}
	}

	return parts
}

func (s *state) claimComment(t token.Trivia) bool {
	if s.renderedComments[t.Start.Offset] {
		return false
	}

	s.renderedComments[t.Start.Offset] = true

	return true
}

func (s *state) trailingDoc(trail []token.Trivia) doc.Doc {
	t, ok := firstTrailingComment(trail)
	if !ok {
		return nil
	}

	if !s.claimComment(t) {
		return nil
	}

	pad := ""
	if n := s.commentPadWidths[t.Start.Offset]; n > 0 {
		pad = strings.Repeat(" ", n)
	}

	text := commentText(t, s.source)
	if strings.HasPrefix(text, "//") {
		return doc.Concat(doc.LineSuffix(doc.Concat(doc.Text(" "+pad), doc.Text(text))), doc.BreakParent())
	}

	return doc.Concat(doc.Text(" "+pad), doc.Text(text))
}

func firstTrailingComment(trail []token.Trivia) (token.Trivia, bool) {
	for _, t := range trail {
		if t.Kind == token.Comment {
			return t, true
		}

		if t.Kind == token.Newline {
			break
		}
	}

	return token.Trivia{}, false
}

func (s *state) applyCommentAlignment(items []*parser.Node) {
	maps.Copy(s.commentPadWidths, s.commentAlignmentWidths(items))
}

func (s *state) commentAlignmentWidths(items []*parser.Node) map[int]int {
	if !s.config.AlignTrailingComments {
		return nil
	}

	widths := make(map[int]int)

	var (
		offsets    []int
		coreWidths []int
	)

	flush := func() {
		if len(offsets) > 1 {
			maxWidth := 0
			for _, w := range coreWidths {
				if w > maxWidth {
					maxWidth = w
				}
			}

			for i, offset := range offsets {
				if pad := maxWidth - coreWidths[i]; pad > 0 {
					widths[offset] = pad
				}
			}
		}

		offsets = nil
		coreWidths = nil
	}

	for i, item := range items {
		if i > 0 && s.blankLinesBefore(item.LeadingTrivia()) > 0 {
			flush()
		}

		t, ok := firstTrailingComment(item.TrailingTrivia())
		if !ok {
			flush()
			continue
		}

		full := s.measureFlat(item)

		suffix := " " + commentText(t, s.source)
		if strings.Contains(full, "\n") || !strings.HasSuffix(full, suffix) {
			flush()
			continue
		}

		offsets = append(offsets, t.Start.Offset)
		coreWidths = append(coreWidths, len(full)-len(suffix))
	}

	flush()

	return widths
}

func (s *state) blankLinesBefore(lead []token.Trivia) int {
	count := 0
	run := 0
	afterComment := false
	flush := func() {
		if afterComment {
			if run > 1 {
				count += run - 1
			}
		} else {
			count += run
		}

		run = 0
	}

	for _, t := range lead {
		//nolint:exhaustive // only blank/comment trivia kinds matter here
		switch t.Kind {
		case token.Newline:
			run++
		case token.Comment:
			flush()

			afterComment = true
		}
	}

	flush()

	if !s.config.CollapseBlankLines {
		return count
	}

	if limit := s.config.MaxBlankLines; count > limit {
		return limit
	}

	return count
}

func leadingStartsNewLine(lead []token.Trivia) bool {
	for _, t := range lead {
		if t.Kind == token.Newline {
			return true
		}
	}

	return false
}

func (s *state) itemSeparatorBefore(item *parser.Node) doc.Doc {
	return s.directiveAwareSeparator(doc.HardLine(), item)
}

func (s *state) directiveAwareSeparator(separator doc.Doc, item *parser.Node) doc.Doc {
	if item != nil && item.Kind == parser.KindLabelStatement && !s.config.IndentGotoLabels {
		return doc.Outdent(separator)
	}

	if !firstLineIsDirective(item) {
		return separator
	}

	if s.config.DirectiveIndent == config.DirectiveIndentNone {
		return doc.ResetIndent(separator)
	}

	if !alignsWithEnclosingBrace(item) {
		return separator
	}

	return doc.Outdent(separator)
}

func alignsWithEnclosingBrace(item *parser.Node) bool {
	//nolint:exhaustive // only kinds that pull back to the enclosing brace matter here
	switch item.Kind {
	case parser.KindSharedConditional, parser.KindSharedConditionalPrefix,
		parser.KindConditionalSplice, parser.KindConditionalFunction,
		parser.KindDirectiveEmit:
		return false
	default:
		return true
	}
}

func firstLineIsDirective(item *parser.Node) bool {
	if item == nil {
		return false
	}

	if item.Kind.IsDirective() {
		return true
	}

	if item.Kind == parser.KindSharedConditional || item.Kind == parser.KindSharedConditionalPrefix ||
		item.Kind == parser.KindConditionalFunction {
		return true
	}

	if item.Kind == parser.KindConditionalRegion && len(item.Children) > 0 {
		return firstLineIsDirective(item.Children[0].Field("directive"))
	}

	return false
}

func blankLineSeparator(n int) doc.Doc {
	if n <= 0 {
		return doc.HardLine()
	}

	parts := make([]doc.Doc, 0, n+1)
	for range n {
		parts = append(parts, doc.HardLine())
	}

	parts = append(parts, doc.HardLine())

	return doc.Concat(parts...)
}

func (s *state) separatorForItem(separator doc.Doc, item *parser.Node) doc.Doc {
	return s.directiveAwareSeparator(separator, item)
}
