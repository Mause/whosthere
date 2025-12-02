package config

const (
	DefaultSplashEnabled = true
	DefaultSplashDelay   = 1
)

// Config captures runtime configuration values loaded from the YAML config file.
type Config struct {
	Splash SplashConfig `yaml:"splash"`
}

// SplashConfig controls the splash screen visibility and timing.
type SplashConfig struct {
	Enabled bool `yaml:"enabled"`
	Delay   int  `yaml:"delay"`
}

// DefaultConfig builds a Config pre-populated with baked-in defaults.
func DefaultConfig() *Config {
	return &Config{
		Splash: SplashConfig{
			Enabled: DefaultSplashEnabled,
			Delay:   DefaultSplashDelay,
		},
	}
}

type fileConfig struct {
	Splash splashOverrides `yaml:"splash"`
}

type splashOverrides struct {
	Enabled *bool `yaml:"enabled"`
	Delay   *int  `yaml:"delay"`
}

func (c *Config) applyFileOverrides(fc fileConfig) {
	if fc.Splash.Enabled != nil {
		c.Splash.Enabled = *fc.Splash.Enabled
	}
	if fc.Splash.Delay != nil && *fc.Splash.Delay >= 0 {
		c.Splash.Delay = *fc.Splash.Delay
	}
}

// ApplySplashOverrides applies CLI overrides to the splash configuration.
func (c *Config) ApplySplashOverrides(enabled *bool, delay *int) {
	if enabled != nil {
		c.Splash.Enabled = *enabled
	}
	if delay != nil && *delay >= 0 {
		c.Splash.Delay = *delay
	}
}
