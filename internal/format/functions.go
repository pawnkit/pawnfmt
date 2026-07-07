package format

import (
	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatFunction(n *parser.Node) doc.Doc {
	var parts []doc.Doc
	for _, c := range n.Children {
		if c.Kind != parser.KindIdentifier {
			break
		}
		if c == n.Field("name") {
			break
		}
		parts = append(parts, doc.Text(c.Text(s.source)), doc.Text(" "))
	}
	if tag := n.Field("tag"); tag != nil {
		parts = append(parts, s.formatTagPrefix(tag, false))
	}
	parts = append(parts, s.formatDimensions(dimsOf(n)))
	parts = append(parts, doc.Text(n.Field("name").Text(s.source)))
	if s.config.SpaceBeforeFunctionParen {
		parts = append(parts, doc.Text(" "))
	}
	parts = append(parts, s.formatParameterList(n.Field("parameters")))
	if state := n.Field("state"); state != nil {
		parts = append(parts, doc.Text(" "), s.formatStateSelector(state))
	}

	body := n.Field("body")
	if body == nil {
		if alias := n.Field("alias"); alias != nil {
			parts = append(parts, s.assignmentSeparator(), s.formatNode(alias))
		}
		if !n.MissingSemi {
			parts = append(parts, doc.Text(";"))
		}
		return doc.Concat(parts...)
	}
	if body.Kind == parser.KindConditionalRegion {
		parts = append(parts, doc.Indent(doc.Concat(doc.HardLine(), s.formatNode(body))))
		return doc.Concat(parts...)
	}
	parts = append(parts, s.joinBraceStyle(s.formatNode(body)))
	return doc.Concat(parts...)
}

func (s *state) formatParameterList(n *parser.Node) doc.Doc {
	if n == nil {
		return doc.Text("()")
	}
	if len(n.Children) == 0 {
		return doc.Text("()")
	}
	if hasConditionalItem(n.Children) {
		return s.formatDirectiveList(n.Children, "(", ")", s.config.TrailingComma == config.TrailingCommaMultiline)
	}
	return s.formatParenList(n.Children, s.config.MultilineFunctionParams, s.hasMagicTrailingComma(n))
}

func (s *state) formatParenList(nodes []*parser.Node, style config.MultilineListStyle, forceMultiline bool) doc.Doc {
	open, close := "(", ")"
	if s.config.SpaceInsideParens {
		open, close = "( ", " )"
	}

	if forceMultiline || style == config.MultilineListOnePerLine {
		return s.formatExplodedList(nodes, open, close)
	}

	sepLine := doc.SoftLine()
	if s.config.SpaceAfterComma {
		sepLine = doc.Line()
	}

	if style == config.MultilineListBinPack {
		items := make([]doc.Doc, len(nodes))
		for i, n := range nodes {
			items[i] = s.formatListItem(n, i < len(nodes)-1)
		}
		var fillParts []doc.Doc
		for i, it := range items {
			if i > 0 {
				fillParts = append(fillParts, sepLine)
			}
			fillParts = append(fillParts, it)
		}
		return doc.Group(doc.Concat(
			doc.Text(open),
			doc.Indent(doc.Concat(doc.SoftLine(), doc.Fill(fillParts...))),
			doc.SoftLine(),
			doc.Text(close),
		))
	}

	items := make([]doc.Doc, len(nodes))
	for i, n := range nodes {
		if i == len(nodes)-1 {
			items[i] = s.formatLastListItem(n)
		} else {
			items[i] = s.formatListItem(n, true)
		}
	}
	joined := doc.Join(sepLine, items...)
	return doc.Group(doc.Concat(
		doc.Text(open),
		doc.Indent(doc.Concat(doc.SoftLine(), joined)),
		doc.SoftLine(),
		doc.Text(close),
	))
}

func (s *state) formatExplodedList(nodes []*parser.Node, open, close string) doc.Doc {
	trailingComma := s.config.TrailingComma == config.TrailingCommaMultiline
	items := make([]doc.Doc, len(nodes))
	for i, n := range nodes {
		last := i == len(nodes)-1
		items[i] = s.formatListItem(n, !last || trailingComma)
	}
	return doc.Concat(
		doc.Text(open),
		doc.Indent(doc.Concat(doc.HardLine(), doc.Join(doc.HardLine(), items...))),
		doc.HardLine(),
		doc.Text(close),
	)
}

func (s *state) formatParameter(n *parser.Node) doc.Doc {
	if len(n.Children) == 0 {
		return doc.Text(n.Text(s.source))
	}
	tag, name := n.Field("tag"), n.Field("name")
	var parts []doc.Doc
	for _, c := range n.Children {
		if c == tag || c == name {
			break
		}
		if c.Kind == parser.KindIdentifier {
			parts = append(parts, doc.Text(c.Text(s.source)), doc.Text(" "))
		}
	}
	if byRefBeforeName(n, s.source) {
		parts = append(parts, doc.Text("&"))
	}
	if tag != nil {
		parts = append(parts, s.formatTagPrefix(tag, false))
	}
	if name != nil {
		parts = append(parts, doc.Text(name.Text(s.source)))
	} else {
		parts = append(parts, doc.Text("..."))
	}
	parts = append(parts, s.formatDimensions(dimsOf(n)))
	if def := n.Field("default_value"); def != nil {
		parts = append(parts, s.assignmentSeparator(), s.formatNode(def))
	}
	return doc.Concat(parts...)
}

func byRefBeforeName(n *parser.Node, source []byte) bool {
	anchor := n.Field("tag")
	if anchor == nil {
		anchor = n.Field("name")
	}
	if anchor == nil {
		return false
	}
	pos := anchor.Start - 1
	for pos >= n.Start && (source[pos] == ' ' || source[pos] == '\t') {
		pos--
	}
	return pos >= n.Start && source[pos] == '&'
}
