package format

import (
	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) joinBraceStyle(blockDoc doc.Doc) doc.Doc {
	switch s.config.BraceStyle {
	case config.BraceStyleAllman:
		return doc.Concat(doc.HardLine(), blockDoc)
	case config.BraceStyleWhitesmiths:
		return doc.Indent(doc.Concat(doc.HardLine(), blockDoc))
	default:
		return doc.Concat(doc.Text(" "), blockDoc)
	}
}

func (s *state) blockTrailingKeyword(word string) doc.Doc {
	if s.config.BraceStyle == config.BraceStyle1TBS {
		return doc.Text(" " + word)
	}

	return doc.Concat(doc.HardLine(), doc.Text(word))
}

func (s *state) formatBranchBody(body *parser.Node) doc.Doc {
	if body == nil {
		return doc.Text("")
	}

	if body.Kind == parser.KindEmptyStatement && body.Start == body.End {
		return doc.Text("")
	}

	body = s.applySingleStatementBraces(body)
	if body.Kind == parser.KindBlock {
		return s.joinBraceStyle(s.formatNode(body))
	}

	if s.config.KeepSimpleStatementsSingleLine {
		return doc.Group(doc.Indent(doc.Concat(doc.Line(), s.formatNode(body))))
	}

	return doc.Indent(doc.Concat(doc.HardLine(), s.formatNode(body)))
}

func (s *state) applySingleStatementBraces(body *parser.Node) *parser.Node {
	switch s.config.SingleStatementBraces {
	case config.SingleStatementBracesAlways:
		if body.Kind != parser.KindBlock && body.Kind != parser.KindEmptyStatement {
			return &parser.Node{Kind: parser.KindBlock, Children: []*parser.Node{body}}
		}
	case config.SingleStatementBracesNever:
		if body.Kind == parser.KindBlock && len(body.Children) == 1 {
			return body.Children[0]
		}
	}

	return body
}

func (s *state) formatBlock(n *parser.Node) doc.Doc {
	if len(n.Children) == 0 {
		return doc.Text("{ }")
	}

	wasTopLevel := s.topLevelContext
	s.topLevelContext = false

	defer func() { s.topLevelContext = wasTopLevel }()

	children := n.Children
	widths := s.alignmentWidths(children)
	macroWidths := s.macroAlignmentWidths(children)
	s.applyCommentAlignment(children)

	var inner []doc.Doc

	for i := 0; i < len(children); i++ {
		item := children[i]
		if i > 0 {
			blanks := blankLineSeparator(s.blankLinesBefore(item.LeadingTrivia()))
			inner = append(inner, s.separatorForItem(blanks, item))
		}

		if item.Kind == parser.KindLabelStatement && i+1 < len(children) &&
			!leadingStartsNewLine(item.TrailingTrivia()) && !leadingStartsNewLine(children[i+1].LeadingTrivia()) {
			next := children[i+1]
			inner = append(inner, s.formatNode(item), doc.Text(" "), s.formatNode(next))
			i++

			continue
		}

		s.hint.alignDeclarationWidth = widths[item]
		s.hint.alignMacroValueWidth = macroWidths[item]
		inner = append(inner, s.formatNode(item))
	}

	return doc.Concat(
		doc.Text("{"),
		doc.Indent(doc.Concat(s.itemSeparatorBefore(children[0]), doc.Concat(inner...))),
		doc.HardLine(),
		doc.Text("}"),
	)
}

func (s *state) formatIfStatement(n *parser.Node) doc.Doc {
	suppressAlternative := s.takeSuppressIfAlternative()
	cond := n.Field("condition")
	consequence := n.Field("consequence")

	alternative := n.Field("alternative")
	if suppressAlternative {
		alternative = nil
	}

	var parts []doc.Doc

	parts = append(parts, doc.Text("if "), s.formatNode(cond))

	parts = append(parts, s.formatIfConsequence(consequence, alternative != nil))
	if condAlt := n.Field("conditional_alternatives"); condAlt != nil {
		parts = append(parts, doc.HardLine(), s.formatNode(condAlt))
	}

	if alt := alternative; alt != nil {
		effectiveConsequence := consequence
		if consequence != nil && !s.mustKeepBracesForDanglingElse(consequence, true) {
			effectiveConsequence = s.applySingleStatementBraces(consequence)
		}

		switch {
		case n.Field("conditional_alternatives") != nil:
			parts = append(parts, doc.HardLine(), doc.Text("else"))
		case consequence != nil && effectiveConsequence.Kind == parser.KindBlock:
			parts = append(parts, s.blockTrailingKeyword("else"))
		case alt.Kind == parser.KindIfStatement:
			parts = append(parts, doc.HardLine(), doc.Text("else"))
		case s.inMacroValue:
			parts = append(parts, doc.Text(" else"))
		default:
			parts = append(parts, doc.HardLine(), doc.Text("else"))
		}

		if alt.Kind == parser.KindIfStatement {
			parts = append(parts, s.chainContinuation(alt))
		} else {
			parts = append(parts, s.formatBranchBody(alt))
		}
	}

	return doc.Concat(parts...)
}

func (s *state) formatIfConsequence(body *parser.Node, hasAlternative bool) doc.Doc {
	if s.mustKeepBracesForDanglingElse(body, hasAlternative) {
		return s.joinBraceStyle(s.formatNode(body))
	}

	return s.formatBranchBody(body)
}

func (s *state) mustKeepBracesForDanglingElse(body *parser.Node, hasAlternative bool) bool {
	if !hasAlternative || s.config.SingleStatementBraces != config.SingleStatementBracesNever ||
		body == nil || body.Kind != parser.KindBlock || len(body.Children) != 1 {
		return false
	}

	inner := body.Children[0]

	return inner.Kind == parser.KindIfStatement && inner.Field("alternative") == nil
}

func (s *state) chainContinuation(alt *parser.Node) doc.Doc {
	return doc.Concat(doc.Text(" "), s.formatNode(alt))
}

func (s *state) formatWhileStatement(n *parser.Node) doc.Doc {
	cond := n.Field("condition")
	body := n.Field("body")

	return doc.Concat(doc.Text("while "), s.formatNode(cond), s.formatBranchBody(body))
}

func (s *state) formatDoWhileStatement(n *parser.Node) doc.Doc {
	body := n.Field("body")
	cond := n.Field("condition")

	effectiveBody := body
	if body != nil {
		effectiveBody = s.applySingleStatementBraces(body)
	}

	var parts []doc.Doc

	parts = append(parts, doc.Text("do"))

	parts = append(parts, s.formatBranchBody(body))
	if effectiveBody != nil && effectiveBody.Kind == parser.KindBlock {
		parts = append(parts, s.blockTrailingKeyword("while"))
		parts = append(parts, doc.Text(" "))
	} else {
		parts = append(parts, doc.Text(" while "))
	}

	parts = append(parts, s.formatNode(cond), semiDoc(n))

	return doc.Concat(parts...)
}

func (s *state) formatForStatement(n *parser.Node) doc.Doc {
	init := n.Field("init")
	cond := n.Field("condition")
	incr := n.Field("increment")
	body := n.Field("body")

	initDoc := doc.Text("")

	if init != nil {
		if init.Kind == parser.KindVariableDeclaration {
			initDoc = s.formatVariableDeclarationCore(init)
		} else {
			initDoc = s.formatNode(init)
		}
	}

	condDoc := doc.Text("")
	if cond != nil {
		condDoc = s.formatNode(cond)
	}

	incrDoc := doc.Text("")
	if incr != nil {
		incrDoc = s.formatNode(incr)
	}

	header := doc.Concat(
		doc.Text("for ("),
		initDoc, doc.Text("; "),
		condDoc, doc.Text("; "),
		incrDoc, doc.Text(")"),
	)

	return doc.Concat(header, s.formatBranchBody(body))
}
