package format

import (
	"fmt"

	"github.com/pawnkit/pawn-parser/lexer"

	"github.com/pawnkit/pawnfmt/internal/config"
)

// CursorResult contains formatted source and the cursor's adjusted byte offset.
type CursorResult struct {
	Source       []byte
	CursorOffset int
}

// FormatSourceWithCursor formats source and maps cursorOffset into the output.
// Include sorting is suppressed because moving whole directives is not a local
// cursor-preserving transformation.
func (formatter *Formatter) FormatSourceWithCursor(source []byte, cursorOffset int) (CursorResult, error) {
	if cursorOffset < 0 || cursorOffset > len(source) {
		return CursorResult{}, fmt.Errorf("invalid cursor offset %d for source of %d bytes", cursorOffset, len(source))
	}

	cursorFormatter := *formatter
	cursorFormatter.config.SortIncludes = false

	formatted, err := cursorFormatter.FormatSource(source)
	if err != nil {
		return CursorResult{}, err
	}

	return CursorResult{
		Source:       formatted,
		CursorOffset: mapCursorOffset(source, formatted, cursorOffset),
	}, nil
}

type positionedSemanticToken struct {
	semanticToken
	start int
	end   int
}

func mapCursorOffset(before, after []byte, cursor int) int {
	if cursor == 0 {
		return 0
	}

	if cursor == len(before) {
		return len(after)
	}

	want := positionedSemanticTokens(before)

	got := positionedSemanticTokens(after)
	if len(want) != len(got) || len(want) == 0 {
		return min(cursor, len(after))
	}

	for i, tok := range want {
		if tok.start <= cursor && cursor <= tok.end {
			relative := min(cursor-tok.start, got[i].end-got[i].start)
			return got[i].start + relative
		}

		if cursor < tok.start {
			if i == 0 {
				return max(got[0].start-(tok.start-cursor), 0)
			}

			previous := want[i-1]
			leftDistance := cursor - previous.end

			rightDistance := tok.start - cursor
			if leftDistance <= rightDistance {
				return min(got[i-1].end+leftDistance, got[i].start)
			}

			return max(got[i].start-rightDistance, got[i-1].end)
		}
	}

	trailing := cursor - want[len(want)-1].end

	return min(got[len(got)-1].end+trailing, len(after))
}

// AdjustCursorOffset maps a valid byte offset from before into after. The
// sources must represent the same semantic token sequence.
func AdjustCursorOffset(before, after []byte, cursor int) (int, error) {
	if cursor < 0 || cursor > len(before) {
		return 0, fmt.Errorf("invalid cursor offset %d for source of %d bytes", cursor, len(before))
	}

	if err := verifySemanticTokens(before, after); err != nil {
		return 0, fmt.Errorf("cannot map cursor across semantic changes: %w", err)
	}

	return mapCursorOffset(before, after, cursor), nil
}

func positionedSemanticTokens(source []byte) []positionedSemanticToken {
	tokens := lexer.Tokenize(source)

	out := make([]positionedSemanticToken, 0, len(tokens))
	for _, tok := range tokens {
		if nonSemanticFormattingToken(tok.Kind) {
			continue
		}

		out = append(out, positionedSemanticToken{
			semanticToken: semanticToken{kind: tok.Kind, text: append([]byte(nil), tok.Text(source)...)},
			start:         tok.Start.Offset,
			end:           tok.End.Offset,
		})
	}

	return out
}

// SourceWithCursor is a convenience wrapper around New and
// Formatter.FormatSourceWithCursor.
func SourceWithCursor(source []byte, cfg config.Config, cursorOffset int) (CursorResult, error) {
	formatter, err := New(cfg)
	if err != nil {
		return CursorResult{}, err
	}

	return formatter.FormatSourceWithCursor(source, cursorOffset)
}
