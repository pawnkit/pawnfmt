package format_test

import (
	"path/filepath"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/check"
	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestCorpusReport(t *testing.T) {
	root := filepath.Join(testdataDir(), "real-world")

	files, err := check.CollectPawnFiles(root)
	if err != nil {
		t.Fatalf("collect real-world corpus files under %s: %v", root, err)
	}

	if len(files) == 0 {
		t.Fatalf("no .pwn/.inc files found under %s -- did testdata/real-world/fetch.sh run?", root)
	}

	cfg := config.Default()

	var failed int

	for _, f := range files {
		r := check.AnalyzeCorpusFile(f, cfg)
		if r.Status == check.CorpusStatusFail {
			failed++

			t.Errorf("%s: %s (%.1f%% raw, idempotent=%v)", r.Path, r.Detail, r.RawPercent, r.Idempotent)
		}
	}

	if failed > 0 {
		t.Fatalf("%d/%d real-world corpus files failed (see above)", failed, len(files))
	}
}
