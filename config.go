package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type BucketConfig struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
}

type Config struct {
	Default string                  `yaml:"default"`
	Buckets map[string]BucketConfig `yaml:"buckets"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cos-sync", "config.yaml"), nil
}

func LoadConfig() (*Config, string, error) {
	p, err := configPath()
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, p, fmt.Errorf("read config %s: %w\n  tip: copy config.example.yaml to %s and fill in credentials", p, err, p)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, p, fmt.Errorf("parse config %s: %w", p, err)
	}
	if len(cfg.Buckets) == 0 {
		return nil, p, fmt.Errorf("no buckets defined in %s", p)
	}
	if info, err := os.Stat(p); err == nil {
		if info.Mode().Perm()&0o077 != 0 {
			fmt.Fprintf(os.Stderr, "warning: %s is mode %o (group/world readable); consider `chmod 600 %s`\n", p, info.Mode().Perm(), p)
		}
	}
	return &cfg, p, nil
}

// Resolve picks the bucket by alias, falling back to Default when alias is empty.
// Returns the resolved alias so callers can show it back to the user.
func (c *Config) Resolve(alias string) (string, *BucketConfig, error) {
	if alias == "" {
		alias = c.Default
		if alias == "" {
			return "", nil, fmt.Errorf("no --bucket given and no 'default' set in config")
		}
	}
	bc, ok := c.Buckets[alias]
	if !ok {
		names := make([]string, 0, len(c.Buckets))
		for k := range c.Buckets {
			names = append(names, k)
		}
		return "", nil, fmt.Errorf("bucket alias %q not found; available: %v", alias, names)
	}
	if bc.SecretID == "" || bc.SecretKey == "" || bc.Region == "" || bc.Bucket == "" {
		return "", nil, fmt.Errorf("bucket alias %q is incomplete (need secret_id, secret_key, region, bucket)", alias)
	}
	return alias, &bc, nil
}
