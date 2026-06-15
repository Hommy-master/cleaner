package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"cleaner/internal/util"
)

const defaultIntervalSeconds = 60

// Config holds cleaner runtime configuration.
type Config struct {
	Interval int         `json:"interval"`
	Dirs     []DirConfig `json:"dirs"`
	Files    []string    `json:"files"`
}

// DirConfig describes a directory to clean and its ignore list.
type DirConfig struct {
	Path          string   `json:"path"`
	Ignore        []string `json:"ignore"`
	MinAgeSeconds int      `json:"minAgeSeconds"`
}

// Load reads and validates configuration from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate applies defaults and checks configuration constraints.
func (c *Config) Validate() error {
	if c.Interval <= 0 {
		c.Interval = defaultIntervalSeconds
	}

	c.normalizeFiles()

	for i, dir := range c.Dirs {
		if err := util.ValidateDirPath(dir.Path); err != nil {
			return fmt.Errorf("dirs[%d].path: %w", i, err)
		}
		if dir.MinAgeSeconds < 0 {
			return fmt.Errorf("dirs[%d].minAgeSeconds: must be >= 0", i)
		}
	}

	for i, file := range c.Files {
		if !util.IsAbsolutePath(file) {
			return fmt.Errorf("files[%d]: %q must be an absolute path", i, file)
		}
	}

	return nil
}

func (c *Config) normalizeFiles() {
	files := make([]string, 0, len(c.Files))
	for _, file := range c.Files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		files = append(files, file)
	}
	c.Files = files
}

// IntervalDuration returns the configured interval as a duration.
func (c *Config) IntervalDuration() time.Duration {
	return time.Duration(c.Interval) * time.Second
}
