package format

import (
	"bytes"
	"fmt"

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
		if want[i].kind != got[i].kind || !bytes.Equal(want[i].text, got[i].text) {
			return fmt.Errorf("semantic token %d changed from %s %q to %s %q", i+1,
				want[i].kind, want[i].text, got[i].kind, got[i].text)
		}
	}
	if len(want) != len(got) {
		return fmt.Errorf("semantic token count changed from %d to %d", len(want), len(got))
	}
	return nil
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
	switch kind {
	case token.EOF, token.LBrace, token.RBrace, token.Comma, token.Semicolon:
		return true
	default:
		return false
	}
}
