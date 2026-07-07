package main

import (
	"fmt"
	"strings"
)

func unifiedDiff(path string, before, after []byte) string {
	return unifiedDiffColored(path, before, after, cliColors{})
}

func unifiedDiffColored(path string, before, after []byte, colors cliColors) string {
	aLines := splitLines(string(before))
	bLines := splitLines(string(after))
	ops := diffLines(aLines, bLines)
	if !opsHaveChanges(ops) {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", colors.red("--- "+path))
	fmt.Fprintf(&b, "%s\n", colors.green("+++ "+path))

	const context = 3
	for i := 0; i < len(ops); {
		if ops[i].kind == opEqual {
			i++
			continue
		}
		start := i
		for i < len(ops) && ops[i].kind != opEqual {
			i++
		}
		end := i

		hunkStart := max(start-context, 0)
		hunkEnd := min(end+context, len(ops))

		writeHunk(&b, ops[hunkStart:hunkEnd], colors)
	}
	return b.String()
}

func opsHaveChanges(ops []lineOp) bool {
	for _, op := range ops {
		if op.kind != opEqual {
			return true
		}
	}
	return false
}

type opKind int

const (
	opEqual opKind = iota
	opDelete
	opInsert
)

type lineOp struct {
	kind opKind
	text string
}

func writeHunk(b *strings.Builder, ops []lineOp, colors cliColors) {
	oldCount, newCount := 0, 0
	for _, op := range ops {
		switch op.kind {
		case opEqual:
			oldCount++
			newCount++
		case opDelete:
			oldCount++
		case opInsert:
			newCount++
		}
	}
	fmt.Fprintf(b, "%s\n", colors.cyan(fmt.Sprintf("@@ -%d,%d +%d,%d @@", 1, oldCount, 1, newCount)))
	for _, op := range ops {
		switch op.kind {
		case opEqual:
			fmt.Fprintf(b, " %s\n", op.text)
		case opDelete:
			fmt.Fprintf(b, "%s\n", colors.red("-"+op.text))
		case opInsert:
			fmt.Fprintf(b, "%s\n", colors.green("+"+op.text))
		}
	}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func diffLines(a, b []string) []lineOp {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				lcs[i][j] = lcs[i+1][j+1] + 1
			case lcs[i+1][j] >= lcs[i][j+1]:
				lcs[i][j] = lcs[i+1][j]
			default:
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	var ops []lineOp
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			ops = append(ops, lineOp{opEqual, a[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, lineOp{opDelete, a[i]})
			i++
		default:
			ops = append(ops, lineOp{opInsert, b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, lineOp{opDelete, a[i]})
	}
	for ; j < m; j++ {
		ops = append(ops, lineOp{opInsert, b[j]})
	}
	return ops
}
