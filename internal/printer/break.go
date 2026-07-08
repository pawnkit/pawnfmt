package printer

import "github.com/pawnkit/pawnfmt/internal/doc"

func hasForcedBreak(d doc.Doc) bool {
	switch node := d.(type) {
	case doc.HardLineDoc:
		return true
	case doc.BreakParentDoc:
		return true
	case doc.ConcatDoc:
		for _, part := range node.Parts {
			if hasForcedBreak(part) {
				return true
			}
		}
		return false
	case doc.IndentDoc:
		return hasForcedBreak(node.Contents)
	case doc.ResetIndentDoc:
		return hasForcedBreak(node.Contents)
	case doc.OutdentDoc:
		return hasForcedBreak(node.Contents)
	case doc.GroupDoc:
		return hasForcedBreak(node.Contents)
	case doc.FillDoc:
		for _, part := range node.Parts {
			if hasForcedBreak(part) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
