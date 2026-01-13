package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Output.Plain {
		t.Error("Default Plain should be false")
	}
	if !cfg.Output.Borders {
		t.Error("Default Borders should be true")
	}
	if !cfg.Output.Colors {
		t.Error("Default Colors should be true")
	}
	if !cfg.Output.Emojis {
		t.Error("Default Emojis should be true")
	}
}

func TestApplyPlainMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ApplyPlainMode()

	if !cfg.Output.Plain {
		t.Error("Plain should be true after ApplyPlainMode")
	}
	if cfg.Output.Borders {
		t.Error("Borders should be false after ApplyPlainMode")
	}
	if cfg.Output.Colors {
		t.Error("Colors should be false after ApplyPlainMode")
	}
	if cfg.Output.Emojis {
		t.Error("Emojis should be false after ApplyPlainMode")
	}
}

func TestShouldShowBorders(t *testing.T) {
	tests := []struct {
		name     string
		plain    bool
		borders  bool
		expected bool
	}{
		{"default", false, true, true},
		{"borders disabled", false, false, false},
		{"plain mode overrides", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Output: OutputConfig{
					Plain:   tt.plain,
					Borders: tt.borders,
				},
			}
			if got := cfg.ShouldShowBorders(); got != tt.expected {
				t.Errorf("ShouldShowBorders() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldShowColors(t *testing.T) {
	tests := []struct {
		name     string
		plain    bool
		colors   bool
		expected bool
	}{
		{"default", false, true, true},
		{"colors disabled", false, false, false},
		{"plain mode overrides", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Output: OutputConfig{
					Plain:  tt.plain,
					Colors: tt.colors,
				},
			}
			if got := cfg.ShouldShowColors(); got != tt.expected {
				t.Errorf("ShouldShowColors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldShowEmojis(t *testing.T) {
	tests := []struct {
		name     string
		plain    bool
		emojis   bool
		expected bool
	}{
		{"default", false, true, true},
		{"emojis disabled", false, false, false},
		{"plain mode overrides", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Output: OutputConfig{
					Plain:  tt.plain,
					Emojis: tt.emojis,
				},
			}
			if got := cfg.ShouldShowEmojis(); got != tt.expected {
				t.Errorf("ShouldShowEmojis() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".certwiz.yaml")

	configContent := `output:
  plain: false
  borders: false
  colors: true
  emojis: false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", origHome)
		Reset()
	}()

	os.Setenv("HOME", tmpDir)
	Reset() // Clear any cached config

	cfg := Load()

	if cfg.Output.Plain {
		t.Error("Plain should be false from config file")
	}
	if cfg.Output.Borders {
		t.Error("Borders should be false from config file")
	}
	if !cfg.Output.Colors {
		t.Error("Colors should be true from config file")
	}
	if cfg.Output.Emojis {
		t.Error("Emojis should be false from config file")
	}
}

func TestLoadWithNoConfigFile(t *testing.T) {
	// Create an empty temp directory (no config file)
	tmpDir := t.TempDir()

	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", origHome)
		Reset()
	}()

	os.Setenv("HOME", tmpDir)
	Reset()

	cfg := Load()

	// Should get defaults
	if cfg.Output.Plain {
		t.Error("Should get default Plain=false")
	}
	if !cfg.Output.Borders {
		t.Error("Should get default Borders=true")
	}
	if !cfg.Output.Colors {
		t.Error("Should get default Colors=true")
	}
	if !cfg.Output.Emojis {
		t.Error("Should get default Emojis=true")
	}
}

func TestLoadWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".certwiz.yaml")

	// Write invalid YAML
	if err := os.WriteFile(configPath, []byte("this is not valid yaml: ["), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", origHome)
		Reset()
	}()

	os.Setenv("HOME", tmpDir)
	Reset()

	cfg := Load()

	// Should fall back to defaults
	if !cfg.Output.Borders {
		t.Error("Should fall back to default Borders=true on invalid YAML")
	}
}

func TestXDGConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	xdgConfig := filepath.Join(tmpDir, ".config", "certwiz")
	if err := os.MkdirAll(xdgConfig, 0755); err != nil {
		t.Fatalf("Failed to create XDG config dir: %v", err)
	}

	configPath := filepath.Join(xdgConfig, "config.yaml")
	configContent := `output:
  borders: false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", origHome)
		Reset()
	}()

	os.Setenv("HOME", tmpDir)
	Reset()

	cfg := Load()

	if cfg.Output.Borders {
		t.Error("XDG config should set Borders=false")
	}
}

func TestReset(t *testing.T) {
	// Load config once
	_ = Load()

	// Reset should allow reloading
	Reset()

	// Should be able to load again without panic
	cfg := Load()
	if cfg == nil {
		t.Error("Config should not be nil after reset and reload")
	}
}
