package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pawnkit/pawnfmt/internal/config"
)

func TestDefaultIsValid(t *testing.T) {
	if err := config.Default().Validate(); err != nil {
		t.Fatalf("Default() must validate cleanly: %v", err)
	}
}

func TestApplyDefaultsBackfillsStringAndWidthFields(t *testing.T) {
	cfg := config.Config{}
	cfg.ApplyDefaults()

	defaults := config.Default()
	if cfg.LineWidth != defaults.LineWidth {
		t.Errorf("LineWidth = %d, want %d", cfg.LineWidth, defaults.LineWidth)
	}

	if cfg.IndentWidth != defaults.IndentWidth {
		t.Errorf("IndentWidth = %d, want %d", cfg.IndentWidth, defaults.IndentWidth)
	}

	if cfg.IndentStyle != defaults.IndentStyle {
		t.Errorf("IndentStyle = %q, want %q", cfg.IndentStyle, defaults.IndentStyle)
	}

	if cfg.NewlineStyle != defaults.NewlineStyle {
		t.Errorf("NewlineStyle = %q, want %q", cfg.NewlineStyle, defaults.NewlineStyle)
	}

	if cfg.BraceStyle != defaults.BraceStyle {
		t.Errorf("BraceStyle = %q, want %q", cfg.BraceStyle, defaults.BraceStyle)
	}

	if cfg.Semicolons != defaults.Semicolons {
		t.Errorf("Semicolons = %q, want %q", cfg.Semicolons, defaults.Semicolons)
	}

	if cfg.SingleStatementBraces != defaults.SingleStatementBraces {
		t.Errorf("SingleStatementBraces = %q, want %q", cfg.SingleStatementBraces, defaults.SingleStatementBraces)
	}

	if cfg.DirectiveIndent != defaults.DirectiveIndent {
		t.Errorf("DirectiveIndent = %q, want %q", cfg.DirectiveIndent, defaults.DirectiveIndent)
	}

	if cfg.EnumTrailingComma != defaults.EnumTrailingComma {
		t.Errorf("EnumTrailingComma = %q, want %q", cfg.EnumTrailingComma, defaults.EnumTrailingComma)
	}

	if cfg.TagColonSpacing != defaults.TagColonSpacing {
		t.Errorf("TagColonSpacing = %q, want %q", cfg.TagColonSpacing, defaults.TagColonSpacing)
	}

	if cfg.MultilineFunctionParams != defaults.MultilineFunctionParams {
		t.Errorf("MultilineFunctionParams = %q, want %q", cfg.MultilineFunctionParams, defaults.MultilineFunctionParams)
	}

	if cfg.MultilineCallArgs != defaults.MultilineCallArgs {
		t.Errorf("MultilineCallArgs = %q, want %q", cfg.MultilineCallArgs, defaults.MultilineCallArgs)
	}
}

func TestApplyDefaultsLeavesBoolAndMaxBlankLinesAlone(t *testing.T) {
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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Default()
			tc.mutate(&cfg)

			if err := cfg.Validate(); err == nil {
				t.Fatal("expected Validate to reject this config, got nil error")
			}
		})
	}
}

func TestValidateAcceptsMaxBlankLinesZero(t *testing.T) {
	cfg := config.Default()

	cfg.MaxBlankLines = 0
	if err := cfg.Validate(); err != nil {
		t.Fatalf("max_blank_lines=0 should be valid (means never preserve blank lines): %v", err)
	}
}

func TestLoadFileTOML(t *testing.T) {
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
	for _, ext := range []string{"pawnfmt.yaml", "pawnfmt.yml"} {
		t.Run(ext, func(t *testing.T) {
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

func TestLoadFileMaxBlankLinesZeroSurvivesRoundTrip(t *testing.T) {
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
	_, err := config.LoadFile(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err == nil {
		t.Fatal("expected an error for a missing config file")
	}
}

func TestLoadFileMalformed(t *testing.T) {
	path := writeFile(t, "pawnfmt.toml", "this is not [ valid toml\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected an error for malformed TOML")
	}
}

func TestLoadFileTOMLRejectsUnknownKey(t *testing.T) {
	path := writeFile(t, "pawnfmt.toml", "lin_width = 80\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject an unknown TOML key (typo of line_width)")
	}
}

func TestLoadFileYAMLRejectsUnknownKey(t *testing.T) {
	path := writeFile(t, "pawnfmt.yaml", "lin_width: 80\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject an unknown YAML key (typo of line_width)")
	}
}

func TestLoadFileInvalidAfterDefaults(t *testing.T) {
	path := writeFile(t, "pawnfmt.toml", "line_width = 5\n")
	if _, err := config.LoadFile(path); err == nil {
		t.Fatal("expected LoadFile to reject a config that fails Validate")
	}
}

func TestDiscoverFindsNearestConfig(t *testing.T) {
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

func TestDiscoverStopsAtGit(t *testing.T) {
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
