package trivia

import (
	"bytes"
	"regexp"
	"strings"
)

var (
	pawnfmtOffPattern = regexp.MustCompile(`(?i)\bpawnfmt\s+off\b`)
	pawnfmtOnPattern  = regexp.MustCompile(`(?i)\bpawnfmt\s+on\b`)
)

type Line struct {
	Number        int
	StartByte     int
	EndByte       int
	Text          string
	Trimmed       string
	IsBlank       bool
	IsDirective   bool
	IsCommentOnly bool
	TurnsOff      bool
	TurnsOn       bool
}

type Region struct {
	StartLine int
	EndLine   int
	StartByte int
	EndByte   int
}

type Index struct {
	Lines           []Line
	Disabled        []Region
	DetectedNewline string
}

func Scan(source []byte) Index {
	detected := detectNewline(source)

	lines := make([]Line, 0, bytes.Count(source, []byte("\n"))+1)
	start := 0
	lineNumber := 1
	for start <= len(source) {
		end := indexOfLineEnd(source, start)
		lineBytes := source[start:end]
		text := string(lineBytes)
		trimmed := strings.TrimSpace(text)
		commentOnly := isCommentOnly(trimmed)
		line := Line{
			Number:        lineNumber,
			StartByte:     start,
			EndByte:       end,
			Text:          text,
			Trimmed:       trimmed,
			IsBlank:       trimmed == "",
			IsDirective:   strings.HasPrefix(strings.TrimLeft(text, " \t"), "#"),
			IsCommentOnly: commentOnly,
			TurnsOff:      commentOnly && pawnfmtOffPattern.MatchString(text),
			TurnsOn:       commentOnly && pawnfmtOnPattern.MatchString(text),
		}
		lines = append(lines, line)
		newlineLen := lineEndingLength(source, end)
		if newlineLen == 0 {
			break
		}
		start = end + newlineLen
		lineNumber++
	}

	return Index{Lines: lines, Disabled: buildDisabledRegions(lines, source), DetectedNewline: detected}
}

func lineEndingLength(source []byte, end int) int {
	if end >= len(source) {
		return 0
	}
	if source[end] == '\r' {
		return 2
	}
	return 1
}

func (index Index) OverlapsDisabled(startByte, endByte uint32) bool {
	for _, region := range index.Disabled {
		if int(endByte) <= region.StartByte {
			continue
		}
		if int(startByte) >= region.EndByte {
			continue
		}
		return true
	}
	return false
}

func detectNewline(source []byte) string {
	if bytes.Contains(source, []byte("\r\n")) {
		return "\r\n"
	}
	return "\n"
}

func buildDisabledRegions(lines []Line, source []byte) []Region {
	sourceLen := len(source)
	regions := make([]Region, 0)
	active := false
	startLine := 0
	startByte := 0
	for _, line := range lines {
		if line.TurnsOff && !active {
			active = true
			startLine = line.Number
			startByte = line.StartByte
		}
		if line.TurnsOn && active {
			regions = append(regions, Region{
				StartLine: startLine,
				EndLine:   line.Number,
				StartByte: startByte,
				EndByte:   min(sourceLen, line.EndByte+lineEndingLength(source, line.EndByte)),
			})
			active = false
		}
	}
	if active {
		regions = append(regions, Region{
			StartLine: startLine,
			EndLine:   len(lines),
			StartByte: startByte,
			EndByte:   sourceLen,
		})
	}
	return regions
}

func isCommentOnly(trimmed string) bool {
	return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
}

func indexOfLineEnd(source []byte, start int) int {
	for index := start; index < len(source); index++ {
		if source[index] == '\n' {
			if index > start && source[index-1] == '\r' {
				return index - 1
			}
			return index
		}
	}
	return len(source)
}
