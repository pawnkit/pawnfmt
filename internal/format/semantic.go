package format

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawn-parser/token"
)

type semanticToken struct {
	kind token.Kind
	text []byte
}

func verifySemanticTokens(before, after []byte) error {
	want := semanticTokens(before)
	got := semanticTokens(after)

	return compareSemanticTokenSlices(want, got)
}

func verifySemanticTokensWithSortedIncludes(beforeSource, afterSource []byte, before, after *parser.File) error {
	want := semanticTokensOutsideIncludes(beforeSource, before)

	got := semanticTokensOutsideIncludes(afterSource, after)
	if err := compareSemanticTokenSlices(want, got); err != nil {
		return err
	}

	wantIncludes := includeSignatures(before)

	gotIncludes := includeSignatures(after)
	if len(wantIncludes) != len(gotIncludes) {
		return fmt.Errorf("top-level include count changed from %d to %d", len(wantIncludes), len(gotIncludes))
	}

	for i := range wantIncludes {
		if wantIncludes[i] != gotIncludes[i] {
			return errors.New("top-level include set changed")
		}
	}

	return nil
}

func compareSemanticTokenSlices(want, got []semanticToken) error {
	wi, gi := 0, 0
	for wi < len(want) && gi < len(got) {
		if sameSemanticToken(want[wi], got[gi]) {
			wi++
			gi++

			continue
		}

		if compactModuloMatches(want[wi], got[gi:]) {
			wi++
			gi += 2

			continue
		}

		return fmt.Errorf("semantic token %d changed from %s %q to %s %q", wi+1,
			want[wi].kind, want[wi].text, got[gi].kind, got[gi].text)
	}

	if wi != len(want) || gi != len(got) {
		return fmt.Errorf("semantic token count changed from %d to %d", len(want), len(got))
	}

	return nil
}

func sameSemanticToken(want, got semanticToken) bool {
	return want.kind == got.kind && equalSemanticText(want.kind, want.text, got.text)
}

func compactModuloMatches(want semanticToken, got []semanticToken) bool {
	return want.kind == token.MacroParam && len(want.text) == 2 && want.text[0] == '%' &&
		want.text[1] >= '0' && want.text[1] <= '9' && len(got) >= 2 &&
		got[0].kind == token.Percent && bytes.Equal(got[0].text, []byte("%")) &&
		got[1].kind == token.IntLiteral && bytes.Equal(got[1].text, want.text[1:])
}

func semanticTokensOutsideIncludes(source []byte, file *parser.File) []semanticToken {
	if file == nil || file.Root == nil {
		return semanticTokens(source)
	}

	var spans [][2]int

	for _, child := range file.Root.Children {
		if isIncludeLike(child.Kind) {
			spans = append(spans, [2]int{child.Start, child.End})
		}
	}

	tokens := lexer.Tokenize(source)

	out := make([]semanticToken, 0, len(tokens))
	for _, tok := range tokens {
		if nonSemanticFormattingToken(tok.Kind) || offsetInSpans(tok.Start.Offset, spans) {
			continue
		}

		out = append(out, semanticToken{kind: tok.Kind, text: append([]byte(nil), tok.Text(source)...)})
	}

	return out
}

func offsetInSpans(offset int, spans [][2]int) bool {
	for _, span := range spans {
		if span[0] <= offset && offset < span[1] {
			return true
		}
	}

	return false
}

func includeSignatures(file *parser.File) []string {
	if file == nil || file.Root == nil {
		return nil
	}

	var signatures []string

	for _, child := range file.Root.Children {
		if isIncludeLike(child.Kind) {
			signatures = append(signatures, child.Kind.String()+"\x00"+includeSortKey(child))
		}
	}

	sort.Strings(signatures)

	return signatures
}

func verifySemanticStructure(before, after *parser.File) error {
	if before == nil || before.Root == nil || after == nil || after.Root == nil {
		return errors.New("cannot compare missing syntax trees")
	}

	return compareSemanticNodes(before.Root, after.Root, "source_file")
}

func compareSemanticNodes(before, after *parser.Node, path string) error {
	if before == nil || after == nil {
		if before == after {
			return nil
		}

		return fmt.Errorf("syntax tree changed at %s", path)
	}

	if before.Kind != after.Kind {
		return compareMismatchedKindNodes(before, after, path)
	}

	if before.Tok.Kind != after.Tok.Kind && !equivalentOperatorKinds(before, after) {
		return fmt.Errorf("operator at %s changed from %s to %s", path, before.Tok.Kind, after.Tok.Kind)
	}

	if before.Kind == parser.KindRaw && !bytes.Equal(before.Raw, after.Raw) {
		return fmt.Errorf("raw syntax at %s changed", path)
	}

	if len(before.Children) != len(after.Children) {
		return fmt.Errorf("%s at %s changed child count from %d to %d",
			before.Kind, path, len(before.Children), len(after.Children))
	}

	return compareChildNodes(before, after, path)
}

func equivalentOperatorKinds(before, after *parser.Node) bool {
	return before.Kind == parser.KindBinaryExpression && after.Kind == parser.KindBinaryExpression &&
		before.Tok.Kind == token.MacroParam && after.Tok.Kind == token.Percent
}

func compareMismatchedKindNodes(before, after *parser.Node, path string) error {
	unwrappedBefore, beforeWasBlock := unwrapSingleStatementBlock(before)

	unwrappedAfter, afterWasBlock := unwrapSingleStatementBlock(after)
	if beforeWasBlock != afterWasBlock {
		return compareSemanticNodes(unwrappedBefore, unwrappedAfter, path)
	}

	return fmt.Errorf("syntax node at %s changed from %s to %s", path, before.Kind, after.Kind)
}

func compareChildNodes(before, after *parser.Node, path string) error {
	for i := 0; i < len(before.Children); {
		if isIncludeLike(before.Children[i].Kind) && isIncludeLike(after.Children[i].Kind) {
			beforeEnd := includeRunEnd(before.Children, i)

			afterEnd := includeRunEnd(after.Children, i)
			if err := compareIncludeNodeRuns(before.Children[i:beforeEnd], after.Children[i:afterEnd], path); err != nil {
				return err
			}

			i = beforeEnd

			continue
		}

		childPath := fmt.Sprintf("%s/%s[%d]", path, before.Kind, i)
		if err := compareSemanticNodes(before.Children[i], after.Children[i], childPath); err != nil {
			return err
		}

		i++
	}

	return nil
}

func includeRunEnd(children []*parser.Node, start int) int {
	end := start
	for end < len(children) && isIncludeLike(children[end].Kind) {
		end++
	}

	return end
}

func compareIncludeNodeRuns(before, after []*parser.Node, path string) error {
	if len(before) != len(after) {
		return fmt.Errorf("include run at %s changed length from %d to %d", path, len(before), len(after))
	}

	want := make([]string, len(before))
	got := make([]string, len(after))

	for i, node := range before {
		want[i] = node.Kind.String() + "\x00" + includeSortKey(node)
	}

	for i, node := range after {
		got[i] = node.Kind.String() + "\x00" + includeSortKey(node)
	}

	sort.Strings(want)
	sort.Strings(got)

	for i := range want {
		if want[i] != got[i] {
			return fmt.Errorf("include run at %s changed contents", path)
		}
	}

	return nil
}

func unwrapSingleStatementBlock(n *parser.Node) (*parser.Node, bool) {
	if n != nil && n.Kind == parser.KindBlock && len(n.Children) == 1 {
		return n.Children[0], true
	}

	return n, false
}

func equalSemanticText(kind token.Kind, want, got []byte) bool {
	if kind == token.IntLiteral || kind == token.FloatLiteral {
		return bytes.EqualFold(want, got)
	}

	return bytes.Equal(want, got)
}

func semanticTokens(source []byte) []semanticToken {
	tokens := lexer.Tokenize(source)

	out := make([]semanticToken, 0, len(tokens))
	for _, tok := range tokens {
		if nonSemanticFormattingToken(tok.Kind) {
			continue
		}

		out = append(out, semanticToken{kind: tok.Kind, text: append([]byte(nil), tok.Text(source)...)})
	}

	return out
}

func nonSemanticFormattingToken(kind token.Kind) bool {
	//nolint:exhaustive // only the purely-cosmetic token kinds matter here
	switch kind {
	case token.EOF, token.LBrace, token.RBrace, token.Comma, token.Semicolon:
		return true
	default:
		return false
	}
}
