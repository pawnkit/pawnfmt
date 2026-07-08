package format

import (
	"bytes"

	parser "github.com/pawnkit/pawn-parser"
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

	if extra := nextSibling(n, target); extra != nil {
		return doc.Concat(doc.Text(prefix), s.formatTagQualifiedTarget(target, extra), semiDoc(n))
	}

	return doc.Concat(doc.Text(prefix), doc.Text(target.Text(s.source)), semiDoc(n))
}

func nextSibling(n, target *parser.Node) *parser.Node {
	for i, c := range n.Children {
		if c == target {
			if i+1 < len(n.Children) {
				return n.Children[i+1]
			}

			return nil
		}
	}

	return nil
}

func (s *state) formatTagQualifiedTarget(tag, name *parser.Node) doc.Doc {
	colon := bytes.IndexByte(s.source[tag.End:name.Start], ':')
	if colon < 0 {
		return doc.Text(string(s.source[tag.Start:name.End]))
	}

	return doc.Concat(s.formatSimpleTag(tag, tag.End+colon+1, false), doc.Text(name.Text(s.source)))
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
