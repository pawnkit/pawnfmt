package printer

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/pawnkit/pawnfmt/internal/doc"
)

// Options configures how a doc.Doc tree is rendered to text.
type Options struct {
	LineWidth              int
	IndentWidth            int
	IndentStyle            string
	Newline                string
	InsertFinalNewline     bool
	TrimTrailingWhitespace bool
}

type mode uint8

const (
	modeBreak mode = iota
	modeFlat
)

type command struct {
	indent int
	mode   mode
	doc    doc.Doc
}

// Print renders root to text using options.
func Print(root doc.Doc, options Options) string {
	if root == nil {
		return finalize("", options)
	}

	options = options.withDefaults()

	return finalize(render(root, options), options)
}

func (options Options) withDefaults() Options {
	if options.LineWidth <= 0 {
		options.LineWidth = 100
	}

	if options.IndentWidth <= 0 {
		options.IndentWidth = 4
	}

	if options.Newline == "" {
		options.Newline = "\n"
	}

	return options
}

// renderer holds the mutable state threaded through render's main loop: the
// output buffer, the pending command stack, deferred line-suffix comments,
// and the current output column.
type renderer struct {
	builder      strings.Builder
	commands     []command
	lineSuffixes []command
	column       int
	options      Options
}

// deferLine re-queues current behind any pending line-suffix comments so
// they get flushed before the line actually breaks. Reports whether it did.
func (r *renderer) deferLine(current command) bool {
	if len(r.lineSuffixes) == 0 {
		return false
	}

	r.commands = append(r.commands, current)
	for _, suffix := range slices.Backward(r.lineSuffixes) {
		r.commands = append(r.commands, suffix)
	}

	r.lineSuffixes = nil

	return true
}

// breakLine writes a newline plus indent for the current command unless a
// pending line-suffix defers it.
func (r *renderer) breakLine(current command) {
	if r.deferLine(current) {
		return
	}

	writeIndent(&r.builder, current.indent, r.options)
	r.column = current.indent * r.options.IndentWidth
}

func (r *renderer) step(current command) {
	switch node := current.doc.(type) {
	case nil:
	case doc.TextDoc:
		r.builder.WriteString(node.Value)
		r.column += utf8.RuneCountInString(node.Value)
	case doc.LineDoc:
		if current.mode == modeFlat {
			r.builder.WriteByte(' ')
			r.column++
		} else {
			r.breakLine(current)
		}
	case doc.SoftLineDoc:
		if current.mode == modeBreak {
			r.breakLine(current)
		}
	case doc.HardLineDoc:
		r.breakLine(current)
	case doc.BreakParentDoc:
	case doc.LineSuffixDoc:
		r.lineSuffixes = append(r.lineSuffixes, command{indent: current.indent, mode: current.mode, doc: node.Contents})
	case doc.GroupDoc:
		r.stepGroup(current, node)
	case doc.FillDoc:
		r.commands = append(r.commands, fillCommands(node.Parts, current.indent, r.options.LineWidth-r.column, r.options)...)
	default:
		r.commands = pushRenderChildren(r.commands, current, node)
	}
}

func (r *renderer) stepGroup(current command, node doc.GroupDoc) {
	next := command{indent: current.indent, mode: modeFlat, doc: node.Contents}
	if !hasForcedBreak(node.Contents) && fits(r.options.LineWidth-r.column, next, r.commands, false, r.options) {
		r.commands = append(r.commands, command{indent: current.indent, mode: modeFlat, doc: node.Contents})
	} else {
		r.commands = append(r.commands, command{indent: current.indent, mode: modeBreak, doc: node.Contents})
	}
}

// pushRenderChildren pushes the child commands of a container Doc (Concat,
// Indent, ResetIndent, Outdent, IfBreak) onto commands for render's traversal.
func pushRenderChildren(commands []command, current command, node doc.Doc) []command {
	switch node := node.(type) {
	case doc.ConcatDoc:
		for _, part := range slices.Backward(node.Parts) {
			commands = append(commands, command{indent: current.indent, mode: current.mode, doc: part})
		}
	case doc.IndentDoc:
		commands = append(commands, command{indent: current.indent + 1, mode: current.mode, doc: node.Contents})
	case doc.ResetIndentDoc:
		commands = append(commands, command{indent: 0, mode: current.mode, doc: node.Contents})
	case doc.OutdentDoc:
		indent := max(current.indent-1, 0)
		commands = append(commands, command{indent: indent, mode: current.mode, doc: node.Contents})
	case doc.IfBreakDoc:
		if current.mode == modeFlat {
			if node.Flat != nil {
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: node.Flat})
			}
		} else if node.Broken != nil {
			commands = append(commands, command{indent: current.indent, mode: current.mode, doc: node.Broken})
		}
	}

	return commands
}

func render(root doc.Doc, options Options) string {
	r := &renderer{
		commands: []command{{indent: 0, mode: modeBreak, doc: root}},
		options:  options,
	}

	for len(r.commands) > 0 || len(r.lineSuffixes) > 0 {
		if len(r.commands) == 0 {
			for _, suffix := range slices.Backward(r.lineSuffixes) {
				r.commands = append(r.commands, suffix)
			}

			r.lineSuffixes = nil

			continue
		}

		current := r.commands[len(r.commands)-1]
		r.commands = r.commands[:len(r.commands)-1]

		r.step(current)
	}

	return r.builder.String()
}
