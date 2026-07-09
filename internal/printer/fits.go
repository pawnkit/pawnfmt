package printer

import (
	"unicode/utf8"

	"github.com/pawnkit/pawnfmt/internal/doc"
)

// fitsLineDoc handles a Line doc during fits: it either consumes one column
// of width (flat mode) or immediately decides the outcome (break mode).
func fitsLineDoc(mode mode, mustBeFlat bool, width int) (newWidth int, stop, result bool) {
	switch {
	case mode == modeFlat:
		return width - 1, false, false
	case mustBeFlat:
		return width, true, false
	default:
		return width, true, true
	}
}

// fitsStack pops from a local (mutable) stack of commands, falling back to
// the printer's outer command stack once local is exhausted.
type fitsStack struct {
	local   []command
	rest    []command
	restIdx int
}

func (s *fitsStack) pop() (command, bool) {
	if n := len(s.local); n > 0 {
		c := s.local[n-1]
		s.local = s.local[:n-1]

		return c, true
	}

	if s.restIdx >= 0 {
		c := s.rest[s.restIdx]
		s.restIdx--

		return c, true
	}

	return command{}, false
}

func fits(width int, next command, rest []command, mustBeFlat bool, _ Options) bool {
	stack := &fitsStack{local: []command{next}, rest: rest, restIdx: len(rest) - 1}

	for width >= 0 {
		current, ok := stack.pop()
		if !ok {
			break
		}

		switch node := current.doc.(type) {
		case nil:
		case doc.TextDoc:
			width -= utf8.RuneCountInString(node.Value)
		case doc.LineDoc:
			w, stop, result := fitsLineDoc(current.mode, mustBeFlat, width)
			if stop {
				return result
			}

			width = w
		case doc.SoftLineDoc:
			if current.mode == modeBreak {
				return !mustBeFlat
			}
		case doc.HardLineDoc:
			if mustBeFlat {
				return false
			}

			return true
		case doc.BreakParentDoc:
		case doc.LineSuffixDoc:
			// skip
		default:
			stack.local = pushFitsChildren(stack.local, current, node)
		}
	}

	return width >= 0
}

// pushFitsChildren pushes the child commands of a container Doc (Concat,
// Indent, Group, ...) onto local for fits' iterative traversal.
func pushFitsChildren(local []command, current command, node doc.Doc) []command {
	switch node := node.(type) {
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

	return local
}
