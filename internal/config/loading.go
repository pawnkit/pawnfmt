package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// LoadFile reads, decodes, defaults, and validates a config file.
func LoadFile(path string) (Config, error) {
	return loadFile(path, nil, make(map[string]int))
}

func loadFile(path string, chain []string, visiting map[string]int) (Config, error) {
	canonical, err := canonicalConfigPath(path)
	if err != nil {
		return Config{}, err
	}
	if start, exists := visiting[canonical]; exists {
		cycle := append(append([]string(nil), chain[start:]...), canonical)
		return Config{}, fmt.Errorf("config inheritance cycle: %s", strings.Join(cycle, " -> "))
	}

	visiting[canonical] = len(chain)
	chain = append(chain, canonical)
	defer delete(visiting, canonical)

	data, err := os.ReadFile(canonical)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", canonical, err)
	}

	child := Config{}
	if err := decodeFile(canonical, data, &child); err != nil {
		return Config{}, err
	}

	cfg := Default()
	if child.Extends != "" {
		parentPath := child.Extends
		if !filepath.IsAbs(parentPath) {
			parentPath = filepath.Join(filepath.Dir(canonical), parentPath)
		}

		cfg, err = loadFile(parentPath, chain, visiting)
		if err != nil {
			return Config{}, fmt.Errorf("extend %s: %w", canonical, err)
		}
	}

	cfg.Extends = ""
	if err := decodeFile(canonical, data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func canonicalConfigPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved, nil
	}

	return filepath.Clean(abs), nil
}

func decodeFile(path string, data []byte, out *Config) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".toml":
		return decodeTOMLStrict(data, out)
	case ".yaml", ".yml":
		return decodeYAMLStrict(data, out)
	case ".json":
		return decodeJSONStrict(data, out)
	default:
		candidate := *out
		if err := decodeTOMLStrict(data, &candidate); err == nil {
			*out = candidate
			return nil
		}

		candidate = *out
		if err := decodeJSONStrict(data, &candidate); err == nil {
			*out = candidate
			return nil
		}

		candidate = *out
		if err := decodeYAMLStrict(data, &candidate); err == nil {
			*out = candidate
			return nil
		}

		return fmt.Errorf("unsupported config file extension %q", ext)
	}
}

func decodeJSONStrict(data []byte, out *Config) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode json config: %w", err)
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("decode json config: multiple JSON values are not allowed")
		}

		return fmt.Errorf("decode json config: %w", err)
	}

	return nil
}

func decodeTOMLStrict(data []byte, out *Config) error {
	md, err := toml.Decode(string(data), out)
	if err != nil {
		return fmt.Errorf("decode toml config: %w", err)
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		return fmt.Errorf("decode toml config: unknown key %q", undecoded[0].String())
	}

	return nil
}

func decodeYAMLStrict(data []byte, out *Config) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err := dec.Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("decode yaml config: %w", err)
	}

	return nil
}

// ConfigFileNames lists the config file names pawnfmt looks for.
var ConfigFileNames = []string{"pawnfmt.toml", "pawnfmt.yaml", "pawnfmt.yml", "pawnfmt.json"}

// Discover walks upward from startDir looking for a config file.
func Discover(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve start directory: %w", err)
	}

	for {
		for _, name := range ConfigFileNames {
			candidate := filepath.Join(dir, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}

		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return "", nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}

		dir = parent
	}
}
