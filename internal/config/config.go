package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Source   string   `yaml:"source"`
	Target   string   `yaml:"target"`
	Excludes []string `yaml:"excludes"`
	Rules    []Rule   `yaml:"rules"`
}

type Rule struct {
	Kind    string `yaml:"kind"`
	From    string `yaml:"from"`
	To      string `yaml:"to"`
	Pattern string `yaml:"pattern,omitempty"`
	Replace string `yaml:"replace,omitempty"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = findConfigFile()
		if path == "" {
			return nil, fmt.Errorf("no config file found (looked for proto-migrate.yaml)")
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return parse(file)
}

func parse(r io.Reader) (*Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Source == "" {
		return fmt.Errorf("source directory is required")
	}
	if c.Target == "" {
		return fmt.Errorf("target directory is required")
	}

	for i, rule := range c.Rules {
		if err := rule.validate(); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}

	return nil
}

func (r *Rule) validate() error {
	switch r.Kind {
	case "package", "service":
		if r.From == "" || r.To == "" {
			return fmt.Errorf("%s rule requires 'from' and 'to' fields", r.Kind)
		}
	case "regexp":
		if r.Pattern == "" || r.Replace == "" {
			return fmt.Errorf("regexp rule requires 'pattern' and 'replace' fields")
		}
	case "import", "option":
		if r.From == "" || r.To == "" {
			return fmt.Errorf("%s rule requires 'from' and 'to' fields", r.Kind)
		}
	default:
		return fmt.Errorf("unknown rule kind: %s", r.Kind)
	}
	return nil
}

func findConfigFile() string {
	candidates := []string{
		".proto-migrate.yaml",
		".proto-migrate.yml",
		"proto-migrate.yaml",
		"proto-migrate.yml",
	}

	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	dir, _ := os.Getwd()
	for dir != "/" && dir != "" {
		for _, name := range candidates {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		dir = filepath.Dir(dir)
	}

	return ""
}
