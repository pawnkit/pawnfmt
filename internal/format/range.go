package format

import (
	"bytes"
	"errors"
	"fmt"

	parser "github.com/pawnkit/pawn-parser"

	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
	"github.com/pawnkit/pawnfmt/internal/printer"
	"github.com/pawnkit/pawnfmt/internal/trivia"
)

// Range is a half-open byte range in Pawn source.
type Range struct {
	Start int
	End   int
}

// RangeResult contains range-formatted source and the complete syntax range
// that was formatted.
type RangeResult struct {
	Source         []byte
	FormattedRange Range
}

// FormatRange formats the syntax units intersecting [start,end).
// Everything outside FormattedRange is preserved byte-for-byte.
func (formatter *Formatter) FormatRange(source []byte, start, end int) (RangeResult, error) {
	if start < 0 || end < start || end > len(source) {
		return RangeResult{}, fmt.Errorf("invalid format range [%d,%d) for source of %d bytes", start, end, len(source))
	}

	parsed, nodes, err := formatter.locateRangeTargets(source, start, end)
	if err != nil {
		return RangeResult{}, err
	}

	if len(nodes) > 1 {
		return formatter.renderTopLevelRange(source, parsed, nodes)
	}

	out, replaceStart, err := formatter.renderRangeReplacement(source, parsed, nodes[0])
	if err != nil {
		return RangeResult{}, err
	}

	return RangeResult{
		Source: out,
		FormattedRange: Range{
			Start: replaceStart,
			End:   nodes[0].End,
		},
	}, nil
}

func (formatter *Formatter) renderTopLevelRange(source []byte, parsed *parser.File, nodes []*parser.Node) (RangeResult, error) {
	rangeFormatter := *formatter
	rangeFormatter.config.SortIncludes = false

	st := newState(parsed, rangeFormatter.config, trivia.Scan(source))
	st.topLevelContext = true
	widths := st.alignmentWidths(nodes)
	macroWidths := st.macroAlignmentWidths(nodes)
	st.applyCommentAlignment(nodes)

	parts := make([]doc.Doc, 0, len(nodes)*2)
	for index, node := range nodes {
		if index > 0 {
			parts = append(parts, st.topLevelSeparator(nodes[index-1], node))
		}

		st.hint.alignDeclarationWidth = widths[node]
		st.hint.alignMacroValueWidth = macroWidths[node]
		parts = append(parts, st.formatNode(node))
	}

	options := st.printerOptions()
	options.InsertFinalNewline = false
	replacement := []byte(printer.Print(doc.Concat(parts...), options))
	replaceStart := indentationStart(source, nodes[0].Start)
	replaceEnd := nodes[len(nodes)-1].End
	original := source[replaceStart:replaceEnd]

	if err := verifySemanticTokens(original, replacement); err != nil {
		return RangeResult{}, fmt.Errorf("range-formatted output changed source semantics: %w", err)
	}

	originalParsed := parser.Parse(original)
	formattedParsed := parser.Parse(replacement)

	if err := compareSemanticNodes(originalParsed.Root, formattedParsed.Root, "range"); err != nil {
		return RangeResult{}, fmt.Errorf("range-formatted output changed source structure: %w", err)
	}

	out := make([]byte, 0, len(source)-len(original)+len(replacement))
	out = append(out, source[:replaceStart]...)
	out = append(out, replacement...)
	out = append(out, source[replaceEnd:]...)

	return RangeResult{Source: out, FormattedRange: Range{Start: replaceStart, End: replaceEnd}}, nil
}

func (formatter *Formatter) locateRangeTargets(source []byte, start, end int) (*parser.File, []*parser.Node, error) {
	parsed := parser.Parse(source)
	if parsed.HasParseErrors() && formatter.config.ParseMode == config.ParseModeStrict {
		return nil, nil, parseDiagnostic(source, parsed, "source")
	}

	if parsed.Root == nil {
		return nil, nil, errors.New("source has no syntax tree")
	}

	nodes, err := rangeNodes(source, parsed.Root, start, end)
	if err != nil {
		return nil, nil, err
	}

	nodes = formatter.expandMacroAlignmentRange(source, parsed, nodes)

	return parsed, nodes, nil
}

func (formatter *Formatter) expandMacroAlignmentRange(source []byte, parsed *parser.File, nodes []*parser.Node) []*parser.Node {
	if !formatter.config.AlignConsecutiveMacros || len(nodes) != 1 || parsed.Root == nil {
		return nodes
	}

	selected := topLevelRangeParent(parsed.Root, nodes[0])
	if !alignableMacro(selected) {
		return nodes
	}

	index := topLevelNodeIndex(parsed.Root, selected)
	if index < 0 {
		return nodes
	}

	st := newState(parsed, formatter.config, trivia.Scan(source))
	first, last := macroGroupBounds(st, parsed.Root.Children, index)

	return parsed.Root.Children[first : last+1]
}

func topLevelNodeIndex(root, selected *parser.Node) int {
	for index, child := range root.Children {
		if child == selected {
			return index
		}
	}

	return -1
}

func macroGroupBounds(st *state, children []*parser.Node, index int) (int, int) {
	first, last := index, index

	for first > 0 && alignableMacro(children[first-1]) &&
		st.blankLinesBefore(children[first].LeadingTrivia()) == 0 &&
		!hasCommentTrivia(children[first].LeadingTrivia()) {
		first--
	}

	for last+1 < len(children) && alignableMacro(children[last+1]) &&
		st.blankLinesBefore(children[last+1].LeadingTrivia()) == 0 &&
		!hasCommentTrivia(children[last+1].LeadingTrivia()) {
		last++
	}

	return first, last
}

func alignableMacro(node *parser.Node) bool {
	return node != nil && node.Kind == parser.KindDirectiveDefine && node.Field("value") != nil
}

func rangeNodes(source []byte, root *parser.Node, start, end int) ([]*parser.Node, error) {
	var overlaps []*parser.Node

	if start != end {
		for _, child := range root.Children {
			if child.End > start && indentationStart(source, child.Start) < end {
				overlaps = append(overlaps, child)
			}
		}
	}

	if len(overlaps) > 1 {
		return overlaps, nil
	}

	node, err := smallestRangeNode(source, root, start, end)
	if err != nil {
		return nil, err
	}

	return []*parser.Node{node}, nil
}

func (formatter *Formatter) renderRangeReplacement(source []byte, parsed *parser.File, node *parser.Node) ([]byte, int, error) {
	rangeFormatter := *formatter
	rangeFormatter.config.SortIncludes = false

	parent := topLevelRangeParent(parsed.Root, node)
	if parent == nil {
		return nil, 0, errors.New("could not locate selected syntax unit")
	}

	replaceStart := indentationStart(source, node.Start)
	st := newState(parsed, rangeFormatter.config, trivia.Scan(source))
	formattedParentSource := []byte(printer.Print(st.formatNode(parent), st.printerOptions()))

	formattedParsed := parser.Parse(formattedParentSource)
	if formattedParsed.Root == nil || len(formattedParsed.Root.Children) == 0 {
		return nil, 0, errors.New("formatted syntax tree no longer matches the selected range")
	}

	formattedParent := formattedParsed.Root.Children[0]

	if err := verifySemanticTokens(source[parent.Start:parent.End], formattedParentSource); err != nil {
		return nil, 0, fmt.Errorf("range-formatted output changed source semantics: %w", err)
	}

	if err := compareSemanticNodes(parent, formattedParent, "range"); err != nil {
		return nil, 0, fmt.Errorf("range-formatted output changed source structure: %w", err)
	}

	formattedNode := correspondingRangeNode(parent, formattedParent, node)
	if formattedNode == nil {
		return nil, 0, errors.New("could not locate selected syntax in formatted output")
	}

	formattedStart := indentationStart(formattedParentSource, formattedNode.Start)
	if replaceStart == node.Start {
		formattedStart = formattedNode.Start
	}

	replacement := formattedParentSource[formattedStart:formattedNode.End]
	if replaceStart < node.Start {
		replacement = indentReplacement(replacement, source[replaceStart:node.Start])
	}

	out := make([]byte, 0, len(source)-(node.End-replaceStart)+len(replacement))
	out = append(out, source[:replaceStart]...)
	out = append(out, replacement...)
	out = append(out, source[node.End:]...)

	return out, replaceStart, nil
}

func topLevelRangeParent(root, selected *parser.Node) *parser.Node {
	if root == nil || selected == nil {
		return nil
	}

	for _, child := range root.Children {
		if child.Start <= selected.Start && selected.End <= child.End {
			return child
		}
	}

	return nil
}

func indentReplacement(source, indent []byte) []byte {
	if len(indent) == 0 || len(source) == 0 {
		return source
	}

	out := make([]byte, 0, len(source)+bytes.Count(source, []byte{'\n'})*len(indent)+len(indent))

	out = append(out, indent...)

	for index, value := range source {
		out = append(out, value)
		if value == '\n' && index+1 < len(source) {
			out = append(out, indent...)
		}
	}

	return out
}

func smallestRangeNode(source []byte, root *parser.Node, start, end int) (*parser.Node, error) {
	var match *parser.Node

	for _, child := range root.Children {
		if rangeContainedByNode(source, child, start, end) {
			if match != nil {
				return nil, errors.New("format range is ambiguous")
			}

			match = child
		}
	}

	if match == nil {
		return nil, errors.New("format range does not select a complete syntax unit")
	}

	selected := deepestRangeNode(source, match, start, end)
	if selected == nil {
		return nil, errors.New("format range does not select a format-safe syntax unit")
	}

	return selected, nil
}

func deepestRangeNode(source []byte, node *parser.Node, start, end int) *parser.Node {
	if rangeBoundaryKind(node.Kind) {
		return node
	}

	for _, child := range node.Children {
		if rangeContainedByNode(source, child, start, end) {
			if candidate := deepestRangeNode(source, child, start, end); candidate != nil {
				return candidate
			}
		}
	}

	if rangeFormatKind(node.Kind) {
		return node
	}

	return nil
}

func rangeContainedByNode(source []byte, node *parser.Node, start, end int) bool {
	nodeStart := indentationStart(source, node.Start)
	if start == end {
		return nodeStart <= start && start < node.End
	}

	return nodeStart <= start && end <= node.End
}

func rangeBoundaryKind(kind parser.Kind) bool {
	return kind == parser.KindConditionalRegion || kind == parser.KindSharedConditional ||
		kind == parser.KindSharedConditionalPrefix || kind == parser.KindConditionalSplice
}

func rangeFormatKind(kind parser.Kind) bool {
	if parser.IsTopLevelDeclaration(kind) || kind.IsDirective() {
		return true
	}

	//nolint:exhaustive // deliberate allowlist of range-formattable statement kinds; default covers the rest
	switch kind {
	case parser.KindRaw, parser.KindVariableDeclaration, parser.KindBlock, parser.KindIfStatement,
		parser.KindWhileStatement, parser.KindDoWhileStatement, parser.KindForStatement,
		parser.KindSwitchStatement, parser.KindCaseClause, parser.KindDefaultClause,
		parser.KindGotoStatement, parser.KindLabelStatement, parser.KindReturnStatement,
		parser.KindBreakStatement, parser.KindContinueStatement, parser.KindStateStatement,
		parser.KindExpressionStatement, parser.KindEmptyStatement, parser.KindMacroInvocationBlock,
		parser.KindParameterList, parser.KindArgumentList, parser.KindArrayLiteral:
		return true
	default:
		return false
	}
}

func correspondingRangeNode(before, after, target *parser.Node) *parser.Node {
	if before == nil || after == nil {
		return nil
	}

	if before == target {
		return after
	}

	if before.Kind != after.Kind {
		if before.Kind == parser.KindBlock && len(before.Children) == 1 {
			return correspondingRangeNode(before.Children[0], after, target)
		}

		if after.Kind == parser.KindBlock && len(after.Children) == 1 {
			return correspondingRangeNode(before, after.Children[0], target)
		}

		return nil
	}

	if len(before.Children) != len(after.Children) {
		return nil
	}

	for i := range before.Children {
		if mapped := correspondingRangeNode(before.Children[i], after.Children[i], target); mapped != nil {
			return mapped
		}
	}

	return nil
}

func indentationStart(source []byte, offset int) int {
	start := offset
	for start > 0 && source[start-1] != '\n' {
		start--
	}

	for i := start; i < offset; i++ {
		if source[i] != ' ' && source[i] != '\t' && source[i] != '\r' {
			return offset
		}
	}

	return start
}

// SourceRange is a convenience wrapper around New and Formatter.FormatRange.
func SourceRange(source []byte, cfg config.Config, start, end int) (RangeResult, error) {
	formatter, err := New(cfg)
	if err != nil {
		return RangeResult{}, err
	}

	return formatter.FormatRange(source, start, end)
}
