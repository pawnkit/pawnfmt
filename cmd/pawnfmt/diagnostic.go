package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

type cliDiagnostic struct {
	Severity string `json:"severity"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Offset   *int   `json:"offset,omitempty"`
}

func writeDiagnostic(w io.Writer, colors cliColors, errorFormat, category, path string, err error) {
	diagnostic := diagnosticFromError(category, path, err)
	switch errorFormat {
	case "json":
		_ = json.NewEncoder(w).Encode(diagnostic)
	case "github":
		writeGitHubDiagnostic(w, diagnostic)
	default:
		prefix := ""
		if path != "" {
			prefix = path + ": "
		}
		writeErrorf(w, colors, "%s%v", prefix, err)
	}
}

func diagnosticFromError(category, path string, err error) cliDiagnostic {
	diagnostic := cliDiagnostic{
		Severity: "error", Category: category, Message: err.Error(), Path: path,
	}

	var parseErr *formatter.ParseError
	if errors.As(err, &parseErr) {
		diagnostic.Category = "parse"
		diagnostic.Message = parseErr.Summary()
		diagnostic.Line = parseErr.Line
		diagnostic.Column = parseErr.Column
		offset := parseErr.Offset
		diagnostic.Offset = &offset
	}

	return diagnostic
}

func writeGitHubDiagnostic(w io.Writer, diagnostic cliDiagnostic) {
	var properties []string
	if diagnostic.Path != "" {
		properties = append(properties, "file="+escapeGitHubProperty(diagnostic.Path))
	}
	if diagnostic.Line > 0 {
		properties = append(properties, fmt.Sprintf("line=%d", diagnostic.Line))
	}
	if diagnostic.Column > 0 {
		properties = append(properties, fmt.Sprintf("col=%d", diagnostic.Column))
	}
	properties = append(properties, "title=pawnfmt "+escapeGitHubProperty(diagnostic.Category))
	_, _ = fmt.Fprintf(w, "::error %s::%s\n", strings.Join(properties, ","), escapeGitHubMessage(diagnostic.Message))
}

func escapeGitHubMessage(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	return strings.ReplaceAll(value, "\n", "%0A")
}

func escapeGitHubProperty(value string) string {
	value = escapeGitHubMessage(value)
	value = strings.ReplaceAll(value, ":", "%3A")
	return strings.ReplaceAll(value, ",", "%2C")
}

func writeOptionErrorf(opts *options, w io.Writer, colors cliColors, category, path, format string, args ...any) {
	errorFormat := "human"
	if opts != nil && opts.ErrorFormat != "" {
		errorFormat = opts.ErrorFormat
	}
	err := fmt.Errorf(format, args...)
	if format == "%v" && len(args) == 1 {
		if original, ok := args[0].(error); ok {
			err = original
		}
	}
	writeDiagnostic(w, colors, errorFormat, category, path, err)
}
