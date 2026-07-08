# Configuration

`pawnfmt` looks for config files named `pawnfmt.toml`, `pawnfmt.yaml`, or `pawnfmt.yml`.

By default it starts from the first input path, walks upward, and stops at the first config it finds. If it reaches a Git root without finding one, it uses the built-in defaults.

Useful flags:

- `--config path/to/pawnfmt.toml` uses a specific config file.
- `--no-config` ignores discovered config files and uses defaults.
- `--print-config` prints the resolved config as TOML.
- `--init-config [path]` writes a fully commented config file. It refuses to overwrite an existing file.
- `--stdin-filename path/to/file.pwn` gives `--stdin` a location to use for config discovery.

Config files are strict. Unknown keys fail fast instead of being ignored.

## Example

```toml
line_width = 100
indent_style = "space"
indent_width = 4
brace_style = "allman"
single_statement_braces = "always"
enum_trailing_comma = "always"

exclude = ["vendor/*", "generated/*"]
```

## Options

| Option | Default | Values | Notes |
| --- | --- | --- | --- |
| `line_width` | `100` | integer, at least `20` | Target wrap width. |
| `indent_style` | `"space"` | `"space"`, `"tab"` | Uses tabs for indentation when set to `"tab"`. |
| `indent_width` | `4` | integer, at least `1` | Spaces per indent level. Ignored for tab indentation. |
| `continuation_indent_width` | `0` | integer, at least `0` | Extra indent for wrapped lines. `0` means use `indent_width`. |
| `newline_style` | `"auto"` | `"auto"`, `"lf"`, `"crlf"` | `auto` keeps the file's existing line ending style. |
| `insert_final_newline` | `true` | boolean | Ends output with one newline. |
| `trim_trailing_whitespace` | `true` | boolean | Removes trailing spaces and tabs. |
| `brace_style` | `"allman"` | `"1tbs"`, `"allman"`, `"whitesmiths"` | Controls opening brace placement. |
| `keep_simple_statements_single_line` | `true` | boolean | Keeps short unbraced bodies like `if (x) return 1;` on one line. |
| `single_statement_braces` | `"always"` | `"preserve"`, `"always"`, `"never"` | Adds, keeps, or removes braces around single-statement control bodies. |
| `indent_case_contents` | `true` | boolean | Indents the body under `case` and `default`. |
| `indent_case_labels` | `true` | boolean | Indents `case` and `default` labels inside `switch`. |
| `indent_goto_labels` | `true` | boolean | Keeps goto labels at the current statement indent. |
| `empty_line_between_top_level_declarations` | `true` | boolean | Adds breathing room between top-level declaration groups. |
| `space_around_operators` | `true` | boolean | Formats `a + b` instead of `a+b`. |
| `space_after_comma` | `true` | boolean | Formats `a, b` instead of `a,b`. |
| `space_inside_parens` | `false` | boolean | Adds spaces inside `( ... )`. |
| `space_inside_brackets` | `false` | boolean | Adds spaces inside `[ ... ]`. |
| `space_inside_braces` | `false` | boolean | Adds spaces inside array literal braces. |
| `space_before_function_paren` | `false` | boolean | Adds a space before function parameter lists. |
| `space_before_array_brackets` | `false` | boolean | Adds a space before array dimensions. |
| `semicolons` | `"preserve"` | `"preserve"`, `"always"` | Applies to optional enum declaration semicolons. |
| `directive_indent` | `"keep_in_block"` | `"none"`, `"keep_in_block"` | Controls whether preprocessor lines keep block indentation. |
| `directive_spacing` | `true` | boolean | Adds a space after `#` before directive text. |
| `indent_nested_directives` | `false` | boolean | Indents a top-level `#if` branch's contents, including nested `#if`s. |
| `align_enum_fields` | `false` | boolean | Aligns enum entry values. |
| `align_consecutive_declarations` | `false` | boolean | Aligns initialized declarations in contiguous runs. |
| `align_consecutive_macros` | `false` | boolean | Aligns macro values in contiguous `#define` runs. |
| `align_trailing_comments` | `false` | boolean | Aligns trailing `//` comments in contiguous runs. |
| `enum_trailing_comma` | `"always"` | `"preserve"`, `"always"` | Controls the final comma in enum bodies.
| `tag_colon_spacing` | `"tight"` | `"tight"`, `"compact"`, `"preserve"` | `tight` formats tag prefixes like `Float: x`; `compact` like `Float:x`. |
| `multiline_function_params` | `"auto"` | `"auto"`, `"one_per_line"`, `"bin_pack"` | Controls wrapping for function parameters. |
| `multiline_call_args` | `"auto"` | `"auto"`, `"one_per_line"`, `"bin_pack"` | Controls wrapping for call arguments. |
| `break_binary_operator` | `"after"` | `"after"`, `"before"` | Places wrapped binary operators at the end of the old line or start of the new one. |
| `format_disabled_regions` | `false` | boolean | Formats code inside `// pawnfmt off` and `// pawnfmt on` regions anyway. |
| `blank_lines_after_include_block` | `true` | boolean | Keeps one blank line after the top include block. |
| `blank_lines_between_publics` | `true` | boolean | Keeps adjacent `public` functions separated. |
| `sort_includes` | `false` | boolean | Sorts contiguous top-level include runs by path. |
| `group_includes_by_brackets` | `false` | boolean | With `sort_includes`, puts angle-bracket includes before quoted includes. |
| `collapse_blank_lines` | `true` | boolean | Caps long blank-line runs. |
| `max_blank_lines` | `2` | integer, at least `0` | Maximum blank lines kept when collapsing is enabled. |
| `include` | `[]` | list of glob strings | Limits directory formatting to matching files. |
| `exclude` | `[]` | list of glob strings | Skips matching files. Exclude wins over include. |

## Include and exclude globs

`include` and `exclude` only apply when formatting directories. They do not affect `--stdin` or direct single-file formatting.

Patterns are matched against both the file name and the full path:

```toml
include = ["*.pwn", "gamemodes/*"]
exclude = ["generated/*", "vendor/*"]
```

When `include` is empty, all `.pwn` and `.inc` files are eligible. When `include` has entries, only matching files are eligible. `exclude` always wins.
