package bench

import (
	"testing"

	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawn-parser/lexer"
	"github.com/pawnkit/pawnfmt/internal/config"
	formatter "github.com/pawnkit/pawnfmt/internal/format"
)

const (
	smallN  = 15
	mediumN = 150
	largeN  = 1500
)

var (
	smallSource  = GenerateSource(smallN)
	mediumSource = GenerateSource(mediumN)
	largeSource  = GenerateSource(largeN)

	macroHeavySource        = GenerateMacroHeavy(mediumN)
	preprocessorHeavySource = GeneratePreprocessorHeavy(mediumN)
)

func BenchmarkLexerSmall(b *testing.B)  { benchLexer(b, smallSource) }
func BenchmarkLexerMedium(b *testing.B) { benchLexer(b, mediumSource) }
func BenchmarkLexerLarge(b *testing.B)  { benchLexer(b, largeSource) }

func benchLexer(b *testing.B, source []byte) {
	b.Helper()
	b.SetBytes(int64(len(source)))
	b.ResetTimer()

	for range b.N {
		lexer.Tokenize(source)
	}
}

func BenchmarkParserSmall(b *testing.B)  { benchParser(b, smallSource) }
func BenchmarkParserMedium(b *testing.B) { benchParser(b, mediumSource) }
func BenchmarkParserLarge(b *testing.B)  { benchParser(b, largeSource) }

func benchParser(b *testing.B, source []byte) {
	b.Helper()
	b.SetBytes(int64(len(source)))
	b.ResetTimer()

	for range b.N {
		parser.Parse(source)
	}
}

func BenchmarkFormatSmall(b *testing.B)  { benchFormat(b, smallSource) }
func BenchmarkFormatMedium(b *testing.B) { benchFormat(b, mediumSource) }
func BenchmarkFormatLarge(b *testing.B)  { benchFormat(b, largeSource) }

func benchFormat(b *testing.B, source []byte) {
	b.Helper()

	cfg := config.Default()

	b.SetBytes(int64(len(source)))
	b.ResetTimer()

	for range b.N {
		if _, err := formatter.Source(source, cfg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFullPipelineSmall(b *testing.B)  { benchFormat(b, smallSource) }
func BenchmarkFullPipelineMedium(b *testing.B) { benchFormat(b, mediumSource) }
func BenchmarkFullPipelineLarge(b *testing.B)  { benchFormat(b, largeSource) }

func BenchmarkMacroHeavy(b *testing.B) {
	benchFormat(b, macroHeavySource)
}

func BenchmarkPreprocessorHeavy(b *testing.B) {
	benchFormat(b, preprocessorHeavySource)
}
