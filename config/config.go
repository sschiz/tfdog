package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

// Config is a project config.
type Config struct {
	PollerTimeout     time.Duration `env:"POLLER_TIMEOUT" envDefault:"10s"`
	Token             string        `env:"BOT_TOKEN,required,unset,notEmpty"`
	HelpText          string        `env:"HELP_TEXT,required"`
	StartText         string        `env:"START_TEXT,required"`
	SchedulerInterval time.Duration `env:"SCHEDULER_INTERVAL" envDefault:"3m"`
	Logger
}

// ParseConfig parses a config.
func ParseConfig() (*Config, error) {
	cfg := new(Config)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Logger is logger config.
type Logger struct {
	IsProduction bool `env:"LOGGER_PRODUCTION,required"`
}
