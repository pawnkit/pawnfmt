package format

import (
	"sort"
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/token"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatSourceFile(n *parser.Node) doc.Doc {
	items := n.Children
	if s.config.SortIncludes {
		items = sortIncludeRuns(s.source, items, s.config.GroupIncludesByBrackets,
			func(item *parser.Node) bool { return !s.isDisabled(item) })
	}

	s.topLevelContext = true

	defer func() { s.topLevelContext = false }()

	widths := s.alignmentWidths(items)
	macroWidths := s.macroAlignmentWidths(items)
	s.applyCommentAlignment(items)

	var (
		parts []doc.Doc
		prev  *parser.Node
	)

	for _, item := range items {
		if prev != nil {
			parts = append(parts, s.topLevelSeparator(prev, item))
		}

		s.hint.alignDeclarationWidth = widths[item]
		s.hint.alignMacroValueWidth = macroWidths[item]
		parts = append(parts, s.formatNode(item))
		prev = item
	}

	if trailing := s.leadingDocs(n.Trailing); len(trailing) > 0 {
		if len(parts) > 0 {
			parts = append(parts, doc.HardLine())
		}

		parts = append(parts, trailing...)
		parts = trimTrailingHardLine(parts)
	}

	return doc.Concat(parts...)
}

func trimTrailingHardLine(parts []doc.Doc) []doc.Doc {
	if len(parts) == 0 {
		return parts
	}

	if _, ok := parts[len(parts)-1].(doc.HardLineDoc); ok {
		return parts[:len(parts)-1]
	}

	return parts
}

func isIncludeLike(k parser.Kind) bool {
	return k == parser.KindDirectiveInclude || k == parser.KindDirectiveTryInclude
}

func (s *state) isPublicFunction(n *parser.Node) bool {
	if n == nil || n.Kind != parser.KindFunctionDefinition {
		return false
	}

	storage := n.Field("storage")

	return storage != nil && storage.Text(s.source) == "public"
}

func (s *state) topLevelSeparator(prev, cur *parser.Node) doc.Doc {
	if s.isDisabled(prev) || s.isDisabled(cur) {
		return blankLineSeparator(s.blankLinesBefore(cur.LeadingTrivia()))
	}

	force := s.config.EmptyLineBetweenTopLevelDecls && shouldSeparateTopLevelDeclarations(prev.Kind, cur.Kind)
	if isTopLevelGroupBoundary(prev.Kind, cur.Kind) {
		force = true
	}

	if s.config.BlankLinesBetweenPublics && s.isPublicFunction(prev) && s.isPublicFunction(cur) {
		force = true
	}

	if s.config.BlankLinesAfterIncludeBlock && isIncludeLike(prev.Kind) && !isIncludeLike(cur.Kind) {
		force = true
	}

	if force {
		return blankLineSeparator(1)
	}

	return blankLineSeparator(s.blankLinesBefore(cur.LeadingTrivia()))
}

func shouldSeparateTopLevelDeclarations(prev, cur parser.Kind) bool {
	if !parser.IsTopLevelDeclaration(prev) || !parser.IsTopLevelDeclaration(cur) {
		return false
	}

	if prev != cur {
		return true
	}

	return prev == parser.KindFunctionDefinition || prev == parser.KindEnumDeclaration
}

func isTopLevelGroupBoundary(prev, cur parser.Kind) bool {
	if isIncludeLike(prev) || isIncludeLike(cur) {
		return false
	}

	prevDecl := parser.IsTopLevelDeclaration(prev)

	curDecl := parser.IsTopLevelDeclaration(cur)
	if prevDecl == curDecl {
		return false
	}

	return prev.IsDirective() || cur.IsDirective() ||
		prev == parser.KindConditionalRegion || cur == parser.KindConditionalRegion
}

func sortIncludeRuns(source []byte, items []*parser.Node, groupByBrackets bool, eligible func(*parser.Node) bool) []*parser.Node {
	out := make([]*parser.Node, len(items))
	copy(out, items)

	i := 0
	for i < len(out) {
		if !isIncludeLike(out[i].Kind) || !eligible(out[i]) {
			i++
			continue
		}

		j := i + 1
		for j < len(out) && isIncludeLike(out[j].Kind) && eligible(out[j]) &&
			!hasBlankLineBetween(source, out[j-1], out[j]) {
			j++
		}

		run := out[i:j]
		if includeRunHasDuplicatePath(run) {
			i = j
			continue
		}

		var (
			runHeader     []token.Trivia
			originalFirst *parser.Node
		)

		startsAfterBlankIncludeGroup := i > 0 && isIncludeLike(items[i-1].Kind) &&
			hasBlankLineBetween(source, items[i-1], items[i])
		if i == 0 && hasCommentTrivia(run[0].Leading) || startsAfterBlankIncludeGroup {
			runHeader = append([]token.Trivia(nil), run[0].Leading...)
			originalFirst = run[0]
			clone := *originalFirst
			clone.Leading = nil
			run[0] = &clone
		}

		sort.SliceStable(run, func(a, b int) bool {
			if groupByBrackets {
				ag, bg := includeBracketGroup(run[a]), includeBracketGroup(run[b])
				if ag != bg {
					return ag < bg
				}
			}

			return includeSortKey(run[a]) < includeSortKey(run[b])
		})

		if len(runHeader) > 0 {
			clone := *run[0]
			clone.Leading = append(runHeader, clone.Leading...)
			run[0] = &clone
		}

		i = j
	}

	return out
}

func hasBlankLineBetween(source []byte, before, after *parser.Node) bool {
	if before == nil || after == nil || before.End < 0 || after.Start > len(source) || before.End > after.Start {
		return true
	}

	between := strings.ReplaceAll(string(source[before.End:after.Start]), "\r\n", "\n")
	lines := strings.Split(between, "\n")
	for i := 1; i < len(lines)-1; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			return true
		}
	}

	return false
}

func includeRunHasDuplicatePath(run []*parser.Node) bool {
	seen := make(map[string]struct{}, len(run))
	for _, item := range run {
		key := includeSortKey(item)
		if _, exists := seen[key]; exists {
			return true
		}
		seen[key] = struct{}{}
	}

	return false
}

func hasCommentTrivia(items []token.Trivia) bool {
	for _, item := range items {
		if item.Kind == token.Comment {
			return true
		}
	}

	return false
}

func includeSortKey(n *parser.Node) string {
	path := n.Field("path")
	if path == nil {
		return ""
	}

	return strings.TrimSpace(string(path.Raw))
}

func includeBracketGroup(n *parser.Node) int {
	if strings.HasPrefix(includeSortKey(n), "<") {
		return 0
	}

	return 1
}
