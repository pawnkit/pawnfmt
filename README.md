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

## Editor / CI integration

A [pre-commit](https://pre-commit.com) hook is available via
`.pre-commit-hooks.yaml`:

```yaml
- repo: https://github.com/pawnkit/pawnfmt
  rev: <tag>
  hooks:
    - id: pawnfmt
```
