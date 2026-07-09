// Package doc implements a Wadler/Prettier-style pretty-printing IR.
package doc

// Doc is a node in the pretty-printing document tree.
type Doc interface {
	docNode()
}

// TextDoc renders literal text verbatim.
type TextDoc struct {
	Value string
}

func (TextDoc) docNode() {}

// LineDoc renders a space in flat mode, or a newline in break mode.
type LineDoc struct{}

func (LineDoc) docNode() {}

// SoftLineDoc renders nothing in flat mode, or a newline in break mode.
type SoftLineDoc struct{}

func (SoftLineDoc) docNode() {}

// HardLineDoc always renders a newline and forces its enclosing Group to break.
type HardLineDoc struct{}

func (HardLineDoc) docNode() {}

// ConcatDoc renders its parts in sequence.
type ConcatDoc struct {
	Parts []Doc
}

func (ConcatDoc) docNode() {}

// IndentDoc renders Contents one indent level deeper.
type IndentDoc struct {
	Contents Doc
}

func (IndentDoc) docNode() {}

// ResetIndentDoc renders Contents at column zero, ignoring ambient indent.
type ResetIndentDoc struct {
	Contents Doc
}

func (ResetIndentDoc) docNode() {}

// OutdentDoc renders Contents one indent level shallower.
type OutdentDoc struct {
	Contents Doc
}

func (OutdentDoc) docNode() {}

// GroupDoc renders Contents flat if it fits on the line, else breaks it.
type GroupDoc struct {
	Contents Doc
}

func (GroupDoc) docNode() {}

// IfBreakDoc renders Broken when its enclosing Group breaks, else Flat.
type IfBreakDoc struct {
	Broken Doc
	Flat   Doc
}

func (IfBreakDoc) docNode() {}

// FillDoc packs parts onto lines, breaking only where a part wouldn't fit.
type FillDoc struct {
	Parts []Doc
}

func (FillDoc) docNode() {}

// BreakParentDoc forces every enclosing Group to render in break mode.
type BreakParentDoc struct{}

func (BreakParentDoc) docNode() {}

// LineSuffixDoc defers printing Contents until just before the next line.
type LineSuffixDoc struct {
	Contents Doc
}

func (LineSuffixDoc) docNode() {}

// Text renders value verbatim.
func Text(value string) Doc {
	return TextDoc{Value: value}
}

// Line renders a space in flat mode, or a newline in break mode.
func Line() Doc {
	return LineDoc{}
}

// SoftLine renders nothing in flat mode, or a newline in break mode.
func SoftLine() Doc {
	return SoftLineDoc{}
}

// HardLine always renders a newline and forces its enclosing Group to break.
func HardLine() Doc {
	return HardLineDoc{}
}

// BreakParent forces every enclosing Group to render in break mode. See
// BreakParentDoc.
func BreakParent() Doc {
	return BreakParentDoc{}
}

// LineSuffix defers contents until the next line break the printer emits.
// See LineSuffixDoc.
func LineSuffix(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}

	return LineSuffixDoc{Contents: contents}
}

// Concat renders parts in sequence, dropping any nil entries.
func Concat(parts ...Doc) Doc {
	filtered := make([]Doc, 0, len(parts))
	for _, part := range parts {
		if part != nil {
			filtered = append(filtered, part)
		}
	}

	if len(filtered) == 0 {
		return Text("")
	}

	if len(filtered) == 1 {
		return filtered[0]
	}

	return ConcatDoc{Parts: filtered}
}

// Indent renders contents one indent level deeper.
func Indent(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}

	return IndentDoc{Contents: contents}
}

// ResetIndent renders contents at column zero, ignoring ambient indent.
func ResetIndent(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}

	return ResetIndentDoc{Contents: contents}
}

// Outdent renders contents one indent level shallower.
func Outdent(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}

	return OutdentDoc{Contents: contents}
}

// Group renders contents flat if it fits on the line, else breaks it.
func Group(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}

	return GroupDoc{Contents: contents}
}

// Join concatenates parts with separator between each, dropping nil entries.
func Join(separator Doc, parts ...Doc) Doc {
	filtered := make([]Doc, 0, len(parts))
	for _, part := range parts {
		if part != nil {
			filtered = append(filtered, part)
		}
	}

	if len(filtered) == 0 {
		return Text("")
	}

	out := make([]Doc, 0, len(filtered)*2-1)
	for index, part := range filtered {
		if index > 0 && separator != nil {
			out = append(out, separator)
		}

		out = append(out, part)
	}

	return Concat(out...)
}

// IfBreak renders broken when its enclosing Group breaks, else flat.
func IfBreak(broken, flat Doc) Doc {
	return IfBreakDoc{Broken: broken, Flat: flat}
}

// Fill packs parts onto lines, breaking only where a part wouldn't fit.
func Fill(parts ...Doc) Doc {
	filtered := make([]Doc, 0, len(parts))
	for _, part := range parts {
		if part != nil {
			filtered = append(filtered, part)
		}
	}

	if len(filtered) == 0 {
		return Text("")
	}

	if len(filtered) == 1 {
		return filtered[0]
	}

	return FillDoc{Parts: filtered}
}

// RawTextBlock renders value verbatim line-by-line at column zero.
func RawTextBlock(value string) Doc {
	if value == "" {
		return Text("")
	}

	parts := make([]Doc, 0, len(value)/8+1)
	start := 0

	for index := range len(value) {
		if value[index] != '\n' {
			continue
		}

		line := value[start:index]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		parts = append(parts, Text(line), HardLine())
		start = index + 1
	}

	if start < len(value) {
		parts = append(parts, Text(value[start:]))
	}

	if len(parts) == 0 {
		return Text("")
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return ResetIndent(Concat(parts...))
}
