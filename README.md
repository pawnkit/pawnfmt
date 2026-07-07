# pawnfmt

A fast, deterministic formatter for Pawn (SA-MP / open.mp) source: `.pwn`
and `.inc` files.

## Editor / CI integration

A [pre-commit](https://pre-commit.com) hook is available via
`.pre-commit-hooks.yaml`:

```yaml
- repo: https://github.com/pawnkit/pawnfmt
  rev: <tag>
  hooks:
    - id: pawnfmt
```
