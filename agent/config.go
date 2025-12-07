package agent

import (
	"fmt"
	"os"
	
	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	// Server settings
	Port        int    `yaml:"port"`
	TLSCertFile string `yaml:"tls_cert_file"`
	TLSKeyFile  string `yaml:"tls_key_file"`
	
	// Authentication
	HMACSecret         string   `yaml:"hmac_secret"`
	AllowedControllers []string `yaml:"allowed_controllers"`
	
	// Execution settings
	MaxConcurrentJobs int      `yaml:"max_concurrent_jobs"`
	AllowedImages     []string `yaml:"allowed_images"`
	DefaultImage      string   `yaml:"default_image"`
	
	// Security
	RunAsUser  string `yaml:"run_as_user"`
	RunAsGroup string `yaml:"run_as_group"`
	
	// Sandbox defaults
	DefaultCPULimit    float64 `yaml:"default_cpu_limit"`
	DefaultMemoryLimit int64   `yaml:"default_memory_limit"`
	DefaultTimeout     int     `yaml:"default_timeout_seconds"`
	
	// Audit
	AuditDBPath string `yaml:"audit_db_path"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:               8443,
		MaxConcurrentJobs:  5,
		AllowedImages:      []string{"alpine:latest", "ubuntu:latest"},
		DefaultImage:       "alpine:latest",
		RunAsUser:          "quickcmd",
		RunAsGroup:         "quickcmd",
		DefaultCPULimit:    0.5,
		DefaultMemoryLimit: 256 * 1024 * 1024,
		DefaultTimeout:     300,
		AuditDBPath:        "/var/lib/quickcmd/agent-audit.db",
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	
	if c.HMACSecret == "" {
		return fmt.Errorf("hmac_secret is required")
	}
	
	if len(c.AllowedControllers) == 0 {
		return fmt.Errorf("at least one allowed controller is required")
	}
	
	if c.MaxConcurrentJobs < 1 {
		return fmt.Errorf("max_concurrent_jobs must be at least 1")
	}
	
	return nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}
