package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Workdir string            `yaml:"workdir"`
	Tempdir string            `yaml:"tempdir"`
	Output  string            `yaml:"output"`
	Timeout string            `yaml:"timeout"`
	Env     map[string]string `yaml:"env"`
	Steps   []Step            `yaml:"steps"`

	WorkdirAbs      string        `yaml:"-"`
	TimeoutDuration time.Duration `yaml:"-"`
}

type Step struct {
	Input    []string   `yaml:"input"`
	Run      []Run      `yaml:"run"`
	Expected []Expected `yaml:"expected"`
}

type Run struct {
	Type string   `yaml:"type"`
	Args []string `yaml:"args"`
}

type Expected struct {
	Type  string `yaml:"type"`
	Key   string `yaml:"key"`
	Value any    `yaml:"value"`
}

func loadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	var c Config
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	for k, v := range c.Env {
		c.Env[k] = os.ExpandEnv(v)
	}

	if c.WorkdirAbs, err = filepath.Abs(c.Workdir); err != nil {
		return nil, err
	}

	if c.Timeout != "" {
		if c.TimeoutDuration, err = time.ParseDuration(c.Timeout); err != nil {
			return nil, fmt.Errorf(errInvalidTimeout, err)
		}
	}
	return &c, nil
}

func (c *Config) Path(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(c.WorkdirAbs, p)
}

func (c *Config) TempdirAbs() string {
	if c.Tempdir == "" {
		return c.WorkdirAbs
	}
	return c.Path(c.Tempdir)
}
