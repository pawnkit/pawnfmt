package main

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var ignoreFileNames = []string{".gitignore", ".pawnfmtignore"}

type ignoreRule struct {
	re       *regexp.Regexp
	negate   bool
	dirOnly  bool
	anchored bool
}

func (r ignoreRule) matches(relPath string, isDir bool) bool {
	if r.dirOnly && !isDir {
		return false
	}

	if r.anchored {
		return r.re.MatchString(relPath)
	}

	return slices.ContainsFunc(strings.Split(relPath, "/"), r.re.MatchString)
}

func parseIgnoreRules(data []byte) []ignoreRule {
	var rules []ignoreRule

	for line := range strings.SplitSeq(string(data), "\n") {
		if rule, ok := compileIgnoreLine(line); ok {
			rules = append(rules, rule)
		}
	}

	return rules
}

func compileIgnoreLine(line string) (ignoreRule, bool) {
	line = strings.TrimRight(line, " \t\r")
	if line == "" || strings.HasPrefix(line, "#") {
		return ignoreRule{}, false
	}

	negate := strings.HasPrefix(line, "!")
	if negate {
		line = line[1:]
	}

	dirOnly := strings.HasSuffix(line, "/")
	line = strings.TrimSuffix(line, "/")
	anchored := strings.HasPrefix(line, "/")

	line = strings.TrimPrefix(line, "/")
	if line == "" {
		return ignoreRule{}, false
	}

	if strings.Contains(line, "/") {
		anchored = true
	}

	re, err := regexp.Compile(gitignoreGlobToRegexp(line))
	if err != nil {
		return ignoreRule{}, false
	}

	return ignoreRule{re: re, negate: negate, dirOnly: dirOnly, anchored: anchored}, true
}

func gitignoreGlobToRegexp(glob string) string {
	var b strings.Builder
	b.WriteString("^")

	runes := []rune(glob)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch c {
		case '*':
			if i+1 < len(runes) && runes[i+1] == '*' {
				switch {
				case i+2 < len(runes) && runes[i+2] == '/':
					b.WriteString("(?:.*/)?")

					i += 2 // outer loop's i++ consumes the trailing '/'
				default:
					b.WriteString(".*")

					i++
				}

				continue
			}

			b.WriteString("[^/]*")
		case '?':
			b.WriteString("[^/]")
		case '[':
			j := i + 1

			var cls strings.Builder
			cls.WriteByte('[')

			if j < len(runes) && runes[j] == '!' {
				cls.WriteByte('^')

				j++
			}

			for j < len(runes) && runes[j] != ']' {
				cls.WriteRune(runes[j])
				j++
			}

			cls.WriteByte(']')
			b.WriteString(cls.String())

			i = j
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}

	b.WriteString("$")

	return b.String()
}

type ignoreFrame struct {
	dir   string
	rules []ignoreRule
}

type ignoreStack struct {
	frames []ignoreFrame
}

func newIgnoreStack(startPath string) *ignoreStack {
	abs, err := filepath.Abs(startPath)
	if err != nil {
		return &ignoreStack{}
	}

	if info, statErr := os.Stat(abs); statErr == nil && !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	var ancestors []string

	dir := abs
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		ancestors = append(ancestors, parent)
		if _, err := os.Stat(filepath.Join(parent, ".git")); err == nil {
			break
		}

		dir = parent
	}

	s := &ignoreStack{}
	for i := len(ancestors) - 1; i >= 0; i-- {
		s.pushDir(ancestors[i])
	}

	return s
}

func (s *ignoreStack) pushDir(dir string) {
	var rules []ignoreRule

	for _, name := range ignoreFileNames {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}

		rules = append(rules, parseIgnoreRules(data)...)
	}

	if len(rules) > 0 {
		s.frames = append(s.frames, ignoreFrame{dir: dir, rules: rules})
	}
}

func (s *ignoreStack) sync(dir string) {
	for len(s.frames) > 0 {
		top := s.frames[len(s.frames)-1]

		rel, err := filepath.Rel(top.dir, dir)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			s.frames = s.frames[:len(s.frames)-1]
			continue
		}

		break
	}
}

func (s *ignoreStack) ignored(path string, isDir bool) bool {
	matched := false

	for _, f := range s.frames {
		rel, err := filepath.Rel(f.dir, path)
		if err != nil {
			continue
		}

		rel = filepath.ToSlash(rel)
		for _, r := range f.rules {
			if r.matches(rel, isDir) {
				matched = !r.negate
			}
		}
	}

	return matched
}
