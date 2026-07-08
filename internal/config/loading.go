package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	cfg := Default()
	if err := decodeFile(path, data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func decodeFile(path string, data []byte, out *Config) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".toml":
		return decodeTOMLStrict(data, out)
	case ".yaml", ".yml":
		return decodeYAMLStrict(data, out)
	default:
		if err := decodeTOMLStrict(data, out); err == nil {
			return nil
		}

		if err := decodeYAMLStrict(data, out); err == nil {
			return nil
		}

		return fmt.Errorf("unsupported config file extension %q", ext)
	}
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

var ConfigFileNames = []string{"pawnfmt.toml", "pawnfmt.yaml", "pawnfmt.yml"}

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
