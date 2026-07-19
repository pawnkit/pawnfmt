# Contributing

PawnKit is maintained by volunteers, so reviews may take a little time.

Formatting bugs are easiest to discuss with a small input and the output you
expected. Fixes should add an idempotence or golden test.

Run the local checks before opening a pull request:

```sh
task check
```

Formatting must be deterministic. Avoid style changes based only on personal
preference; explain the Pawn construct or project compatibility issue behind
the change.
