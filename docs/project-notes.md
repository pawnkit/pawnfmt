# Project notes

`pawnfmt` is a Go CLI formatter for Pawn `.pwn` and `.inc` files. It formats deterministically, supports stdin and file modes, and can be used locally or in CI.

## Common commands

```sh
go build -o ./bin/pawnfmt/pawnfmt ./cmd/pawnfmt
go test ./...
go vet ./...
go test ./internal/bench/... -bench=. -benchmem -run=^$
```

If you use Task:

```sh
task
task test
task vet
task bench
```

## CLI behavior

For one file, running without `--write`, `--check`, or `--diff` prints the formatted output to stdout:

```sh
pawnfmt script.pwn
```

For multiple files or directories, choose an explicit mode:

```sh
pawnfmt --write gamemodes includes
pawnfmt --check gamemodes includes
pawnfmt --diff script.pwn
```

Other handy flags:

- `--stdin` reads source from stdin and writes formatted source to stdout.
- `--color=auto|always|never` controls colour in output. `auto` colours terminal output and keeps redirected output plain.
- `--debug-tokens` prints the lexer token stream for one input.
- `--debug-cst` prints the parsed CST for one input.
- `--debug-format-doc` prints the formatter's intermediate document tree.
- `--version` prints version information.

Debug modes require exactly one input file, unless they are used with `--stdin`.

## Files and ignores

Directory walks only collect `.pwn` and `.inc` files.

These directories are skipped by default:

```text
.git
.svn
.hg
node_modules
vendor
.cache
dist
build
```

`pawnfmt` also reads `.gitignore` and `.pawnfmtignore` files while walking directories. Pass `--no-gitignore` to ignore those files.

Ignore matching supports normal gitignore-style patterns, including negation with `!`, directory-only patterns with a trailing slash, anchored patterns with a leading slash, and `**`.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success. |
| `1` | `--check` found files that would change. |
| `2` | Formatting failed, file walking failed, or an invalid multi-file mode was used. |
| `3` | CLI or config error. |
| `4` | Internal error. |

## Package map

| Path | Purpose |
| --- | --- |
| `cmd/pawnfmt` | CLI parsing, config resolution, file walking, diff/write/check behavior. |
| `internal/config` | Config schema, defaults, validation, loading, and config template generation. |
| `internal/format` | Formatter logic and formatting tests. |
| `internal/printer` | Document printer used by the formatter. |
| `internal/trivia` | Trivia handling helpers. |
| `internal/check` | Corpus/idempotence checking helpers. |
| `internal/bench` | Benchmark fixtures and benchmark tests. |
| `testdata/input` | Formatter input fixtures. |
| `testdata/expected` | Expected formatted output fixtures. |
| `testdata/idempotence` | Inputs used to check stable repeated formatting. |
| `testdata/real-world` | Optional real-world corpus material and reporting. |

## Notes for contributors

Keep fixtures paired: when behavior changes, update the matching files under `testdata/input` and `testdata/expected`.

For formatter changes, run `go test ./...`. For performance-sensitive changes, run the benchmark task too.
