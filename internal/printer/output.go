package printer

import "strings"

func finalize(value string, options Options) string {
	if options.TrimTrailingWhitespace {
		value = trimTrailingWhitespace(value, options.Newline)
	}

	if options.InsertFinalNewline {
		if !strings.HasSuffix(value, options.Newline) {
			value += options.Newline
		}
	} else {
		value = strings.TrimSuffix(value, "\n")
		value = strings.TrimSuffix(value, "\r")
	}

	return value
}

func trimTrailingWhitespace(value, newline string) string {
	parts := strings.Split(value, newline)
	for index, part := range parts {
		parts[index] = strings.TrimRight(part, " \t")
	}

	return strings.Join(parts, newline)
}

func writeIndent(builder *strings.Builder, indent int, options Options) {
	builder.WriteString(options.Newline)

	if indent <= 0 {
		return
	}

	if options.IndentStyle == "tab" {
		for range indent {
			builder.WriteByte('\t')
		}

		return
	}

	for range indent * options.IndentWidth {
		builder.WriteByte(' ')
	}
}
