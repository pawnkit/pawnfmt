package config

func ResolveNewline(style NewlineStyle, detected string) string {
	switch style {
	case NewlineStyleCRLF:
		return "\r\n"
	case NewlineStyleLF:
		return "\n"
	default:
		if detected == "\r\n" {
			return detected
		}

		return "\n"
	}
}
