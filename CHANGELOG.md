# Changelog

Notable changes are recorded here.

## 1.4.2 - 2026-07-29

### Fixed

- Do not duplicate a leading file header during range formatting.

## 1.4.1 - 2026-07-29

### Fixed

- Preserve grouped alignment when formatting a range on save.

## 1.4.0 - 2026-07-29

### Changed

- Align values in consecutive `#define` groups by default.

## 1.3.9 - 2026-07-29

### Fixed

- Format editor ranges that span more than one top-level declaration.

## 1.3.8 - 2026-07-29

### Performance

- Verify range formatting within the changed declaration instead of reparsing
  the full output.

## 1.3.7 - 2026-07-29

### Performance

- Format selected syntax units without formatting the full file first.
- Added file and range benchmarks for large projects.

## 1.3.6 - 2026-07-25

### Fixed

- Stopped re-tokenizing the source and formatted output a second time
  during semantic verification; it now reuses the tokens each parse
  already produced.

### Changed

- Added the repository support record with CI validation.

## 1.3.5 - 2026-07-23

### Changed

- Updated to the current Pawn project release.

## 1.3.4 - 2026-07-23

### Fixed

- Updated project loading for consistent paths on Windows.

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
