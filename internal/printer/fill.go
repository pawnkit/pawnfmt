package printer

import "github.com/pawnkit/pawnfmt/internal/doc"

func fillCommands(parts []doc.Doc, indent, remaining int, options Options) []command {
	if len(parts) == 0 {
		return nil
	}
	content := parts[0]
	contentFits := fits(remaining, command{indent: indent, mode: modeFlat, doc: content}, nil, true, options)
	contentMode := modeBreak
	if contentFits {
		contentMode = modeFlat
	}
	if len(parts) == 1 {
		return []command{{indent: indent, mode: contentMode, doc: content}}
	}

	separator := parts[1]
	if len(parts) == 2 {
		return []command{
			{indent: indent, mode: contentMode, doc: separator},
			{indent: indent, mode: contentMode, doc: content},
		}
	}

	secondContent := parts[2]
	pairFits := fits(remaining, command{indent: indent, mode: modeFlat, doc: doc.Concat(content, separator, secondContent)}, nil, true, options)
	remainingCmd := command{indent: indent, mode: modeBreak, doc: doc.FillDoc{Parts: parts[2:]}}

	switch {
	case pairFits:
		return []command{remainingCmd, {indent: indent, mode: modeFlat, doc: separator}, {indent: indent, mode: modeFlat, doc: content}}
	case contentFits:
		return []command{remainingCmd, {indent: indent, mode: modeBreak, doc: separator}, {indent: indent, mode: modeFlat, doc: content}}
	default:
		return []command{remainingCmd, {indent: indent, mode: modeBreak, doc: separator}, {indent: indent, mode: modeBreak, doc: content}}
	}
}
