package config

// ResolveNewline picks the line ending for style, using detected when style is auto.
func ResolveNewline(style NewlineStyle, detected string) string {
	switch style {
	case NewlineStyleCRLF:
		return "\r\n"
	case NewlineStyleLF:
		return "\n"
	case NewlineStyleAuto:
	default:
	}

	if detected == "\r\n" {
		return detected
	}

	return "\n"
}
