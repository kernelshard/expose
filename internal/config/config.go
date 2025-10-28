package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigFile = "expose.yaml"

// Config represents the structure of the configuration file.
type Config struct {
	Project     string `yaml:"project"`
	DefaultPort int    `yaml:"default_port"`
}

// Load reads the configuration from the specified or default file path.
func Load(path string) (*Config, error) {

	// Use default config file if no path is provided
	if path == "" {
		path = DefaultConfigFile
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	// Unmarshal YAML data into Config struct to populate cfg fields
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Init creates a default configuration file in the current directory.
func Init() (*Config, error) {
	// Check if default config file exists
	if _, err := os.Stat(DefaultConfigFile); err == nil {
		return nil, fmt.Errorf("config already exists")
	}

	// Get project name from current directory
	dir, _ := os.Getwd()
	projectName := filepath.Base(dir)

	cfg := &Config{
		Project:     projectName,
		DefaultPort: 3000,
	}

	// Write config file
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(DefaultConfigFile, data, 0644); err != nil {
		return nil, err
	}

	return cfg, nil

}
