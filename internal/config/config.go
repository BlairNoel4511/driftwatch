package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level driftwatch daemon configuration.
type Config struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	LogLevel     string        `yaml:"log_level"`
	Targets      []Target      `yaml:"targets"`
	Alerting     AlertingConfig `yaml:"alerting"`
}

// Target describes a single infrastructure config file to monitor.
type Target struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum,omitempty"`
}

// AlertingConfig holds alerting backend settings.
type AlertingConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Email      string `yaml:"email"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.PollInterval == 0 {
		cfg.PollInterval = 30 * time.Second
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Targets) == 0 {
		return fmt.Errorf("at least one target must be specified")
	}
	for i, t := range c.Targets {
		if t.Name == "" {
			return fmt.Errorf("target[%d]: name is required", i)
		}
		if t.Path == "" {
			return fmt.Errorf("target[%d] %q: path is required", i, t.Name)
		}
	}
	return nil
}
