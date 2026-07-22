package format

import (
	"strings"

	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) assignmentSeparator() doc.Doc {
	if s.config.SpaceAroundOperators {
		return doc.Text(" = ")
	}

	return doc.Text("=")
}

func (s *state) formatTagPrefix(tag *parser.Node, forceTight bool) doc.Doc {
	if tag == nil {
		return nil
	}

	if tag.Start < len(s.source) && s.source[tag.Start] == '{' {
		return s.formatWildcardTag(tag)
	}

	if tag.Field("generic") != nil {
		return doc.Text(tag.Text(s.source))
	}

	if len(tag.Children) == 0 {
		return doc.Text(tag.Text(s.source))
	}

	return s.formatSimpleTag(tag.Children[0], tag.End, forceTight)
}

func (s *state) formatSimpleTag(nameLeaf *parser.Node, tagEnd int, forceTight bool) doc.Doc {
	name := nameLeaf.Text(s.source)
	if forceTight || s.config.TagColonSpacing == config.TagColonSpacingCompact {
		return doc.Text(name + ":")
	}

	if s.config.TagColonSpacing == config.TagColonSpacingTight {
		return doc.Text(name + ": ")
	}

	colonStart := tagEnd - 1

	var b strings.Builder
	b.WriteString(name)

	if colonStart > nameLeaf.End {
		b.WriteByte(' ')
	}

	b.WriteByte(':')

	if tagEnd < len(s.source) && (s.source[tagEnd] == ' ' || s.source[tagEnd] == '\t') {
		b.WriteByte(' ')
	}

	return doc.Text(b.String())
}

func (s *state) formatWildcardTag(tag *parser.Node) doc.Doc {
	names := make([]string, 0, len(tag.Children))
	for _, c := range tag.Children {
		names = append(names, c.Text(s.source))
	}

	separator := ","
	if s.config.SpaceAfterComma {
		separator = ", "
	}

	return doc.Text("{" + strings.Join(names, separator) + "}:")
}

func (s *state) formatDimensions(dims []*parser.Node) doc.Doc {
	var parts []doc.Doc
	if len(dims) > 0 && s.config.SpaceBeforeArrayBrackets {
		parts = append(parts, doc.Text(" "))
	}

	for _, d := range dims {
		inner := doc.Text("")
		if size := d.Field("size"); size != nil {
			inner = s.formatNode(size)
		}

		if packed := d.Field("packed"); packed != nil {
			inner = doc.Concat(inner, doc.Text(" char"))
		}

		if s.config.SpaceInsideBrackets {
			parts = append(parts, doc.Text("[ "), inner, doc.Text(" ]"))
		} else {
			parts = append(parts, doc.Text("["), inner, doc.Text("]"))
		}
	}

	return doc.Concat(parts...)
}

func dimsOf(n *parser.Node) []*parser.Node {
	var dims []*parser.Node

	for _, c := range n.Children {
		if c.Kind == parser.KindDimension {
			dims = append(dims, c)
		}
	}

	return dims
}

func (s *state) formatStateSelector(n *parser.Node) doc.Doc {
	if n.Kind != parser.KindTaggedType {
		return doc.Text(n.Text(s.source))
	}

	names := make([]string, 0, len(n.Children))
	for _, c := range n.Children {
		names = append(names, c.Text(s.source))
	}

	sep := ","
	if s.config.SpaceAfterComma {
		sep = ", "
	}

	return doc.Text("<" + strings.Join(names, sep) + ">")
}
