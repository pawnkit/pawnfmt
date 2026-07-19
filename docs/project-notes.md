# Project notes

## Development

```sh
go build -o ./bin/pawnfmt/pawnfmt ./cmd/pawnfmt
go test ./...
go vet ./...
go test ./internal/bench/... -bench=. -benchmem -run=^$
```

The matching Task commands are `task`, `task test`, `task vet`, and `task bench`.

Formatter fixtures come in pairs under `testdata/input` and `testdata/expected`. Update both when intended output changes. Keep the idempotence suite passing.

## CLI details

One file without a mode flag is formatted to standard output:

```sh
pawnfmt script.pwn
```

Multiple files or directories require `--write`, `--check`, or `--diff`.

Range formatting uses `--range-start` and `--range-end` byte offsets. It expands the request to the smallest safe syntax unit and leaves bytes outside that unit untouched.

`--cursor-offset N --output-format=json` returns formatted source and the adjusted cursor. Include sorting is disabled for cursor-aware requests because moving directives would make the cursor mapping unreliable.

The debug flags `--debug-tokens`, `--debug-cst`, and `--debug-format-doc` accept one file or standard input.

## File discovery

Directory walks include `.pwn` and `.inc` files. They skip common metadata, dependency, cache, build, and output directories. `.gitignore` and `.pawnfmtignore` rules also apply unless `--no-gitignore` is set.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | `--check` found a file that would change |
| `2` | Formatting, file walking, or mode failure |
| `3` | CLI or configuration error |
| `4` | Internal error |
