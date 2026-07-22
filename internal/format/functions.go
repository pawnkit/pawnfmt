package format

import (
	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatFunction(n *parser.Node) doc.Doc { //nolint:gocyclo // Function syntax has several optional parts.
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

	name := n.Field("name")

	dimensions := dimsOf(n)
	for _, dimension := range dimensions {
		if dimension.Start < name.Start {
			parts = append(parts, s.formatDimensions([]*parser.Node{dimension}))
		}
	}

	parts = append(parts, doc.Text(name.Text(s.source)))
	for _, dimension := range dimensions {
		if dimension.Start >= name.End {
			parts = append(parts, s.formatDimensions([]*parser.Node{dimension}))
		}
	}

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

	children := s.mergeBareParameterQualifiers(n.Children)
	if hasConditionalItem(children) {
		return s.formatDirectiveList(children, "(", ")", false)
	}

	return s.formatParenList(children, s.config.MultilineFunctionParams)
}

func (s *state) mergeBareParameterQualifiers(children []*parser.Node) []*parser.Node {
	merged := make([]*parser.Node, 0, len(children))
	for i, c := range children {
		if isBareParameterIdentifier(c) && i+1 < len(children) && children[i+1].Kind == parser.KindParameter &&
			!hasCommaBetween(s.source, c.End, children[i+1].Start) {
			if s.paramQualifiers == nil {
				s.paramQualifiers = make(map[*parser.Node]*parser.Node)
			}

			s.paramQualifiers[children[i+1]] = c

			continue
		}

		merged = append(merged, c)
	}

	return merged
}

func isBareParameterIdentifier(n *parser.Node) bool {
	return n.Kind == parser.KindParameter && len(n.Children) == 1 &&
		n.Children[0].Kind == parser.KindIdentifier &&
		n.Start == n.Children[0].Start && n.End == n.Children[0].End
}

func hasCommaBetween(source []byte, from, to int) bool {
	for i := from; i < to && i < len(source); i++ {
		if source[i] == ',' {
			return true
		}
	}

	return false
}

func (s *state) formatParenList(nodes []*parser.Node, style config.MultilineListStyle) doc.Doc {
	open, closeStr := "(", ")"
	if s.config.SpaceInsideParens {
		open, closeStr = "( ", " )"
	}

	if style == config.MultilineListOnePerLine {
		return s.formatExplodedList(nodes, open, closeStr)
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
			doc.Text(closeStr),
		))
	}

	items := make([]doc.Doc, len(nodes))
	for i, n := range nodes {
		items[i] = s.formatListItem(n, i < len(nodes)-1)
	}

	joined := doc.Join(sepLine, items...)

	return doc.Group(doc.Concat(
		doc.Text(open),
		doc.Indent(doc.Concat(doc.SoftLine(), joined)),
		doc.SoftLine(),
		doc.Text(closeStr),
	))
}

func (s *state) formatExplodedList(nodes []*parser.Node, open, closeTok string) doc.Doc {
	items := make([]doc.Doc, len(nodes))
	for i, n := range nodes {
		items[i] = s.formatListItem(n, i < len(nodes)-1)
	}

	return doc.Concat(
		doc.Text(open),
		doc.Indent(doc.Concat(doc.HardLine(), doc.Join(doc.HardLine(), items...))),
		doc.HardLine(),
		doc.Text(closeTok),
	)
}

func (s *state) formatParameter(n *parser.Node) doc.Doc {
	var qualifier doc.Doc
	if q := s.paramQualifiers[n]; q != nil {
		qualifier = doc.Concat(doc.Text(q.Text(s.source)), doc.Text(" "))
	}

	if len(n.Children) == 0 {
		return doc.Concat(qualifier, doc.Text(n.Text(s.source)))
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

		if s.config.SpaceAfterUnaryOperator {
			parts = append(parts, doc.Text(" "))
		}
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
	if generic := n.Field("generic"); generic != nil {
		parts = append(parts, s.formatStateSelector(generic))
	}

	if def := n.Field("default_value"); def != nil {
		parts = append(parts, s.assignmentSeparator(), s.formatNode(def))
	}

	return doc.Concat(qualifier, doc.Concat(parts...))
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
