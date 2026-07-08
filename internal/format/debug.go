package format

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
	"github.com/pawnkit/pawnfmt/internal/trivia"
)

func DebugDocTree(source []byte, cfg config.Config) (string, error) {
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return "", err
	}
	parsed := parser.Parse(source)
	if parsed.HasParseErrors() {
		return "", errors.New("source does not parse cleanly")
	}
	index := trivia.Scan(source)
	st := newState(parsed, cfg, index)
	root := st.formatNode(parsed.Root)

	var b strings.Builder
	dumpDoc(&b, root, 0)
	return b.String(), nil
}

func dumpDoc(b *strings.Builder, d doc.Doc, depth int) {
	indent := strings.Repeat("  ", depth)
	switch n := d.(type) {
	case doc.TextDoc:
		fmt.Fprintf(b, "%sText(%q)\n", indent, n.Value)
	case doc.LineDoc:
		fmt.Fprintf(b, "%sLine\n", indent)
	case doc.SoftLineDoc:
		fmt.Fprintf(b, "%sSoftLine\n", indent)
	case doc.HardLineDoc:
		fmt.Fprintf(b, "%sHardLine\n", indent)
	case doc.BreakParentDoc:
		fmt.Fprintf(b, "%sBreakParent\n", indent)
	case doc.LineSuffixDoc:
		fmt.Fprintf(b, "%sLineSuffix\n", indent)
		dumpDoc(b, n.Contents, depth+1)
	case doc.ConcatDoc:
		fmt.Fprintf(b, "%sConcat\n", indent)
		for _, part := range n.Parts {
			dumpDoc(b, part, depth+1)
		}
	case doc.IndentDoc:
		fmt.Fprintf(b, "%sIndent\n", indent)
		dumpDoc(b, n.Contents, depth+1)
	case doc.ResetIndentDoc:
		fmt.Fprintf(b, "%sResetIndent\n", indent)
		dumpDoc(b, n.Contents, depth+1)
	case doc.OutdentDoc:
		fmt.Fprintf(b, "%sOutdent\n", indent)
		dumpDoc(b, n.Contents, depth+1)
	case doc.GroupDoc:
		fmt.Fprintf(b, "%sGroup\n", indent)
		dumpDoc(b, n.Contents, depth+1)
	case doc.IfBreakDoc:
		fmt.Fprintf(b, "%sIfBreak\n", indent)
		if n.Broken != nil {
			fmt.Fprintf(b, "%s  broken:\n", indent)
			dumpDoc(b, n.Broken, depth+2)
		}
		if n.Flat != nil {
			fmt.Fprintf(b, "%s  flat:\n", indent)
			dumpDoc(b, n.Flat, depth+2)
		}
	case doc.FillDoc:
		fmt.Fprintf(b, "%sFill\n", indent)
		for _, part := range n.Parts {
			dumpDoc(b, part, depth+1)
		}
	default:
		fmt.Fprintf(b, "%s<unknown doc node>\n", indent)
	}
}

func DebugCST(source []byte) string {
	parsed := parser.Parse(source)
	var b strings.Builder
	if parsed.Broken {
		fmt.Fprintf(&b, "parse reports broken syntax\n")
	}
	dumpNode(&b, parsed.Root, 0)
	return b.String()
}

func dumpNode(b *strings.Builder, n *parser.Node, depth int) {
	if n == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	marker := ""
	if n.HasError {
		marker = " [raw/error]"
	}
	fmt.Fprintf(b, "%s%s [%d:%d]%s\n", indent, n.Kind, n.Start, n.End, marker)
	for _, c := range n.Children {
		dumpNode(b, c, depth+1)
	}
}
