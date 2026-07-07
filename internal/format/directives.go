package format

import (
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) directiveSpacer() string {
	if s.config.DirectiveSpacing {
		return " "
	}
	return ""
}

func (s *state) formatIncludeDirective(n *parser.Node) doc.Doc {
	keyword := "include"
	if n.Kind == parser.KindDirectiveTryInclude {
		keyword = "tryinclude"
	}
	path := n.Field("path")
	pathText := ""
	if path != nil {
		pathText = strings.TrimSpace(string(path.Raw))
	}
	core := doc.Text("#" + keyword + s.directiveSpacer() + pathText)
	if path == nil {
		return core
	}
	if trail := s.trailingDoc(path.TrailingTrivia()); trail != nil {
		return doc.Concat(core, trail)
	}
	return core
}

func (s *state) formatConditionDirective(n *parser.Node) doc.Doc {
	cond := n.Field("condition")
	if cond == nil {
		return s.formatRawDirectiveLine(n)
	}
	keyword := directiveKeywordFor(n.Kind)
	rendered := s.renderDoc(s.formatNode(cond))
	if !strings.Contains(rendered, "\n") {
		return doc.Text("#" + keyword + " " + rendered)
	}
	lines := strings.Split(backslashContinue(rendered), "\n")
	parts := make([]doc.Doc, 0, len(lines)*2)
	parts = append(parts, doc.Text("#"+keyword+" "+lines[0]))
	for _, line := range lines[1:] {
		parts = append(parts, doc.HardLine(), doc.Text(line))
	}
	return doc.Concat(parts[0], doc.Indent(doc.Concat(parts[1:]...)))
}

func directiveKeywordFor(k parser.Kind) string {
	switch k {
	case parser.KindDirectiveIf:
		return "if"
	case parser.KindDirectiveElseif:
		return "elseif"
	case parser.KindDirectiveAssert:
		return "assert"
	default:
		return ""
	}
}

func (s *state) formatRawDirectiveLine(n *parser.Node) doc.Doc {
	text := strings.TrimRight(n.Text(s.source), " \t\r\n")
	return doc.RawTextBlock(normalizeDirectiveKeywordSpacing(text))
}

func normalizeDirectiveKeywordSpacing(text string) string {
	i := 1 // skip '#'
	for i < len(text) && isIdentByte(text[i]) {
		i++
	}
	j := i
	for j < len(text) && (text[j] == ' ' || text[j] == '\t') {
		j++
	}
	if j == i || j >= len(text) {
		return text
	}
	return text[:i] + " " + text[j:]
}

func ensureDirectiveKeywordSpacing(text string) string {
	i := 1 // skip '#'
	for i < len(text) && isIdentByte(text[i]) {
		i++
	}
	j := i
	for j < len(text) && (text[j] == ' ' || text[j] == '\t') {
		j++
	}
	if j >= len(text) {
		return text
	}
	return text[:i] + " " + text[j:]
}

func isIdentByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func (s *state) formatConditionalRegion(n *parser.Node) doc.Doc {
	forceTopLevel := s.topLevelContext
	var parts []doc.Doc
	for bi, branch := range n.Children {
		directive := branch.Field("directive")
		if bi > 0 {
			parts = append(parts, s.itemSeparatorBefore(directive))
		}
		parts = append(parts, s.formatNode(directive))

		var items []*parser.Node
		for _, item := range branch.Children {
			if item != directive {
				items = append(items, item)
			}
		}

		var prev *parser.Node
		for i := 0; i < len(items); i++ {
			item := items[i]
			var separator doc.Doc
			switch {
			case prev == nil:
				separator = s.itemSeparatorBefore(item)
			case forceTopLevel:
				separator = s.separatorForItem(s.topLevelSeparator(prev, item), item)
			default:
				blanks := blankLineSeparator(s.blankLinesBefore(item.LeadingTrivia()))
				separator = s.separatorForItem(blanks, item)
			}
			if item.Kind == parser.KindLabelStatement && i+1 < len(items) &&
				!leadingStartsNewLine(item.TrailingTrivia()) && !leadingStartsNewLine(items[i+1].LeadingTrivia()) {
				next := items[i+1]
				parts = append(parts, separator, s.formatNode(item), doc.Text(" "), s.formatNode(next))
				prev = next
				i++
				continue
			}
			if i == len(items)-1 && item.Kind == parser.KindIfStatement && branch.Field("shared_alternative") != nil {
				s.hint.suppressIfAlternative = true
			}
			parts = append(parts, separator, s.formatNode(item))
			prev = item
		}
	}
	if alt := n.Field("alternative"); alt != nil {
		parts = append(parts, doc.HardLine(), doc.Text("else"))
		if alt.Kind == parser.KindIfStatement {
			parts = append(parts, s.chainContinuation(alt))
		} else {
			parts = append(parts, s.formatBranchBody(alt))
		}
	}
	return doc.Concat(parts...)
}
