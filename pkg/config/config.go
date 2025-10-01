package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"github.com/adriannajera/project-manager-cli/internal/domain"
)

const (
	configFileName = "config.yaml"
	configDirName  = ".pm"
)

// Load loads the configuration from the default locations
func Load() (*domain.Config, error) {
	configPath := getConfigPath()

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := getDefaultConfig()
		if err := Save(defaultConfig); err != nil {
			return nil, err
		}
		return defaultConfig, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config domain.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Ensure database path is absolute
	if !filepath.IsAbs(config.DatabasePath) {
		homeDir, _ := os.UserHomeDir()
		config.DatabasePath = filepath.Join(homeDir, configDirName, config.DatabasePath)
	}

	return &config, nil
}

// Save saves the configuration to the default location
func Save(config *domain.Config) error {
	configPath := getConfigPath()

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return filepath.Join(".", configDirName, configFileName)
	}

	return filepath.Join(homeDir, configDirName, configFileName)
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *domain.Config {
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, configDirName, "tasks.db")

	return &domain.Config{
		DatabasePath:   dbPath,
		DefaultProject: "",
		GitIntegration: true,
		TimeFormat:     "15:04",
		DateFormat:     "2006-01-02",
		Theme: domain.Theme{
			Primary:   "#3b82f6",
			Secondary: "#64748b",
			Success:   "#10b981",
			Warning:   "#f59e0b",
			Error:     "#ef4444",
			Muted:     "#6b7280",
		},
		Aliases: map[string]string{
			"ls":   "list",
			"new":  "add",
			"rm":   "delete",
			"done": "complete",
		},
	}
}