package format

import (
	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func semiDoc(n *parser.Node) doc.Doc {
	if n.MissingSemi {
		return nil
	}
	return doc.Text(";")
}

func (s *state) formatReturnStatement(n *parser.Node) doc.Doc {
	if val := n.Field("value"); val != nil {
		return doc.Concat(doc.Text("return "), s.formatNode(val), semiDoc(n))
	}
	return doc.Concat(doc.Text("return"), semiDoc(n))
}

func (s *state) formatSimpleTrailingSemi(n *parser.Node, prefix, field string) doc.Doc {
	target := n.Field(field)
	if target == nil {
		return doc.Concat(doc.Text(prefix), semiDoc(n))
	}
	return doc.Concat(doc.Text(prefix), doc.Text(target.Text(s.source)), semiDoc(n))
}

func (s *state) formatLabelStatement(n *parser.Node) doc.Doc {
	label := n.Field("label")
	return doc.Concat(doc.Text(label.Text(s.source)), doc.Text(":"))
}

func (s *state) formatExpressionStatement(n *parser.Node) doc.Doc {
	expr := n.Field("expression")
	if n.HasError && expr == nil {
		return s.raw(n)
	}
	return doc.Group(doc.Concat(s.formatNode(expr), semiDoc(n)))
}

func (s *state) formatMacroInvocationBlock(n *parser.Node) doc.Doc {
	fn := n.Field("function")
	args := n.Field("arguments")
	body := n.Field("body")
	return doc.Concat(
		doc.Text(fn.Text(s.source)),
		s.formatArgumentList(args),
		s.formatBranchBody(body),
	)
}
