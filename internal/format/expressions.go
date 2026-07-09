package format

import (
	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) binaryOperatorText(n *parser.Node) string {
	return n.Tok.Text(s.source)
}

func (s *state) operatorTokenRaw(n *parser.Node) (doc.Doc, bool) {
	if n.OperatorTokenHasComment() {
		return s.raw(n), true
	}

	return nil, false
}

func (s *state) formatBinaryExpression(n *parser.Node) doc.Doc {
	if raw, ok := s.operatorTokenRaw(n); ok {
		return raw
	}

	left := n.Field("left")
	right := n.Field("right")
	op := s.binaryOperatorText(n)

	spaced := s.config.SpaceAroundOperators
	if s.config.BreakBinaryOperator == config.BinaryOperatorBreakBefore {
		opText, lineBreak := op, doc.SoftLine()
		if spaced {
			opText, lineBreak = op+" ", doc.Line()
		}

		return doc.Group(doc.Concat(
			s.formatNode(left),
			doc.Indent(doc.Concat(lineBreak, doc.Text(opText), s.formatNode(right))),
		))
	}

	if spaced {
		return doc.Group(doc.Concat(s.formatNode(left), doc.Text(" "+op), doc.Line(), s.formatNode(right)))
	}

	return doc.Group(doc.Concat(s.formatNode(left), doc.Text(op), doc.SoftLine(), s.formatNode(right)))
}

func (s *state) formatAssignmentExpression(n *parser.Node) doc.Doc {
	if raw, ok := s.operatorTokenRaw(n); ok {
		return raw
	}

	if n.Field("right") != nil && n.Field("right").Kind == parser.KindAssignmentExpression {
		return s.formatAssignmentChain(n)
	}

	left := n.Field("left")
	right := n.Field("right")

	op := s.binaryOperatorText(n)
	if s.config.SpaceAroundOperators {
		if right != nil && right.Kind != parser.KindAssignmentExpression {
			return doc.Concat(s.formatNode(left), doc.Text(" "+op+" "), s.formatNode(right))
		}

		return doc.Group(doc.Concat(
			s.formatNode(left), doc.Text(" "+op),
			doc.Indent(doc.Concat(doc.Line(), s.formatNode(right))),
		))
	}

	if right != nil && right.Kind != parser.KindAssignmentExpression {
		return doc.Concat(s.formatNode(left), doc.Text(op), s.formatNode(right))
	}

	return doc.Group(doc.Concat(s.formatNode(left), doc.Text(op), doc.Indent(doc.Concat(doc.SoftLine(), s.formatNode(right)))))
}

func (s *state) formatAssignmentChain(n *parser.Node) doc.Doc {
	var items []doc.Doc

	current := n
	for current != nil && current.Kind == parser.KindAssignmentExpression {
		if raw, ok := s.operatorTokenRaw(current); ok {
			return raw
		}

		left := current.Field("left")

		op := s.binaryOperatorText(current)
		if s.config.SpaceAroundOperators {
			items = append(items, doc.Concat(s.formatNode(left), doc.Text(" "+op)))
		} else {
			items = append(items, doc.Concat(s.formatNode(left), doc.Text(op)))
		}

		current = current.Field("right")
	}

	if len(items) == 0 {
		return s.formatNode(current)
	}

	separator := doc.Text("")
	if s.config.SpaceAroundOperators {
		separator = doc.Text(" ")
	}

	items[len(items)-1] = doc.Concat(items[len(items)-1], separator, s.formatNode(current))
	if len(items) == 1 {
		return doc.Concat(items...)
	}

	return doc.Group(doc.Concat(
		items[0],
		doc.Indent(doc.Concat(doc.Line(), doc.Join(doc.Line(), items[1:]...))),
	))
}

func (s *state) formatUnaryExpression(n *parser.Node) doc.Doc {
	if raw, ok := s.operatorTokenRaw(n); ok {
		return raw
	}

	op := s.binaryOperatorText(n)

	return doc.Concat(doc.Text(op), s.formatNode(n.Field("expression")))
}

func (s *state) formatUpdateExpression(n *parser.Node) doc.Doc {
	if raw, ok := s.operatorTokenRaw(n); ok {
		return raw
	}

	op := s.binaryOperatorText(n)

	return doc.Concat(s.formatNode(n.Field("expression")), doc.Text(op))
}

func (s *state) formatTernaryExpression(n *parser.Node) doc.Doc {
	cond := n.Field("condition")
	cons := n.Field("consequence")
	alt := n.Field("alternative")

	return doc.Group(doc.Concat(
		s.formatNode(cond),
		doc.Indent(doc.Concat(doc.Line(), doc.Text("? "), s.formatNode(cons))),
		doc.Indent(doc.Concat(doc.Line(), doc.Text(": "), s.formatNode(alt))),
	))
}

func (s *state) formatSizeofLikeExpression(n *parser.Node, keyword string) doc.Doc {
	expr := n.Field("expression")
	if expr != nil && n.End > expr.End {
		return doc.Concat(doc.Text(keyword+"("), s.formatNode(expr), doc.Text(")"))
	}

	return doc.Concat(doc.Text(keyword+" "), s.formatNode(expr))
}

func (s *state) formatDefinedExpression(n *parser.Node) doc.Doc {
	name := n.Field("name")
	if name == nil {
		return doc.Text("defined()")
	}

	if n.End > name.End {
		return doc.Concat(doc.Text("defined("), doc.Text(name.Text(s.source)), doc.Text(")"))
	}

	return doc.Concat(doc.Text("defined "), doc.Text(name.Text(s.source)))
}

func (s *state) formatTaggedExpression(n *parser.Node) doc.Doc {
	tag := n.Field("tag")
	return doc.Concat(doc.Text(tag.Text(s.source)), doc.Text(":"), s.formatNode(n.Field("expression")))
}

func (s *state) formatParenthesizedExpression(n *parser.Node) doc.Doc {
	inner := n.Field("expression")

	open, closeStr := "(", ")"
	if s.config.SpaceInsideParens {
		open, closeStr = "( ", " )"
	}

	return doc.Concat(doc.Text(open), doc.Indent(s.formatNode(inner)), doc.Text(closeStr))
}

func (s *state) formatCallExpression(n *parser.Node) doc.Doc {
	fn := n.Field("function")
	args := n.Field("arguments")

	return doc.Concat(s.formatNode(fn), s.formatArgumentList(args))
}

func (s *state) formatArgumentList(n *parser.Node) doc.Doc {
	if n == nil || len(n.Children) == 0 {
		return doc.Text("()")
	}

	if hasConditionalItem(n.Children) {
		return s.formatDirectiveList(n.Children, "(", ")", false)
	}

	return s.formatParenList(n.Children, s.config.MultilineCallArgs)
}

func (s *state) formatSubscriptExpression(n *parser.Node) doc.Doc {
	target := n.Field("array")
	index := n.Field("index")
	open, closeStr := "[", "]"

	switch {
	case target != nil && subscriptOpensWithBrace(s.source, target.End):
		open, closeStr = "{", "}"
	case s.config.SpaceInsideBrackets:
		open, closeStr = "[ ", " ]"
	}

	idxDoc := doc.Text("")
	if index != nil {
		idxDoc = s.formatNode(index)
	}

	return doc.Concat(s.formatNode(target), doc.Text(open), idxDoc, doc.Text(closeStr))
}

func subscriptOpensWithBrace(source []byte, from int) bool {
	for i := from; i < len(source); i++ {
		switch source[i] {
		case ' ', '\t', '\r', '\n':
			continue
		case '{':
			return true
		default:
			return false
		}
	}

	return false
}

func (s *state) formatStringConcat(n *parser.Node) doc.Doc {
	parts := make([]doc.Doc, len(n.Children))
	for i, c := range n.Children {
		parts[i] = s.formatNode(c)
	}

	return doc.Join(doc.Text(" "), parts...)
}

func (s *state) formatExpressionList(n *parser.Node) doc.Doc {
	var parts []doc.Doc

	for i, c := range n.Children {
		if i > 0 {
			if s.config.SpaceAfterComma {
				parts = append(parts, doc.Line())
			} else {
				parts = append(parts, doc.SoftLine())
			}
		}

		parts = append(parts, s.formatListItem(c, i < len(n.Children)-1))
	}

	return doc.Group(doc.Concat(parts...))
}

func (s *state) formatArrayLiteral(n *parser.Node) doc.Doc {
	if len(n.Children) == 0 {
		if s.config.SpaceInsideBraces {
			return doc.Text("{ }")
		}

		return doc.Text("{}")
	}

	if hasConditionalItem(n.Children) {
		return s.formatDirectiveList(n.Children, "{", "}", false)
	}

	open, closeStr := "{", "}"
	if s.config.SpaceInsideBraces {
		open, closeStr = "{ ", " }"
	}

	items := make([]doc.Doc, len(n.Children))
	for i, c := range n.Children {
		items[i] = s.formatListItem(c, i < len(n.Children)-1)
	}

	separator := doc.Line()
	if !s.config.SpaceAfterComma {
		separator = doc.SoftLine()
	}

	joined := doc.Join(separator, items...)

	return doc.Group(doc.Concat(
		doc.Text(open),
		doc.Indent(doc.Concat(doc.SoftLine(), joined)),
		doc.SoftLine(),
		doc.Text(closeStr),
	))
}
