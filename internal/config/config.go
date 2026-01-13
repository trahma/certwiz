package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// OutputConfig controls output formatting
type OutputConfig struct {
	Plain   bool `yaml:"plain"`   // Master switch - disables borders, colors, emojis
	Borders bool `yaml:"borders"` // Show bordered panels
	Colors  bool `yaml:"colors"`  // Use colored output
	Emojis  bool `yaml:"emojis"`  // Show emojis (checkmarks, etc.)
}

// Config holds all certwiz configuration
type Config struct {
	Output OutputConfig `yaml:"output"`
}

var (
	globalConfig *Config
	configOnce   sync.Once
)

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Output: OutputConfig{
			Plain:   false,
			Borders: true,
			Colors:  true,
			Emojis:  true,
		},
	}
}

// configPaths returns the list of config file paths to check, in order of priority
func configPaths() []string {
	var paths []string

	home, err := os.UserHomeDir()
	if err != nil {
		return paths
	}

	// XDG standard location (highest priority)
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(home, ".config")
	}
	paths = append(paths, filepath.Join(xdgConfig, "certwiz", "config.yaml"))

	// Simple dotfile fallback
	paths = append(paths, filepath.Join(home, ".certwiz.yaml"))

	return paths
}

// Load reads the configuration from disk, falling back to defaults
func Load() *Config {
	configOnce.Do(func() {
		globalConfig = loadFromDisk()
	})
	return globalConfig
}

// loadFromDisk attempts to load config from known locations
func loadFromDisk() *Config {
	cfg := DefaultConfig()

	for _, path := range configPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // File doesn't exist or can't be read
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			// Invalid YAML, continue to next file
			continue
		}

		// Successfully loaded config
		return cfg
	}

	return cfg
}

// Get returns the global configuration, loading it if necessary
func Get() *Config {
	return Load()
}

// Reset clears the cached config (useful for testing)
func Reset() {
	configOnce = sync.Once{}
	globalConfig = nil
}

// ApplyPlainMode sets all output options for plain mode
func (c *Config) ApplyPlainMode() {
	c.Output.Plain = true
	c.Output.Borders = false
	c.Output.Colors = false
	c.Output.Emojis = false
}

// ShouldShowBorders returns true if borders should be displayed
func (c *Config) ShouldShowBorders() bool {
	if c.Output.Plain {
		return false
	}
	return c.Output.Borders
}

// ShouldShowColors returns true if colors should be used
func (c *Config) ShouldShowColors() bool {
	if c.Output.Plain {
		return false
	}
	return c.Output.Colors
}

// ShouldShowEmojis returns true if emojis should be displayed
func (c *Config) ShouldShowEmojis() bool {
	if c.Output.Plain {
		return false
	}
	return c.Output.Emojis
}
