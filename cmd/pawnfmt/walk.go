package main

import (
	"os"
	"path/filepath"
	"strings"
)

var defaultIgnoredDirs = map[string]bool{
	".git":         true,
	".svn":         true,
	".hg":          true,
	"node_modules": true,
	"vendor":       true,
	".cache":       true,
	"dist":         true,
	"build":        true,
}

func isFormattableExt(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pwn", ".inc":
		return true
	default:
		return false
	}
}

type fileCollector struct {
	root               string
	include, exclude   []string
	respectIgnoreFiles bool
	stack              *ignoreStack
	addFile            func(string)
}

func (c *fileCollector) visit(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		return c.visitDir(path, d)
	}
	return c.visitFile(path)
}

func (c *fileCollector) visitDir(path string, d os.DirEntry) error {
	if c.respectIgnoreFiles {
		c.stack.sync(path)
	}
	if path != c.root && !matchesAny(path, c.include) {
		if defaultIgnoredDirs[d.Name()] || (c.respectIgnoreFiles && c.stack.ignored(path, true)) {
			return filepath.SkipDir
		}
	}
	if c.respectIgnoreFiles {
		c.stack.pushDir(path)
	}
	return nil
}

func (c *fileCollector) visitFile(path string) error {
	if matchesAny(path, c.exclude) {
		return nil
	}
	if len(c.include) > 0 && !matchesAny(path, c.include) {
		return nil
	}
	if !isFormattableExt(path) {
		return nil
	}
	if c.respectIgnoreFiles && !matchesAny(path, c.include) {
		c.stack.sync(filepath.Dir(path))
		if c.stack.ignored(path, false) {
			return nil
		}
	}
	c.addFile(path)
	return nil
}

func collectFiles(paths, include, exclude []string, respectIgnoreFiles bool) ([]string, error) {
	var out []string
	seen := make(map[string]bool)

	addFile := func(path string) {
		abs, err := filepath.Abs(path)
		if err != nil {
			abs = path
		}
		if !seen[abs] {
			seen[abs] = true
			out = append(out, path)
		}
	}

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			addFile(p)
			continue
		}
		stack := &ignoreStack{}
		if respectIgnoreFiles {
			stack = newIgnoreStack(p)
		}
		c := &fileCollector{
			root:               p,
			include:            include,
			exclude:            exclude,
			respectIgnoreFiles: respectIgnoreFiles,
			stack:              stack,
			addFile:            addFile,
		}
		if err := filepath.WalkDir(p, c.visit); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func matchesAny(path string, patterns []string) bool {
	base := filepath.Base(path)
	for _, pat := range patterns {
		if ok, err := filepath.Match(pat, base); err == nil && ok {
			return true
		}
		if ok, err := filepath.Match(pat, path); err == nil && ok {
			return true
		}
	}
	return false
}
