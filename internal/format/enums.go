package format

import (
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatEnumDeclaration(n *parser.Node) doc.Doc {
	var parts []doc.Doc

	for _, c := range n.Children {
		if c.Kind == parser.KindIdentifier && c != n.Field("name") {
			parts = append(parts, doc.Text(c.Text(s.source)), doc.Text(" "))
		}
	}

	parts = append(parts, doc.Text("enum"))
	if name := n.Field("name"); name != nil {
		parts = append(parts, doc.Text(" "), doc.Text(name.Text(s.source)))
	}

	if tag := n.Field("tag"); tag != nil {
		parts = append(parts, doc.Text(tag.Text(s.source)))
	}

	if increment := n.Field("increment"); increment != nil {
		parts = append(parts, doc.Text(" "), doc.Text(increment.Text(s.source)))
	}

	body := n.Field("body")
	parts = append(parts, s.joinBraceStyle(s.formatEnumBody(body)))

	hadSemi := body != nil && n.End > body.End
	if hadSemi || s.config.Semicolons == config.SemicolonsAlways {
		parts = append(parts, doc.Text(";"))
	}

	return doc.Concat(parts...)
}

func (s *state) formatEnumBody(body *parser.Node) doc.Doc {
	if body == nil || len(body.Children) == 0 {
		return doc.Text("{ }")
	}

	trailingComma := s.config.EnumTrailingComma == config.EnumTrailingCommaAlways || sourceHasTrailingComma(body, s.source)
	if hasConditionalItem(body.Children) {
		return s.formatDirectiveList(body.Children, "{", "}", trailingComma)
	}

	var prefixWidths []int

	maxPrefixWidth := 0

	if s.config.AlignEnumFields {
		prefixWidths = make([]int, len(body.Children))
		for i, entry := range body.Children {
			w := len(s.renderFlat(s.formatEnumEntryPrefix(entry)))

			prefixWidths[i] = w
			if entry.Field("value") != nil && w > maxPrefixWidth {
				maxPrefixWidth = w
			}
		}
	}

	var parts []doc.Doc

	parts = append(parts, doc.Text("{"))

	var inner []doc.Doc

	for i, entry := range body.Children {
		if i > 0 {
			inner = append(inner, doc.HardLine())

			if blanks := s.blankLinesBefore(entry.LeadingTrivia()); blanks > 0 {
				for range blanks {
					inner = append(inner, doc.HardLine())
				}
			}
		}

		addComma := i < len(body.Children)-1 || trailingComma
		if s.config.AlignEnumFields && entry.Field("value") != nil && prefixWidths[i] < maxPrefixWidth {
			padding := doc.Text(strings.Repeat(" ", maxPrefixWidth-prefixWidths[i]))

			entryDoc := doc.Concat(s.formatEnumEntryPrefix(entry), padding, s.assignmentSeparator(), s.formatNode(entry.Field("value")))
			if addComma {
				entryDoc = doc.Concat(entryDoc, doc.Text(","))
			}

			if trail := s.trailingDoc(entry.TrailingTrivia()); trail != nil {
				entryDoc = doc.Concat(entryDoc, trail)
			}

			inner = append(inner, entryDoc)

			continue
		}

		inner = append(inner, s.formatListItem(entry, addComma))
	}

	parts = append(parts, doc.Indent(doc.Concat(doc.HardLine(), doc.Concat(inner...))))
	parts = append(parts, doc.HardLine(), doc.Text("}"))

	return doc.Concat(parts...)
}

func sourceHasTrailingComma(body *parser.Node, source []byte) bool {
	if len(body.Children) == 0 {
		return false
	}

	last := body.Children[len(body.Children)-1]

	end := last.End
	for end < body.End && end < len(source) {
		c := source[end]
		if c == ',' {
			return true
		}

		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			end++
			continue
		}

		break
	}

	return false
}

func (s *state) formatEnumEntry(n *parser.Node) doc.Doc {
	parts := []doc.Doc{s.formatEnumEntryPrefix(n)}
	if val := n.Field("value"); val != nil {
		parts = append(parts, s.assignmentSeparator(), s.formatNode(val))
	}

	return doc.Concat(parts...)
}

func (s *state) formatEnumEntryPrefix(n *parser.Node) doc.Doc {
	var parts []doc.Doc
	if tag := n.Field("tag"); tag != nil {
		parts = append(parts, s.formatTagPrefix(tag, false))
	}

	if name := n.Field("name"); name != nil {
		parts = append(parts, doc.Text(name.Text(s.source)))
	}

	parts = append(parts, s.formatDimensions(dimsOf(n)))

	return doc.Concat(parts...)
}
