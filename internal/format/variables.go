package format

import (
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatVariableDeclaration(n *parser.Node) doc.Doc {
	return doc.Group(doc.Concat(s.formatVariableDeclarationCore(n), doc.Text(";")))
}

func (s *state) formatVariableDeclarationCore(n *parser.Node) doc.Doc {
	width := s.takeAlignDeclarationWidth()
	if width > 0 {
		if decl := singleInitializedDeclarator(n); decl != nil {
			return s.formatAlignedSingleDeclaration(n, decl, width)
		}
	}
	var parts []doc.Doc
	var declarators []*parser.Node
	for _, c := range n.Children {
		if c.Kind == parser.KindIdentifier {
			parts = append(parts, doc.Text(c.Text(s.source)), doc.Text(" "))
			continue
		}
		declarators = append(declarators, c)
	}
	if hasConditionalItem(declarators) {
		if len(declarators) > 0 && declarators[0].Kind != parser.KindConditionalRegion {
			first := s.formatListItem(declarators[0], len(declarators) > 1)
			if len(declarators) == 1 {
				return doc.Concat(doc.Concat(parts...), first)
			}
			return doc.Concat(
				doc.Concat(parts...), first,
				doc.Indent(doc.Concat(s.itemSeparatorBefore(declarators[1]), s.formatListItemsWithDirectives(declarators[1:], false))),
			)
		}
		return doc.Concat(
			doc.Concat(parts...),
			doc.Indent(doc.Concat(s.itemSeparatorBefore(declarators[0]), s.formatListItemsWithDirectives(declarators, false))),
		)
	}
	var joined []doc.Doc
	for i, d := range declarators {
		item := s.formatListItem(d, i < len(declarators)-1)
		if i == 0 {
			joined = append(joined, item)
			continue
		}
		separator := doc.SoftLine()
		if s.config.SpaceAfterComma {
			separator = doc.Line()
		}
		joined = append(joined, doc.Indent(doc.Concat(separator, item)))
	}
	parts = append(parts, doc.Concat(joined...))
	return doc.Group(doc.Concat(parts...))
}

func (s *state) formatVariableDeclarator(n *parser.Node) doc.Doc {
	var parts []doc.Doc
	if tag := n.Field("tag"); tag != nil {
		parts = append(parts, s.formatTagPrefix(tag, false))
	}
	name := n.Field("name")
	if name != nil {
		parts = append(parts, doc.Text(name.Text(s.source)))
	}
	if capacity := n.Field("capacity"); capacity != nil {
		parts = append(parts, s.formatStateSelector(capacity))
	}
	parts = append(parts, s.formatDimensions(dimsOf(n)))
	if init := n.Field("initializer"); init != nil {
		parts = append(parts, s.assignmentSeparator(), s.formatNode(init))
	}
	return doc.Concat(parts...)
}

func singleInitializedDeclarator(n *parser.Node) *parser.Node {
	if n.Kind != parser.KindVariableDeclaration {
		return nil
	}
	var decl *parser.Node
	for _, c := range n.Children {
		if c.Kind == parser.KindIdentifier {
			continue
		}
		if decl != nil {
			return nil
		}
		decl = c
	}
	if decl == nil || decl.Field("initializer") == nil {
		return nil
	}
	return decl
}

func (s *state) formatVariableDeclarationPrefix(n, decl *parser.Node) doc.Doc {
	var parts []doc.Doc
	for _, c := range n.Children {
		if c.Kind == parser.KindIdentifier {
			parts = append(parts, doc.Text(c.Text(s.source)), doc.Text(" "))
		}
	}
	if tag := decl.Field("tag"); tag != nil {
		parts = append(parts, s.formatTagPrefix(tag, false))
	}
	if name := decl.Field("name"); name != nil {
		parts = append(parts, doc.Text(name.Text(s.source)))
	}
	if capacity := decl.Field("capacity"); capacity != nil {
		parts = append(parts, s.formatStateSelector(capacity))
	}
	parts = append(parts, s.formatDimensions(dimsOf(decl)))
	return doc.Concat(parts...)
}

func (s *state) formatAlignedSingleDeclaration(n, decl *parser.Node, width int) doc.Doc {
	prefix := s.formatVariableDeclarationPrefix(n, decl)
	pad := width - len(s.renderFlat(prefix))
	padding := doc.Text("")
	if pad > 0 {
		padding = doc.Text(strings.Repeat(" ", pad))
	}
	init := decl.Field("initializer")
	return doc.Concat(prefix, padding, s.assignmentSeparator(), s.formatNode(init))
}

func (s *state) alignmentWidths(items []*parser.Node) map[*parser.Node]int {
	if !s.config.AlignConsecutiveDeclarations {
		return nil
	}
	widths := make(map[*parser.Node]int)
	var run []*parser.Node
	var runDecls []*parser.Node
	flush := func() {
		if len(run) > 1 {
			maxWidth := 0
			measured := make([]int, len(run))
			for i, item := range run {
				w := len(s.renderFlat(s.formatVariableDeclarationPrefix(item, runDecls[i])))
				measured[i] = w
				if w > maxWidth {
					maxWidth = w
				}
			}
			for i, item := range run {
				if measured[i] < maxWidth {
					widths[item] = maxWidth
				}
			}
		}
		run = nil
		runDecls = nil
	}
	for i, item := range items {
		if i > 0 && (s.blankLinesBefore(item.LeadingTrivia()) > 0 || hasCommentTrivia(item.LeadingTrivia())) {
			flush()
		}
		decl := singleInitializedDeclarator(item)
		if decl == nil {
			flush()
			continue
		}
		run = append(run, item)
		runDecls = append(runDecls, decl)
	}
	flush()
	return widths
}
