package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestDefaultIsValid(t *testing.T) {
	t.Parallel()

	if err := config.Default().Validate(); err != nil {
		t.Fatalf("Default() must validate cleanly: %v", err)
	}
}

func TestApplyDefaultsBackfillsStringAndWidthFields(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.ApplyDefaults()

	defaults := config.Default()

	fields := []struct {
		name string
		got  any
		want any
	}{
		{"LineWidth", cfg.LineWidth, defaults.LineWidth},
		{"IndentWidth", cfg.IndentWidth, defaults.IndentWidth},
		{"IndentStyle", cfg.IndentStyle, defaults.IndentStyle},
		{"NewlineStyle", cfg.NewlineStyle, defaults.NewlineStyle},
		{"ParseMode", cfg.ParseMode, defaults.ParseMode},
		{"BraceStyle", cfg.BraceStyle, defaults.BraceStyle},
		{"Semicolons", cfg.Semicolons, defaults.Semicolons},
		{"SingleStatementBraces", cfg.SingleStatementBraces, defaults.SingleStatementBraces},
		{"DirectiveIndent", cfg.DirectiveIndent, defaults.DirectiveIndent},
		{"EnumTrailingComma", cfg.EnumTrailingComma, defaults.EnumTrailingComma},
		{"TagColonSpacing", cfg.TagColonSpacing, defaults.TagColonSpacing},
		{"MultilineFunctionParams", cfg.MultilineFunctionParams, defaults.MultilineFunctionParams},
		{"MultilineCallArgs", cfg.MultilineCallArgs, defaults.MultilineCallArgs},
	}

	for _, field := range fields {
		if field.got != field.want {
			t.Errorf("%s = %v, want %v", field.name, field.got, field.want)
		}
	}
}

func TestApplyDefaultsLeavesBoolAndMaxBlankLinesAlone(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.ApplyDefaults()

	if cfg.InsertFinalNewline {
		t.Error("InsertFinalNewline should stay false on a bare Config{}, even though Default() is true")
	}

	if cfg.MaxBlankLines != 0 {
		t.Errorf("MaxBlankLines = %d, want 0 (unmodified)", cfg.MaxBlankLines)
	}
}

func TestValidateRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mutate func(*config.Config)
	}{
		{"line width too small", func(c *config.Config) { c.LineWidth = 19 }},
		{"indent width zero", func(c *config.Config) { c.IndentWidth = 0 }},
		{"indent width negative", func(c *config.Config) { c.IndentWidth = -1 }},
		{"max blank lines negative", func(c *config.Config) { c.MaxBlankLines = -1 }},
		{"bad indent style", func(c *config.Config) { c.IndentStyle = "spaces" }},
		{"bad newline style", func(c *config.Config) { c.NewlineStyle = "unix" }},
		{"bad brace style", func(c *config.Config) { c.BraceStyle = "kandr" }},
		{"bad semicolons", func(c *config.Config) { c.Semicolons = "sometimes" }},
		{"bad single statement braces", func(c *config.Config) { c.SingleStatementBraces = "maybe" }},
		{"bad directive indent", func(c *config.Config) { c.DirectiveIndent = "auto" }},
		{"bad enum trailing comma", func(c *config.Config) { c.EnumTrailingComma = "sometimes" }},
		{"bad tag colon spacing", func(c *config.Config) { c.TagColonSpacing = "loose" }},
		{"bad multiline function params", func(c *config.Config) { c.MultilineFunctionParams = "compact" }},
		{"bad multiline call args", func(c *config.Config) { c.MultilineCallArgs = "compact" }},
		{"bad break binary operator", func(c *config.Config) { c.BreakBinaryOperator = "sideways" }},
		{"continuation indent width negative", func(c *config.Config) { c.ContinuationIndentWidth = -1 }},
		{"bad parse mode", func(c *config.Config) { c.ParseMode = "hopeful" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Default()
			tc.mutate(&cfg)

			if err := cfg.Validate(); err == nil {
				t.Fatal("expected Validate to reject this config, got nil error")
			}
		})
	}
}

func TestValidateAcceptsMaxBlankLinesZero(t *testing.T) {
	t.Parallel()

	cfg := config.Default()

	cfg.MaxBlankLines = 0
	if err := cfg.Validate(); err != nil {
		t.Fatalf("max_blank_lines=0 should be valid (means never preserve blank lines): %v", err)
	}
}

func TestLoadFileTOML(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "pawnfmt.toml", "line_width = 80\nindent_style = \"tab\"\n")

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.LineWidth != 80 {
		t.Errorf("LineWidth = %d, want 80", cfg.LineWidth)
	}

	if cfg.IndentStyle != config.IndentStyleTab {
		t.Errorf("IndentStyle = %q, want tab", cfg.IndentStyle)
	}

	if cfg.SpaceAfterComma != config.Default().SpaceAfterComma {
		t.Errorf("SpaceAfterComma should keep its default when unset in the file")
	}
}

func TestLoadFileYAML(t *testing.T) {
	t.Parallel()

	for _, ext := range []string{"pawnfmt.yaml", "pawnfmt.yml"} {
		t.Run(ext, func(t *testing.T) {
			t.Parallel()
			path := writeFile(t, ext, "line_width: 72\nindent_width: 2\n")

			cfg, err := config.LoadFile(path)
			if err != nil {
				t.Fatalf("LoadFile: %v", err)
			}

			if cfg.LineWidth != 72 {
				t.Errorf("LineWidth = %d, want 72", cfg.LineWidth)
			}

			if cfg.IndentWidth != 2 {
				t.Errorf("IndentWidth = %d, want 2", cfg.IndentWidth)
			}
		})
	}
}

func TestLoadFileJSON(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "pawnfmt.json", "{\n  \"line_width\": 88,\n  \"indent_width\": 3\n}\n")

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.LineWidth != 88 {
		t.Errorf("LineWidth = %d, want 88", cfg.LineWidth)
	}

	if cfg.IndentWidth != 3 {
		t.Errorf("IndentWidth = %d, want 3", cfg.IndentWidth)
	}

	if cfg.SpaceAfterComma != config.Default().SpaceAfterComma {
		t.Errorf("SpaceAfterComma should keep its default when unset in the file")
	}
}

func TestLoadFileExtendsParentAndOverridesValues(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	parent := filepath.Join(dir, "base.toml")

	childDir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(parent, []byte("line_width = 120\nindent_width = 2\nsort_includes = true\nmax_blank_lines = 4\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	child := filepath.Join(childDir, "pawnfmt.toml")
	if err := os.WriteFile(child, []byte("extends = \"../base.toml\"\nindent_width = 6\nsort_includes = false\nmax_blank_lines = 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFile(child)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.LineWidth != 120 || cfg.IndentWidth != 6 {
		t.Fatalf("inherited/overridden values = line_width %d, indent_width %d", cfg.LineWidth, cfg.IndentWidth)
	}

	if cfg.SortIncludes {
		t.Fatal("explicit false in child must override true in parent")
	}

	if cfg.MaxBlankLines != 0 {
		t.Fatalf("explicit zero in child must override parent, got %d", cfg.MaxBlankLines)
	}
}

func TestLoadFileExtendsAcrossConfigFormats(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	parent := filepath.Join(dir, "base.yaml")
	child := filepath.Join(dir, "pawnfmt.json")

	if err := os.WriteFile(parent, []byte("line_width: 120\nindent_width: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(child, []byte("{\"extends\": \"base.yaml\", \"indent_width\": 3}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFile(child)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.LineWidth != 120 || cfg.IndentWidth != 3 {
		t.Fatalf("cross-format inheritance failed: line_width %d, indent_width %d", cfg.LineWidth, cfg.IndentWidth)
	}
}

func TestLoadFileRejectsInheritanceCycle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.toml")
	b := filepath.Join(dir, "b.json")

	if err := os.WriteFile(a, []byte("extends = \"b.json\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(b, []byte("{\"extends\": \"a.toml\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := config.LoadFile(a)
	if err == nil || !strings.Contains(err.Error(), "inheritance cycle") {
		t.Fatalf("expected an inheritance-cycle error, got %v", err)
	}
}

func TestLoadFileReportsMissingExtendedConfig(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "pawnfmt.toml", "extends = \"missing.toml\"\n")

	_, err := config.LoadFile(path)
	if err == nil || !strings.Contains(err.Error(), "missing.toml") {
		t.Fatalf("expected missing parent path in error, got %v", err)
	}
}

func TestLoadFileMaxBlankLinesZeroSurvivesRoundTrip(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "pawnfmt.toml", "max_blank_lines = 0\n")

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.MaxBlankLines != 0 {
		t.Errorf("MaxBlankLines = %d, want 0 (explicit zero must not be overwritten by the default)", cfg.MaxBlankLines)
	}
}

func TestLoadFileUnknownExtensionSniffsFormat(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "pawnfmt.conf", "line_width = 90\n")

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.LineWidth != 90 {
		t.Errorf("LineWidth = %d, want 90", cfg.LineWidth)
	}
}

func TestLoadFileMissing(t *testing.T) {
	t.Parallel()

	_, err := config.LoadFile(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err == nil {
		t.Fatal("expected an error for a missing config file")
	}
}

func TestLoadFileMalformed(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "pawnfmt.toml", "this is not [ valid toml\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected an error for malformed TOML")
	}
}

func TestLoadFileTOMLRejectsUnknownKey(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "pawnfmt.toml", "lin_width = 80\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject an unknown TOML key (typo of line_width)")
	}
}

func TestLoadFileYAMLRejectsUnknownKey(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "pawnfmt.yaml", "lin_width: 80\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject an unknown YAML key (typo of line_width)")
	}
}

func TestLoadFileJSONRejectsUnknownKey(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "pawnfmt.json", "{\"lin_width\": 80}\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject an unknown JSON key (typo of line_width)")
	}
}

func TestLoadFileJSONRejectsMultipleValues(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "pawnfmt.json", "{\"line_width\": 80} {\"line_width\": 90}\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject multiple JSON values")
	}
}

func TestLoadFileInvalidAfterDefaults(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "pawnfmt.toml", "line_width = 5\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject a config that fails Validate")
	}
}

func TestDiscoverFindsNearestConfig(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	sub := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	nearest := filepath.Join(root, "a", "b", "pawnfmt.toml")
	if err := os.WriteFile(nearest, []byte("line_width = 60\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	farther := filepath.Join(root, "pawnfmt.toml")
	if err := os.WriteFile(farther, []byte("line_width = 120\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := config.Discover(sub)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if found != nearest {
		t.Errorf("Discover found %q, want %q", found, nearest)
	}
}

func TestDiscoverFindsJSONConfig(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	sub := filepath.Join(root, "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(root, "pawnfmt.json")
	if err := os.WriteFile(want, []byte("{\"line_width\": 80}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := config.Discover(sub)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if found != want {
		t.Errorf("Discover found %q, want %q", found, want)
	}
}

func TestDiscoverStopsAtGit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(root, "a", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "pawnfmt.toml"), []byte("line_width = 60\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := config.Discover(sub)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if found != "" {
		t.Errorf("Discover found %q, want \"\" (should have stopped at .git)", found)
	}
}

func TestDiscoverConfigInSameDirAsGitStillWins(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(root, "pawnfmt.toml")
	if err := os.WriteFile(cfgPath, []byte("line_width = 60\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := config.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if found != cfgPath {
		t.Errorf("Discover found %q, want %q (a repo root's own config must still be found)", found, cfgPath)
	}
}

func TestDiscoverNoneFound(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	found, err := config.Discover(sub)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if found != "" {
		t.Errorf("Discover found %q, want \"\" in a tree with no config or .git", found)
	}
}

func TestDiscoverNamePriority(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pawnfmt.yaml"), []byte("line_width: 60\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tomlPath := filepath.Join(root, "pawnfmt.toml")
	if err := os.WriteFile(tomlPath, []byte("line_width = 90\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := config.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if found != tomlPath {
		t.Errorf("Discover found %q, want %q (.toml should take priority over .yaml)", found, tomlPath)
	}
}

func TestResolveNewline(t *testing.T) {
	t.Parallel()

	cases := []struct {
		style    config.NewlineStyle
		detected string
		want     string
	}{
		{config.NewlineStyleCRLF, "\n", "\r\n"},
		{config.NewlineStyleLF, "\r\n", "\n"},
		{config.NewlineStyleAuto, "\r\n", "\r\n"},
		{config.NewlineStyleAuto, "\n", "\n"},
		{config.NewlineStyleAuto, "", "\n"},
	}
	for _, tc := range cases {
		got := config.ResolveNewline(tc.style, tc.detected)
		if got != tc.want {
			t.Errorf("ResolveNewline(%q, %q) = %q, want %q", tc.style, tc.detected, got, tc.want)
		}
	}
}

func TestDefaultTOMLRoundTrips(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "pawnfmt.toml", config.DefaultTOML())

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile(generated config): %v\n---\n%s", err, config.DefaultTOML())
	}

	if len(cfg.Include) != 0 || len(cfg.Exclude) != 0 {
		t.Fatalf("expected empty Include/Exclude, got %+v/%+v", cfg.Include, cfg.Exclude)
	}

	cfg.Include, cfg.Exclude = nil, nil
	if !reflect.DeepEqual(cfg, config.Default()) {
		t.Fatalf("generated config != Default()\ngenerated: %+v\ndefault:   %+v", cfg, config.Default())
	}
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}
