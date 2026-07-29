package format_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/pawnkit/pawnfmt/internal/config"
)

var compilerDiagnosticCode = regexp.MustCompile(`(?m)(?:error|warning) 0*([0-9]+):`)

func TestCompilerEquivalence(t *testing.T) {
	t.Parallel()

	compiler := os.Getenv("PAWNFMT_PAWNCC")
	corpus := os.Getenv("PAWN_CORPUS_DIR")

	if compiler == "" || corpus == "" {
		t.Skip("set PAWNFMT_PAWNCC and PAWN_CORPUS_DIR to run compiler equivalence")
	}

	fixture := filepath.Join(corpus, "format", "pairs", "compiler_equivalence")
	input := readFile(t, filepath.Join(fixture, "input.pwn"))
	expected := readFile(t, filepath.Join(fixture, "expected.pwn"))
	formatted := mustFormat(t, input, config.Default())

	if !bytes.Equal(formatted, expected) {
		t.Fatalf("formatter output does not match the shared corpus fixture\nexpected:\n%s\nactual:\n%s", expected, formatted)
	}

	before := compilePawn(t, compiler, input)
	after := compilePawn(t, compiler, formatted)

	if before.exitCode != after.exitCode || !slices.Equal(before.diagnosticCodes, after.diagnosticCodes) {
		t.Fatalf(
			"compiler result changed after formatting\nbefore: exit %d, diagnostics %v\n%s\nafter: exit %d, diagnostics %v\n%s",
			before.exitCode, before.diagnosticCodes, before.output,
			after.exitCode, after.diagnosticCodes, after.output,
		)
	}

	if before.exitCode == 0 && !bytes.Equal(before.artifact, after.artifact) {
		t.Fatal("compiled AMX changed after formatting")
	}
}

type compilerResult struct {
	exitCode        int
	diagnosticCodes []string
	output          []byte
	artifact        []byte
}

func compilePawn(t *testing.T, compiler string, source []byte) compilerResult {
	t.Helper()
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "fixture.pwn")
	artifactPath := filepath.Join(dir, "fixture.amx")

	if err := os.WriteFile(sourcePath, source, 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, compiler, "fixture.pwn", "-ofixture.amx")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if ctx.Err() != nil {
		t.Fatalf("pawncc timed out: %v", ctx.Err())
	}

	exitCode := 0

	if err != nil {
		var exitErr *exec.ExitError

		if !errors.As(err, &exitErr) {
			t.Fatalf("run pawncc: %v", err)
		}

		exitCode = exitErr.ExitCode()
	}

	codes := compilerDiagnosticCode.FindAllStringSubmatch(string(output), -1)
	diagnostics := make([]string, 0, len(codes))

	for _, code := range codes {
		diagnostics = append(diagnostics, code[1])
	}

	artifact, artifactErr := os.ReadFile(artifactPath)

	if exitCode == 0 && artifactErr != nil {
		t.Fatalf("read compiler artifact: %v", artifactErr)
	}

	return compilerResult{
		exitCode: exitCode, diagnosticCodes: diagnostics,
		output: output, artifact: artifact,
	}
}
