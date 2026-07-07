package printer

import (
	"strings"
	"unicode/utf8"

	"github.com/pawnkit/pawnfmt/internal/doc"
)

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

func render(root doc.Doc, options Options) string {
	var builder strings.Builder
	commands := []command{{indent: 0, mode: modeBreak, doc: root}}
	column := 0

	for len(commands) > 0 {
		current := commands[len(commands)-1]
		commands = commands[:len(commands)-1]

		switch node := current.doc.(type) {
		case nil:
		case doc.TextDoc:
			builder.WriteString(node.Value)
			column += utf8.RuneCountInString(node.Value)
		case doc.LineDoc:
			if current.mode == modeFlat {
				builder.WriteByte(' ')
				column++
			} else {
				writeIndent(&builder, current.indent, options)
				column = current.indent * options.IndentWidth
			}
		case doc.SoftLineDoc:
			if current.mode == modeBreak {
				writeIndent(&builder, current.indent, options)
				column = current.indent * options.IndentWidth
			}
		case doc.HardLineDoc:
			writeIndent(&builder, current.indent, options)
			column = current.indent * options.IndentWidth
		case doc.ConcatDoc:
			for index := len(node.Parts) - 1; index >= 0; index-- {
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: node.Parts[index]})
			}
		case doc.IndentDoc:
			commands = append(commands, command{indent: current.indent + 1, mode: current.mode, doc: node.Contents})
		case doc.ResetIndentDoc:
			commands = append(commands, command{indent: 0, mode: current.mode, doc: node.Contents})
		case doc.OutdentDoc:
			indent := max(current.indent-1, 0)
			commands = append(commands, command{indent: indent, mode: current.mode, doc: node.Contents})
		case doc.GroupDoc:
			next := command{indent: current.indent, mode: modeFlat, doc: node.Contents}
			if fits(options.LineWidth-column, next, commands, false, options) {
				commands = append(commands, command{indent: current.indent, mode: modeFlat, doc: node.Contents})
			} else {
				commands = append(commands, command{indent: current.indent, mode: modeBreak, doc: node.Contents})
			}
		case doc.IfBreakDoc:
			if current.mode == modeFlat {
				if node.Flat != nil {
					commands = append(commands, command{indent: current.indent, mode: current.mode, doc: node.Flat})
				}
			} else if node.Broken != nil {
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: node.Broken})
			}
		case doc.FillDoc:
			commands = append(commands, fillCommands(node.Parts, current.indent, options.LineWidth-column, options)...)
		}
	}

	return builder.String()
}
