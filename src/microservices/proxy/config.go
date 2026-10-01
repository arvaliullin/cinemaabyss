package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Port                   string `envconfig:"PORT" default:"8000"`
	MonolithURL            string `envconfig:"MONOLITH_URL" required:"true"`
	MoviesServiceURL       string `envconfig:"MOVIES_SERVICE_URL" required:"true"`
	GradualMigration       bool   `envconfig:"GRADUAL_MIGRATION" default:"false"`
	MoviesMigrationPercent int    `envconfig:"MOVIES_MIGRATION_PERCENT" default:"100"`

	monolithTarget      *url.URL
	moviesServiceTarget *url.URL
}

// loadConfig загружает и проверяет конфигурацию из переменных окружения.
func loadConfig() (config, error) {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return config{}, fmt.Errorf("not able to read environment variables: %w", err)
	}

	monolithTarget, err := parseURL("MONOLITH_URL", cfg.MonolithURL)
	if err != nil {
		return config{}, err
	}

	moviesServiceTarget, err := parseURL("MOVIES_SERVICE_URL", cfg.MoviesServiceURL)
	if err != nil {
		return config{}, err
	}

	if cfg.MoviesMigrationPercent < 0 || cfg.MoviesMigrationPercent > 100 {
		return config{}, fmt.Errorf("MOVIES_MIGRATION_PERCENT must be an integer from 0 to 100")
	}

	cfg.monolithTarget = monolithTarget
	cfg.moviesServiceTarget = moviesServiceTarget
	return cfg, nil
}

// parseURL проверяет URL upstream-сервиса.
func parseURL(name, value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%s must be a valid URL", name)
	}
	return parsed, nil
}

// isMoviesPath проверяет, относится ли путь к Movies API.
func isMoviesPath(path string) bool {
	return path == "/api/movies" || strings.HasPrefix(path, "/api/movies/")
}
