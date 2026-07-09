# pawnfmt

A fast, deterministic formatter for Pawn (SA-MP / open.mp) source: `.pwn`
and `.inc` files.

## Documentation

These docs cover the parts of `pawnfmt` people usually need once they move
past the quick start:

- [Configuration reference](docs/configuration.md)
- [Project notes](docs/project-notes.md)

For a ready-to-edit config file, run:

```sh
pawnfmt --init-config
```

That writes a commented `pawnfmt.toml` with every option set to its default
value.

## Install

Download a prebuilt binary from
[GitHub Releases](https://github.com/pawnkit/pawnfmt/releases). Release
archives are built for Linux, macOS, and Windows on amd64 and arm64.

If you already have Go installed, you can also install from source:

```sh
go install github.com/pawnkit/pawnfmt/cmd/pawnfmt@latest
```

Use a version tag instead of `latest` when you want reproducible CI builds:

```sh
go install github.com/pawnkit/pawnfmt/cmd/pawnfmt@<tag>
```

## Usage

Format files in place:

```sh
pawnfmt --write gamemodes includes
```

Check formatting without changing files:

```sh
pawnfmt --check gamemodes includes
```

Print a diff:

```sh
pawnfmt --diff script.pwn
```

## Editor / CI Integration

Call `pawnfmt` directly from your editor or CI:

```sh
pawnfmt --check gamemodes includes
```

For Git hooks, choose the hook manager that fits your project:

- [Lefthook](https://lefthook.dev/)
- [prek](https://prek.j178.dev/)
- [pre-commit](https://pre-commit.com)

Lefthook:

```yaml
pre-commit:
  jobs:
    - run: pawnfmt --write {staged_files}
      glob: "*.{pwn,inc}"
      stage_fixed: true
```

prek:

```yaml
repos:
  - repo: local
    hooks:
      - id: pawnfmt
        name: pawnfmt
        entry: pawnfmt --write
        language: system
        files: \.(pwn|inc)$
```

pre-commit:

```yaml
- repo: https://github.com/pawnkit/pawnfmt
  rev: <tag>
  hooks:
    - id: pawnfmt
```

A plain Git hook also works if you do not want a hook manager:

```sh
#!/bin/sh
pawnfmt --check gamemodes includes
```

Save it as `.git/hooks/pre-commit` and make it executable.
