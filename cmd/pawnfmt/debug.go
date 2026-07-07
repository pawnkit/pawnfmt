package main

import (
	"fmt"
	"io"

	"github.com/pawnkit/pawn-parser/lexer"
)

func debugTokens(source []byte, w io.Writer) {
	toks := lexer.Tokenize(source)
	for _, t := range toks {
		_, _ = fmt.Fprintf(w, "%-16s [%d:%d-%d:%d] %q\n",
			t.Kind.String(), t.Start.Line, t.Start.Col, t.End.Line, t.End.Col, t.Text(source))
	}
}
