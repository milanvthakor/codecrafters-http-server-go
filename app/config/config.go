package config

import (
	"os"
)

// Config holds the server configuration
type Config struct {
	Directory string
	Port      string
}

// NewConfig creates a new configuration from command-line flags
func NewConfig(directory, port string) *Config {
	return &Config{
		Directory: directory,
		Port:      port,
	}
}

// ValidateDirectory checks if the configured directory exists and is valid
func (c *Config) ValidateDirectory() bool {
	if c.Directory == "" {
		return false
	}

	info, err := os.Stat(c.Directory)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// GetDirectory returns the directory path if valid, empty string otherwise
func (c *Config) GetDirectory() string {
	if !c.ValidateDirectory() {
		return ""
	}
	return c.Directory
}
