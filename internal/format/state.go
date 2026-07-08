package format

import (
	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
	"github.com/pawnkit/pawnfmt/internal/printer"
	"github.com/pawnkit/pawnfmt/internal/trivia"
)

type state struct {
	file   *parser.File
	source []byte
	config config.Config
	trivia trivia.Index

	topLevelContext bool

	inMacroValue bool

	hint nextItemHint

	renderedComments map[int]bool

	commentPadWidths map[int]int

	paramQualifiers map[*parser.Node]*parser.Node
}

type nextItemHint struct {
	suppressIfAlternative bool
	alignDeclarationWidth int
	alignMacroValueWidth  int
}

func (s *state) takeSuppressIfAlternative() bool {
	v := s.hint.suppressIfAlternative
	s.hint.suppressIfAlternative = false
	return v
}

func (s *state) takeAlignDeclarationWidth() int {
	v := s.hint.alignDeclarationWidth
	s.hint.alignDeclarationWidth = 0
	return v
}

func (s *state) takeAlignMacroValueWidth() int {
	v := s.hint.alignMacroValueWidth
	s.hint.alignMacroValueWidth = 0
	return v
}

func newState(file *parser.File, cfg config.Config, index trivia.Index) *state {
	return &state{
		file:             file,
		source:           file.Source,
		config:           cfg,
		trivia:           index,
		renderedComments: make(map[int]bool),
		commentPadWidths: make(map[int]int),
	}
}

func (s *state) continuationIndentWidth() int {
	if s.config.ContinuationIndentWidth > 0 {
		return s.config.ContinuationIndentWidth
	}
	return s.config.IndentWidth
}

func (s *state) printerOptions() printer.Options {
	newline := config.ResolveNewline(s.config.NewlineStyle, s.trivia.DetectedNewline)
	return printer.Options{
		LineWidth:              s.config.LineWidth,
		IndentWidth:            s.config.IndentWidth,
		IndentStyle:            string(s.config.IndentStyle),
		Newline:                newline,
		InsertFinalNewline:     s.config.InsertFinalNewline,
		TrimTrailingWhitespace: s.config.TrimTrailingWhitespace,
	}
}

func (s *state) renderFlat(node doc.Doc) string {
	opts := s.printerOptions()
	opts.LineWidth = 1 << 30
	opts.InsertFinalNewline = false
	opts.TrimTrailingWhitespace = false
	return printer.Print(node, opts)
}

func (s *state) renderDoc(node doc.Doc) string {
	opts := s.printerOptions()
	opts.InsertFinalNewline = false
	opts.TrimTrailingWhitespace = false
	return printer.Print(node, opts)
}

func (s *state) measureFlat(n *parser.Node) string {
	tmp := *s
	tmp.renderedComments = make(map[int]bool, len(s.renderedComments))
	tmp.hint = nextItemHint{}
	return tmp.renderFlat(tmp.formatNode(n))
}
