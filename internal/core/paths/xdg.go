package paths

import (
	"os"
	"path/filepath"
)

const (
	appName          = "whosthere"
	xdgConfigDirEnv  = "XDG_CONFIG_HOME"
	xdgStateDirEnv   = "XDG_STATE_HOME"
	defaultConfigDir = ".config"
	defaultStateDir  = ".local/state"
)

// ConfigDir returns the XDG config directory for this app without creating it.
// It follows XDG_CONFIG_HOME when set, otherwise falls back to ~/.config/<appName>.
func ConfigDir() (string, error) {
	base := os.Getenv(xdgConfigDirEnv)
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, defaultConfigDir)
	}
	return filepath.Join(base, appName), nil
}

// StateDir returns the XDG state directory for this app, creating it if
// necessary. It follows XDG_STATE_HOME when set, otherwise falls back to
// ~/.local/state/<appName>.
func StateDir() (string, error) {
	base := os.Getenv(xdgStateDirEnv)
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, defaultStateDir)
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
