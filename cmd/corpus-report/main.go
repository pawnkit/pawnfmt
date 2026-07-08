package main

import (
	"fmt"
	"os"

	"github.com/pawnkit/pawnfmt/internal/check"
	"github.com/pawnkit/pawnfmt/internal/config"
)

func main() {
	root := "testdata/real-world"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	files, err := check.CollectPawnFiles(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "no .pwn/.inc files found under %s (did you run fetch.sh?)\n", root)
		os.Exit(1)
	}

	cfg := config.Default()

	results := make([]check.CorpusResult, len(files))
	for i, f := range files {
		results[i] = check.AnalyzeCorpusFile(f, cfg)
	}

	counts := map[check.CorpusStatus]int{}
	for _, r := range results {
		counts[r.Status]++
		fmt.Printf("%-9s %6.1f%% raw  idempotent=%-5v  %s\n", r.Status, r.RawPercent, r.Idempotent, r.Path)

		if r.Detail != "" {
			fmt.Printf("          %s\n", r.Detail)
		}
	}

	fmt.Println()
	fmt.Printf("total=%d full=%d safe=%d preserve=%d fail=%d\n",
		len(results), counts[check.CorpusStatusFull], counts[check.CorpusStatusSafe],
		counts[check.CorpusStatusPreserve], counts[check.CorpusStatusFail])

	if counts[check.CorpusStatusFail] > 0 {
		os.Exit(1)
	}
}
