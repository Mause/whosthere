package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/ramonvermeulen/whosthere/internal/paths"
)

const (
	defaultConfigFileName = "config.yaml"
	// Environment variable to override config file path
	configEnvVar = "WHOSTHERE_CONFIG"
)

var ErrConfigNil = errors.New("config is nil")

// Load resolves the config path, reads/creates YAML, and returns the merged config.
func Load(pathOverride string) (*Config, error) {
	resolvedPath, err := resolveConfigPath(pathOverride)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	if err := ensureConfigFile(resolvedPath, cfg); err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(resolvedPath)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if err := validateAndNormalize(cfg); err != nil {
		return cfg, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func ensureConfigFile(path string, defaults *Config) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data := renderDefaultConfigFile(defaults)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}

	return nil
}

// resolveConfigPath returns the path using precedence: flag override > env var > XDG default.
func resolveConfigPath(pathOverride string) (string, error) {
	if pathOverride != "" {
		return pathOverride, nil
	}

	if env := os.Getenv(configEnvVar); env != "" {
		return env, nil
	}

	dir, err := paths.ConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	return filepath.Join(dir, defaultConfigFileName), nil
}

func validateAndNormalize(cfg *Config) error {
	if cfg == nil {
		return ErrConfigNil
	}

	return cfg.validateAndNormalize()
}

func renderDefaultConfigFile(cfg *Config) []byte {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	themeCfg := cfg.Theme

	return []byte(fmt.Sprintf(`scan_interval: %s
scan_duration: %s
splash:
  enabled: %t
  delay: %s
theme:
  # pick a built-in theme by name; unknown names fall back to "default"
  name: %s
  # default palette (uncomment and set name: custom to tweak)
  # primitive_background_color: "#000a1a"
  # contrast_background_color: "#001a33"
  # more_contrast_background_color: "#003366"
  # border_color: "#0088ff"
  # title_color: "#00ffff"
  # graphics_color: "#00ffaa"
  # primary_text_color: "#cceeff"
  # secondary_text_color: "#6699ff"
  # tertiary_text_color: "#ffaa00"
  # inverse_text_color: "#000a1a"
  # contrast_secondary_text_color: "#88ddff"
scanners:
  mdns:
    enabled: %t
  ssdp:
    enabled: %t
  arp:
    enabled: %t
`,
		cfg.ScanInterval,
		cfg.ScanDuration,
		cfg.Splash.Enabled,
		cfg.Splash.Delay,
		themeCfg.Name,
		cfg.Scanners.MDNS.Enabled,
		cfg.Scanners.SSDP.Enabled,
		cfg.Scanners.ARP.Enabled,
	))
}

// Save writes the config to the specified path (or resolves the default path if empty).
func Save(cfg *Config, pathOverride string) error {
	if cfg == nil {
		return ErrConfigNil
	}

	resolvedPath, err := resolveConfigPath(pathOverride)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	data := renderDefaultConfigFile(cfg)

	if err := os.WriteFile(resolvedPath, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
