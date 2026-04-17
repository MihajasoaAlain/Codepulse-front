package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	BaseURL             string
	APIKey              string
	GitHubToken         string // New: GitHub OAuth token
	PollInterval        time.Duration
	DefaultReminderHour int
}

const configFileName = "codepulse_config.json"

// GetConfigDir returns the directory where config is stored
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".codepulse"), nil
}

// GetConfigFilePath returns the full path to the config file
func GetConfigFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// LoadFromFile loads config from disk
func LoadFromFile() (*Config, error) {
	filePath, err := GetConfigFilePath()
	if err != nil {
		return Default(), nil // Return default if can't find home dir
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil // File doesn't exist, use default
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save persists the config to disk
func (c *Config) Save() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filePath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600)
}

// IsAuthenticated returns true if GitHubToken is set
func (c *Config) IsAuthenticated() bool {
	return c.GitHubToken != "" && c.GitHubToken != "YOUR_API_KEY_HERE"
}

func Default() *Config {
	return &Config{
		BaseURL:             "https://api.example.com",
		APIKey:              "YOUR_API_KEY_HERE",
		GitHubToken:         "", // Empty until authenticated
		PollInterval:        5 * time.Minute,
		DefaultReminderHour: 21, // 9 PM
	}
}
