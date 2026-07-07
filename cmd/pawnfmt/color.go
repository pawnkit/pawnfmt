package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type cliColors struct {
	enabled bool
}

func colorsFor(mode string, w io.Writer) cliColors {
	switch mode {
	case "always":
		return cliColors{enabled: true}
	case "never":
		return cliColors{}
	}
	if os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return cliColors{}
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return cliColors{enabled: true}
	}
	f, ok := w.(*os.File)
	if !ok {
		return cliColors{}
	}
	info, err := f.Stat()
	return cliColors{enabled: err == nil && info.Mode()&os.ModeCharDevice != 0}
}

func (c cliColors) paint(code, s string) string {
	if !c.enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (c cliColors) red(s string) string     { return c.paint("31", s) }
func (c cliColors) green(s string) string   { return c.paint("32", s) }
func (c cliColors) yellow(s string) string  { return c.paint("33", s) }
func (c cliColors) cyan(s string) string    { return c.paint("36", s) }
func (c cliColors) magenta(s string) string { return c.paint("35", s) }
func (c cliColors) bold(s string) string    { return c.paint("1", s) }

func writeErrorf(w io.Writer, colors cliColors, format string, args ...any) {
	_, _ = fmt.Fprintf(w, "%s %s\n", colors.red("pawnfmt:"), fmt.Sprintf(format, args...))
}
