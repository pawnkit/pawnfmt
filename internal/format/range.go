package format

import (
	"errors"
	"fmt"

	parser "github.com/pawnkit/pawn-parser"

	"github.com/pawnkit/pawnfmt/internal/config"
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

// FormatRange formats the single top-level syntax unit containing [start,end).
// Everything outside FormattedRange is preserved byte-for-byte.
func (formatter *Formatter) FormatRange(source []byte, start, end int) (RangeResult, error) {
	if start < 0 || end < start || end > len(source) {
		return RangeResult{}, fmt.Errorf("invalid format range [%d,%d) for source of %d bytes", start, end, len(source))
	}

	parsed, node, err := formatter.locateRangeTarget(source, start, end)
	if err != nil {
		return RangeResult{}, err
	}

	out, replaceStart, err := formatter.renderRangeReplacement(source, parsed, node)
	if err != nil {
		return RangeResult{}, err
	}

	if err := verifyRangeOutput(source, out, parsed); err != nil {
		return RangeResult{}, err
	}

	return RangeResult{Source: out, FormattedRange: Range{Start: replaceStart, End: node.End}}, nil
}

func (formatter *Formatter) locateRangeTarget(source []byte, start, end int) (*parser.File, *parser.Node, error) {
	parsed := parser.Parse(source)
	if parsed.HasParseErrors() && formatter.config.ParseMode == config.ParseModeStrict {
		return nil, nil, parseDiagnostic(source, parsed, "source")
	}

	if parsed.Root == nil {
		return nil, nil, errors.New("source has no syntax tree")
	}

	node, err := smallestRangeNode(source, parsed.Root, start, end)
	if err != nil {
		return nil, nil, err
	}

	return parsed, node, nil
}

func (formatter *Formatter) renderRangeReplacement(source []byte, parsed *parser.File, node *parser.Node) ([]byte, int, error) {
	rangeFormatter := *formatter
	rangeFormatter.config.SortIncludes = false

	fullyFormatted, err := rangeFormatter.FormatSource(source)
	if err != nil {
		return nil, 0, err
	}

	formattedParsed := parser.Parse(fullyFormatted)
	if formattedParsed.Root == nil {
		return nil, 0, errors.New("formatted syntax tree no longer matches the selected range")
	}

	formattedNode := correspondingRangeNode(parsed.Root, formattedParsed.Root, node)
	if formattedNode == nil {
		return nil, 0, errors.New("could not locate selected syntax in formatted output")
	}

	replaceStart := indentationStart(source, node.Start)

	formattedStart := indentationStart(fullyFormatted, formattedNode.Start)
	if replaceStart == node.Start {
		formattedStart = formattedNode.Start
	}

	replacement := fullyFormatted[formattedStart:formattedNode.End]
	out := make([]byte, 0, len(source)-(node.End-replaceStart)+len(replacement))
	out = append(out, source[:replaceStart]...)
	out = append(out, replacement...)
	out = append(out, source[node.End:]...)

	return out, replaceStart, nil
}

func verifyRangeOutput(source, out []byte, parsed *parser.File) error {
	verified := parser.Parse(out)
	if verified.HasParseErrors() && !parsed.HasParseErrors() {
		return parseDiagnostic(out, verified, "range-formatted output")
	}

	if err := verifySemanticTokens(source, out); err != nil {
		return fmt.Errorf("range-formatted output changed source semantics: %w", err)
	}

	if err := verifySemanticStructure(parsed, verified); err != nil {
		return fmt.Errorf("range-formatted output changed source structure: %w", err)
	}

	return nil
}

func smallestRangeNode(source []byte, root *parser.Node, start, end int) (*parser.Node, error) {
	overlaps := 0

	if start != end {
		for _, child := range root.Children {
			if child.End > start && indentationStart(source, child.Start) < end {
				overlaps++
			}
		}

		if overlaps > 1 {
			return nil, errors.New("format range crosses multiple top-level syntax units")
		}
	}

	var match *parser.Node

	for _, child := range root.Children {
		if rangeContainedByNode(source, child, start, end) {
			if match != nil {
				return nil, errors.New("format range crosses multiple top-level syntax units")
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
