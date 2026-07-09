package format

import (
	"bytes"
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
			return fmt.Errorf("top-level include set changed")
		}
	}

	return nil
}

func compareSemanticTokenSlices(want, got []semanticToken) error {

	limit := min(len(got), len(want))
	for i := range limit {
		if want[i].kind != got[i].kind || !equalSemanticText(want[i].kind, want[i].text, got[i].text) {
			return fmt.Errorf("semantic token %d changed from %s %q to %s %q", i+1,
				want[i].kind, want[i].text, got[i].kind, got[i].text)
		}
	}

	if len(want) != len(got) {
		return fmt.Errorf("semantic token count changed from %d to %d", len(want), len(got))
	}

	return nil
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
		return fmt.Errorf("cannot compare missing syntax trees")
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
		unwrappedBefore, beforeWasBlock := unwrapSingleStatementBlock(before)
		unwrappedAfter, afterWasBlock := unwrapSingleStatementBlock(after)
		if beforeWasBlock != afterWasBlock {
			return compareSemanticNodes(unwrappedBefore, unwrappedAfter, path)
		}

		return fmt.Errorf("syntax node at %s changed from %s to %s", path, before.Kind, after.Kind)
	}

	if before.Tok.Kind != after.Tok.Kind {
		return fmt.Errorf("operator at %s changed from %s to %s", path, before.Tok.Kind, after.Tok.Kind)
	}
	if before.Kind == parser.KindRaw && !bytes.Equal(before.Raw, after.Raw) {
		return fmt.Errorf("raw syntax at %s changed", path)
	}

	if len(before.Children) != len(after.Children) {
		return fmt.Errorf("%s at %s changed child count from %d to %d",
			before.Kind, path, len(before.Children), len(after.Children))
	}

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
