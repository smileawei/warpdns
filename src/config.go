package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Listen   string         `toml:"listen"`
	LogLevel string         `toml:"log_level"`
	Upstream UpstreamConfig `toml:"upstream"`
	ECS      ECSConfig      `toml:"ecs"`
}

type UpstreamConfig struct {
	URL                string            `toml:"url"`
	Path               string            `toml:"path"`
	Method             string            `toml:"method"`
	Timeout            Duration          `toml:"timeout"`
	Headers            map[string]string `toml:"headers"`
	InsecureSkipVerify bool              `toml:"insecure_skip_verify"`
	ServerName         string            `toml:"server_name"`
}

type ECSConfig struct {
	Enabled bool   `toml:"enabled"`
	Subnet  string `toml:"subnet"`
}

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	td, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(td)
	return nil
}

func (d Duration) Std() time.Duration { return time.Duration(d) }

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Listen:   "0.0.0.0:1053",
		LogLevel: "info",
		Upstream: UpstreamConfig{
			Method:  "POST",
			Path:    "/dns-query",
			Timeout: Duration(10 * time.Second),
		},
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.Upstream.URL == "" {
		return nil, fmt.Errorf("upstream.url is required")
	}
	method := strings.ToUpper(strings.TrimSpace(cfg.Upstream.Method))
	if method != "GET" && method != "POST" {
		return nil, fmt.Errorf("upstream.method must be GET or POST (got %q)", cfg.Upstream.Method)
	}
	cfg.Upstream.Method = method
	if cfg.Upstream.Timeout <= 0 {
		cfg.Upstream.Timeout = Duration(10 * time.Second)
	}

	if cfg.ECS.Enabled {
		if _, _, err := net.ParseCIDR(cfg.ECS.Subnet); err != nil {
			return nil, fmt.Errorf("invalid ecs.subnet %q: %w", cfg.ECS.Subnet, err)
		}
	}
	return cfg, nil
}
