package doc

type Doc interface {
	docNode()
}

type TextDoc struct {
	Value string
}

func (TextDoc) docNode() {}

type LineDoc struct{}

func (LineDoc) docNode() {}

type SoftLineDoc struct{}

func (SoftLineDoc) docNode() {}

type HardLineDoc struct{}

func (HardLineDoc) docNode() {}

type ConcatDoc struct {
	Parts []Doc
}

func (ConcatDoc) docNode() {}

type IndentDoc struct {
	Contents Doc
}

func (IndentDoc) docNode() {}

type ResetIndentDoc struct {
	Contents Doc
}

func (ResetIndentDoc) docNode() {}

type OutdentDoc struct {
	Contents Doc
}

func (OutdentDoc) docNode() {}

type GroupDoc struct {
	Contents Doc
}

func (GroupDoc) docNode() {}

type IfBreakDoc struct {
	Broken Doc
	Flat   Doc
}

func (IfBreakDoc) docNode() {}

type FillDoc struct {
	Parts []Doc
}

func (FillDoc) docNode() {}

// BreakParentDoc forces every enclosing Group to render in break mode
type BreakParentDoc struct{}

func (BreakParentDoc) docNode() {}

// LineSuffixDoc defers printing Contents until just before the next line
type LineSuffixDoc struct {
	Contents Doc
}

func (LineSuffixDoc) docNode() {}

func Text(value string) Doc {
	return TextDoc{Value: value}
}

func Line() Doc {
	return LineDoc{}
}

func SoftLine() Doc {
	return SoftLineDoc{}
}

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

func Indent(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}
	return IndentDoc{Contents: contents}
}

func ResetIndent(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}
	return ResetIndentDoc{Contents: contents}
}

func Outdent(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}
	return OutdentDoc{Contents: contents}
}

func Group(contents Doc) Doc {
	if contents == nil {
		return Text("")
	}
	return GroupDoc{Contents: contents}
}

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

func IfBreak(broken, flat Doc) Doc {
	return IfBreakDoc{Broken: broken, Flat: flat}
}

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
