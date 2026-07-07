package printer

import (
	"unicode/utf8"

	"github.com/pawnkit/pawnfmt/internal/doc"
)

func fits(width int, next command, rest []command, mustBeFlat bool, _ Options) bool {
	local := []command{next}
	restIdx := len(rest) - 1

	pop := func() (command, bool) {
		if n := len(local); n > 0 {
			c := local[n-1]
			local = local[:n-1]
			return c, true
		}
		if restIdx >= 0 {
			c := rest[restIdx]
			restIdx--
			return c, true
		}
		return command{}, false
	}

	for width >= 0 {
		current, ok := pop()
		if !ok {
			break
		}

		switch node := current.doc.(type) {
		case nil:
		case doc.TextDoc:
			width -= utf8.RuneCountInString(node.Value)
		case doc.LineDoc:
			switch {
			case current.mode == modeFlat:
				width--
			case mustBeFlat:
				return false
			default:
				return true
			}
		case doc.SoftLineDoc:
			if current.mode == modeBreak {
				return !mustBeFlat
			}
		case doc.HardLineDoc:
			if mustBeFlat {
				return false
			}
			return true
		case doc.ConcatDoc:
			for index := len(node.Parts) - 1; index >= 0; index-- {
				local = append(local, command{indent: current.indent, mode: current.mode, doc: node.Parts[index]})
			}
		case doc.IndentDoc:
			local = append(local, command{indent: current.indent + 1, mode: current.mode, doc: node.Contents})
		case doc.ResetIndentDoc:
			local = append(local, command{indent: 0, mode: current.mode, doc: node.Contents})
		case doc.OutdentDoc:
			indent := max(current.indent-1, 0)
			local = append(local, command{indent: indent, mode: current.mode, doc: node.Contents})
		case doc.GroupDoc:
			local = append(local, command{indent: current.indent, mode: modeFlat, doc: node.Contents})
		case doc.IfBreakDoc:
			if current.mode == modeFlat {
				local = append(local, command{indent: current.indent, mode: current.mode, doc: node.Flat})
			} else {
				local = append(local, command{indent: current.indent, mode: current.mode, doc: node.Broken})
			}
		case doc.FillDoc:
			for index := len(node.Parts) - 1; index >= 0; index-- {
				local = append(local, command{indent: current.indent, mode: current.mode, doc: node.Parts[index]})
			}
		}
	}
	return width >= 0
}
