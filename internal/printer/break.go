// Package printer renders a doc.
package printer

import (
	"slices"

	"github.com/pawnkit/pawnfmt/internal/doc"
)

func hasForcedBreak(d doc.Doc) bool {
	switch node := d.(type) {
	case doc.HardLineDoc:
		return true
	case doc.BreakParentDoc:
		return true
	case doc.ConcatDoc:
		return slices.ContainsFunc(node.Parts, hasForcedBreak)
	case doc.IndentDoc:
		return hasForcedBreak(node.Contents)
	case doc.ResetIndentDoc:
		return hasForcedBreak(node.Contents)
	case doc.OutdentDoc:
		return hasForcedBreak(node.Contents)
	case doc.GroupDoc:
		return hasForcedBreak(node.Contents)
	case doc.FillDoc:
		return slices.ContainsFunc(node.Parts, hasForcedBreak)
	default:
		return false
	}
}
