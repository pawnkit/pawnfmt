package format

import (
	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatSwitchStatement(n *parser.Node) doc.Doc {
	cond := n.Field("condition")

	var clauses []*parser.Node

	for _, c := range n.Children {
		if c == cond {
			continue
		}

		clauses = append(clauses, c)
	}

	var body []doc.Doc

	for i, clause := range clauses {
		if i > 0 {
			body = append(body, s.separatorForItem(doc.HardLine(), clause))
		}

		body = append(body, s.formatNode(clause))
	}

	bodyDoc := doc.Text("{ }")

	if len(clauses) > 0 {
		content := doc.Concat(doc.HardLine(), doc.Concat(body...))
		if s.config.IndentCaseLabels {
			content = doc.Indent(content)
		}

		bodyDoc = doc.Concat(
			doc.Text("{"),
			content,
			doc.HardLine(),
			doc.Text("}"),
		)
	}

	return doc.Concat(doc.Text("switch "), s.formatNode(cond), s.joinBraceStyle(bodyDoc))
}

func (s *state) formatSwitchClauseBody(body *parser.Node) doc.Doc {
	if body == nil || body.Kind == parser.KindEmptyStatement {
		return doc.Text("")
	}

	rendered := doc.Concat(doc.HardLine(), s.formatNode(body))
	if s.config.IndentCaseContents {
		return doc.Indent(rendered)
	}

	return rendered
}

func (s *state) formatCaseClause(n *parser.Node) doc.Doc {
	if n.Kind == parser.KindDefaultClause {
		return doc.Concat(doc.Text("default:"), s.formatSwitchClauseBody(n.Field("body")))
	}

	values := n.Field("values")

	return doc.Concat(doc.Text("case "), s.formatNode(values), doc.Text(":"), s.formatSwitchClauseBody(n.Field("body")))
}

func (s *state) formatCaseValueList(n *parser.Node) doc.Doc {
	items := make([]doc.Doc, 0, len(n.Children))
	for _, c := range n.Children {
		items = append(items, s.formatNode(c))
	}

	separator := doc.Text(",")
	if s.config.SpaceAfterComma {
		separator = doc.Text(", ")
	}

	return doc.Join(separator, items...)
}

func (s *state) formatCaseRange(n *parser.Node) doc.Doc {
	return doc.Concat(s.formatNode(n.Field("start")), doc.Text(" .. "), s.formatNode(n.Field("end")))
}
