package check

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	parser "github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

func CollectPawnFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		switch strings.ToLower(filepath.Ext(path)) {
		case ".pwn", ".inc":
			files = append(files, path)
		}

		return nil
	})

	sort.Strings(files)

	return files, err
}

type CorpusStatus string

const (
	CorpusStatusFull     CorpusStatus = "full"
	CorpusStatusSafe     CorpusStatus = "safe"
	CorpusStatusPreserve CorpusStatus = "preserve"
	CorpusStatusFail     CorpusStatus = "fail"
)

// CorpusResult is the outcome of analyzing one real-world source file.
type CorpusResult struct {
	Path       string
	Status     CorpusStatus
	RawPercent float64
	Idempotent bool
	Detail     string
}

func AnalyzeCorpusFile(path string, cfg config.Config) (r CorpusResult) {
	r.Path = path

	defer func() {
		if rec := recover(); rec != nil {
			r.Status = CorpusStatusFail
			r.Detail = fmt.Sprintf("PANIC: %v", rec)
		}
	}()

	source, err := os.ReadFile(path)
	if err != nil {
		r.Status = CorpusStatusFail
		r.Detail = err.Error()

		return r
	}

	parsed := parser.Parse(source)
	if parsed.HasParseErrors() {
		r.Status = CorpusStatusFail
		r.Detail = "parser reported Broken (internal confusion, not just a raw region)"

		return r
	}

	total, raw := corpusRawCoverage(parsed.Root)
	if total > 0 {
		r.RawPercent = 100 * float64(raw) / float64(total)
	}

	formatted, ferr := formatter.FormatSource(source, cfg)
	if ferr != nil {
		r.Status = CorpusStatusFail
		r.Detail = "format: " + ferr.Error()

		return r
	}

	idempotent, ferr2 := Idempotent(formatted, func(b []byte) ([]byte, error) {
		return formatter.FormatSource(b, cfg)
	})
	if ferr2 != nil {
		r.Status = CorpusStatusFail
		r.Detail = "second-pass format: " + ferr2.Error()

		return r
	}

	r.Idempotent = idempotent
	if !r.Idempotent {
		r.Status = CorpusStatusFail
		r.Detail = "not idempotent (format(format(x)) != format(x))"

		return r
	}

	switch {
	case r.RawPercent < 1:
		r.Status = CorpusStatusFull
	case r.RawPercent < 50:
		r.Status = CorpusStatusSafe
	default:
		r.Status = CorpusStatusPreserve
	}

	return r
}

func corpusRawCoverage(n *parser.Node) (total, raw int) {
	if n == nil {
		return 0, 0
	}

	span := max(n.End-n.Start, 0)
	if n.Kind != parser.KindSourceFile && n.Kind != parser.KindConditionalRegion && n.Kind != parser.KindConditionalBranch {
		if n.Kind == parser.KindRaw || n.HasError {
			return span, span
		}
	}

	if len(n.Children) == 0 {
		return span, 0
	}

	covered := 0

	for _, c := range n.Children {
		ct, cr := corpusRawCoverage(c)
		total += ct
		raw += cr
		covered += c.End - c.Start
	}

	total += span - covered

	return total, raw
}
