package format

import (
	"slices"
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatDefineDirective(n *parser.Node) doc.Doc {
	name := n.Field("name")
	params := n.Field("parameters")
	value := n.Field("value")

	prefix := s.formatDefineDirectivePrefix(name, params)
	if value == nil {
		return prefix
	}

	if width := s.takeAlignMacroValueWidth(); width > 0 {
		pad := width - len(s.renderFlat(prefix))

		padding := doc.Text("")
		if pad > 0 {
			padding = doc.Text(strings.Repeat(" ", pad))
		}

		return doc.Concat(prefix, padding, doc.Text(" "), s.formatMacroValue(value))
	}

	separator := " "

	previousEnd := name.End
	if params != nil {
		previousEnd = params.End
	}

	if value.Start == previousEnd {
		separator = ""
	}

	return doc.Concat(prefix, doc.Text(separator), s.formatMacroValue(value))
}

func (s *state) formatDefineDirectivePrefix(name, params *parser.Node) doc.Doc {
	parts := []doc.Doc{doc.Text("#define "), doc.Text(name.Text(s.source))}
	if params != nil {
		parts = append(parts, s.formatMacroParamList(params))
	}

	return doc.Concat(parts...)
}

func (s *state) macroAlignmentWidths(items []*parser.Node) map[*parser.Node]int {
	if !s.config.AlignConsecutiveMacros {
		return nil
	}

	widths := make(map[*parser.Node]int)

	var run []*parser.Node

	flush := func() {
		if len(run) > 1 {
			maxWidth := 0

			measured := make([]int, len(run))
			for i, item := range run {
				w := len(s.renderFlat(s.formatDefineDirectivePrefix(item.Field("name"), item.Field("parameters"))))

				measured[i] = w
				if w > maxWidth {
					maxWidth = w
				}
			}

			for i, item := range run {
				if measured[i] < maxWidth {
					widths[item] = maxWidth
				}
			}
		}

		run = nil
	}

	for i, item := range items {
		if i > 0 && (s.blankLinesBefore(item.LeadingTrivia()) > 0 || hasCommentTrivia(item.LeadingTrivia())) {
			flush()
		}

		if item.Kind != parser.KindDirectiveDefine || item.Field("value") == nil {
			flush()
			continue
		}

		run = append(run, item)
	}

	flush()

	return widths
}

func (s *state) formatMacroParamList(n *parser.Node) doc.Doc {
	if len(n.Children) == 0 {
		return doc.Text("()")
	}

	sep := ","
	if s.config.SpaceAfterComma {
		sep = ", "
	}

	names := make([]string, 0, len(n.Children))
	for _, c := range n.Children {
		names = append(names, c.Text(s.source))
	}

	return doc.Text("(" + strings.Join(names, sep) + ")")
}

func (s *state) formatMacroValue(value *parser.Node) doc.Doc {
	if value.Kind == parser.KindRaw || value.Kind == parser.KindMacroBody {
		return doc.RawTextBlock(strings.TrimRight(value.Text(s.source), " \t"))
	}

	wasInMacro := s.inMacroValue
	s.inMacroValue = true

	defer func() { s.inMacroValue = wasInMacro }()

	rendered := s.formatNode(value)
	if containsHardLine(rendered) {
		return doc.RawTextBlock(backslashContinue(s.renderDoc(rendered)))
	}

	return doc.RawTextBlock(s.renderFlat(rendered))
}

func containsHardLine(d doc.Doc) bool {
	switch v := d.(type) {
	case doc.HardLineDoc:
		return true
	case doc.BreakParentDoc:
		return true
	case doc.ConcatDoc:
		if slices.ContainsFunc(v.Parts, containsHardLine) {
			return true
		}
	case doc.FillDoc:
		if slices.ContainsFunc(v.Parts, containsHardLine) {
			return true
		}
	case doc.IndentDoc:
		return containsHardLine(v.Contents)
	case doc.ResetIndentDoc:
		return containsHardLine(v.Contents)
	case doc.OutdentDoc:
		return containsHardLine(v.Contents)
	case doc.GroupDoc:
		return containsHardLine(v.Contents)
	case doc.LineSuffixDoc:
		return containsHardLine(v.Contents)
	case doc.IfBreakDoc:
		return containsHardLine(v.Broken) || containsHardLine(v.Flat)
	}

	return false
}

func backslashContinue(text string) string {
	if !strings.Contains(text, "\n") {
		return text
	}

	lines := strings.Split(text, "\n")
	for i := range len(lines) - 1 {
		lines[i] += " \\"
	}

	return strings.Join(lines, "\n")
}
