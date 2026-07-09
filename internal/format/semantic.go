package format

import (
	"bytes"
	"fmt"

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

	if len(before.Children) != len(after.Children) {
		return fmt.Errorf("%s at %s changed child count from %d to %d",
			before.Kind, path, len(before.Children), len(after.Children))
	}

	for i := range before.Children {
		childPath := fmt.Sprintf("%s/%s[%d]", path, before.Kind, i)
		if err := compareSemanticNodes(before.Children[i], after.Children[i], childPath); err != nil {
			return err
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
