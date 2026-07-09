package format

import (
	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func hasConditionalItem(items []*parser.Node) bool {
	for _, it := range items {
		if it.Kind == parser.KindConditionalRegion {
			return true
		}
	}

	return false
}

func (s *state) formatDirectiveList(items []*parser.Node, open, closeStr string, trailingComma bool) doc.Doc {
	return doc.Concat(
		doc.Text(open),
		doc.Indent(doc.Concat(s.itemSeparatorBefore(items[0]), s.formatListItemsWithDirectives(items, trailingComma))),
		doc.HardLine(),
		doc.Text(closeStr),
	)
}

func (s *state) formatListItemsWithDirectives(items []*parser.Node, trailingLast bool) doc.Doc {
	var parts []doc.Doc

	for i, it := range items {
		if i > 0 {
			parts = append(parts, s.itemSeparatorBefore(it))
		}

		if it.Kind == parser.KindConditionalRegion {
			parts = append(parts, s.formatConditionalRegionInList(it))
			continue
		}

		parts = append(parts, s.formatListItem(it, i < len(items)-1 || trailingLast))
	}

	return doc.Concat(parts...)
}

func (s *state) formatConditionalRegionInList(n *parser.Node) doc.Doc {
	var parts []doc.Doc

	for i, branch := range n.Children {
		directive := branch.Field("directive")
		if i > 0 {
			parts = append(parts, s.itemSeparatorBefore(directive))
		}

		parts = append(parts, s.formatNode(directive))
		for _, item := range branch.Children {
			if item == directive {
				continue
			}

			parts = append(parts, s.itemSeparatorBefore(item))
			if item.Kind == parser.KindConditionalRegion {
				parts = append(parts, s.formatConditionalRegionInList(item))
				continue
			}

			parts = append(parts, s.formatListItem(item, true))
		}
	}

	return doc.Concat(parts...)
}
