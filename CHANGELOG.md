# Changelog

Notable changes are recorded here.

## 1.3.3 - 2026-07-22

### Fixed

- Format legacy statement macros and inline operator blocks.
- Preserve timer dimensions, iterator capacities, generic suffixes, and postfix `char` operators.
- Format compact modulo expressions without changing their meaning.

## 1.3.2 - 2026-07-22

### Fixed

- Updated parser compatibility for packed dimensions and conditional arguments.
- Preserved PawnPlus generic tags while formatting.
- Kept one-line `do ... while` macros on one line.

## 1.3.1 - 2026-07-22

### Added

- Added tolerant formatting to the public Go API for editor clients.

## 1.3.0 - 2026-07-21

### Added

- Added a public API for formatting a selected top-level syntax unit.

## Unreleased

### Added

- Public Go formatting API.
- Automatic Pawn project discovery when paths are omitted.
- Project documentation, contribution guidance, and security policy.

### Changed

- Project discovery now uses the public `pawn-project` v0.1.0 module.
