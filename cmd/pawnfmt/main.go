package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
)

const (
	exitOK            = 0
	exitCheckChanges  = 1
	exitFormatError   = 2
	exitConfigError   = 3
	exitInternalError = 4
)

type options struct {
	Write          bool             `short:"w" xor:"mode" help:"write formatted output back to each file"`
	Check          bool             `xor:"mode" help:"exit with status 1 if any file would change, without writing"`
	Diff           bool             `xor:"mode" help:"print a unified diff of formatting changes"`
	Stdin          bool             `xor:"mode" help:"read source from stdin, write formatted output to stdout"`
	Color          string           `default:"auto" enum:"auto,always,never" help:"when to use colour in output"`
	StdinFilename  string           `help:"filename to use for config discovery when reading from stdin"`
	Config         string           `help:"path to a pawnfmt config file (.toml/.yaml/.json)"`
	NoConfig       bool             `help:"ignore any discovered config file and use built-in defaults"`
	NoGitignore    bool             `help:"do not respect .gitignore/.pawnfmtignore files when walking directories"`
	PrintConfig    bool             `help:"print the resolved configuration and exit"`
	InitConfig     bool             `help:"write a fully-commented pawnfmt.toml with default values and exit (pass a path as the first argument to write elsewhere)"`
	DebugTokens    bool             `help:"print the lexer token stream for the input instead of formatting"`
	DebugCST       bool             `help:"print the parsed CST for the input instead of formatting"`
	DebugFormatDoc bool             `help:"print the formatter's intermediate document tree instead of formatted output"`
	Version        kong.VersionFlag `help:"print version information and exit"`
	Paths          []string         `arg:"" optional:"" name:"path" help:"files or directories to format"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) (code int) {
	colors := colorsFor("auto", stderr)

	defer func() {
		if r := recover(); r != nil {
			writeErrorf(stderr, colors, "internal error: %v", r)

			code = exitInternalError
		}
	}()

	opts, exitCode, done := parseCLI(args, stdout, stderr)
	if opts != nil {
		colors = colorsFor(opts.Color, stderr)
	}

	if done {
		return exitCode
	}

	return dispatch(opts, stdin, stdout, stderr)
}

type kongEagerExit struct{ code int }

func parseCLI(args []string, stdout, stderr io.Writer) (opts *options, code int, done bool) {
	opts = &options{}

	parser, err := kong.New(opts,
		kong.Name("pawnfmt"),
		kong.Description("A formatter for Pawn (SA-MP/open.mp) source files."),
		kong.Writers(stdout, stderr),
		kong.Exit(func(c int) { panic(kongEagerExit{code: c}) }),
		kong.Vars{"version": versionString()},
	)
	if err != nil {
		panic(fmt.Errorf("cli setup: %w", err))
	}

	defer func() {
		r := recover()
		if r == nil {
			return
		}

		eager, ok := r.(kongEagerExit)
		if !ok {
			panic(r)
		}

		code = eager.code
		done = true
	}()

	if _, err := parser.Parse(args); err != nil {
		writeErrorf(stderr, colorsFor(opts.Color, stderr), "%v", err)
		return opts, exitConfigError, true
	}

	return opts, exitOK, false
}
